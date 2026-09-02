package integration_test

import (
	"context"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cmtprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	auctionmodulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/auction/module/v1"
	bondmodulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/bond/module/v1"
	registrymodulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/registry/module/v1"
	"git.vdb.to/cerc-io/laconicd/app/params"
	"git.vdb.to/cerc-io/laconicd/server"
	"git.vdb.to/cerc-io/laconicd/utils"
	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
	auctionkeeper "git.vdb.to/cerc-io/laconicd/x/auction/keeper"
	auctionmodule "git.vdb.to/cerc-io/laconicd/x/auction/module"
	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
	bondkeeper "git.vdb.to/cerc-io/laconicd/x/bond/keeper"
	bondmodule "git.vdb.to/cerc-io/laconicd/x/bond/module"
	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
	registrykeeper "git.vdb.to/cerc-io/laconicd/x/registry/keeper"
	registrymodule "git.vdb.to/cerc-io/laconicd/x/registry/module"
)

type TestFixture struct {
	App *integration.App

	SdkCtx sdk.Context
	cdc    codec.Codec
	keys   map[string]*storetypes.KVStoreKey

	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper

	AuctionKeeper  *auctionkeeper.Keeper
	BondKeeper     *bondkeeper.Keeper
	RegistryKeeper registrykeeper.Keeper
}

func (tf *TestFixture) Setup() error {
	logger := log.NewNopLogger()
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	moduleAccPerms := []*authmodulev1.ModuleAccountPermission{
		{Account: minttypes.ModuleName, Permissions: []string{authtypes.Minter}},
		{Account: auctiontypes.ModuleName},
		{Account: auctiontypes.AuctionBurnModuleAccountName},
		{Account: bondtypes.ModuleName},
		{Account: registrytypes.ModuleName},
		{Account: registrytypes.RecordRentModuleAccountName},
		{Account: registrytypes.AuthorityRentModuleAccountName},
	}

	moduleConfigs := []*appv1alpha1.ModuleConfig{
		{
			Name: runtime.ModuleName,
			Config: appconfig.WrapAny(&runtimev1alpha1.Module{
				AppName: "TestApp",
				BeginBlockers: []string{
					authtypes.ModuleName,
					banktypes.ModuleName,
					auctiontypes.ModuleName,
					bondtypes.ModuleName,
					registrytypes.ModuleName,
				},
				EndBlockers: []string{
					auctiontypes.ModuleName,
					bondtypes.ModuleName,
					registrytypes.ModuleName,
					banktypes.ModuleName,
					authtypes.ModuleName,
				},
				InitGenesis: []string{
					authtypes.ModuleName,
					banktypes.ModuleName,
					auctiontypes.ModuleName,
					bondtypes.ModuleName,
					registrytypes.ModuleName,
				},
			}),
		},
		{
			Name: authtypes.ModuleName,
			Config: appconfig.WrapAny(&authmodulev1.Module{
				Bech32Prefix:                params.Bech32PrefixAccAddr,
				ModuleAccountPermissions:    moduleAccPerms,
				EnableUnorderedTransactions: true,
			}),
		},
		{
			Name:   banktypes.ModuleName,
			Config: appconfig.WrapAny(&bankmodulev1.Module{}),
		},
		{
			Name:   auctiontypes.ModuleName,
			Config: appconfig.WrapAny(&auctionmodulev1.Module{}),
		},
		{
			Name:   bondtypes.ModuleName,
			Config: appconfig.WrapAny(&bondmodulev1.Module{}),
		},
		{
			Name:   registrytypes.ModuleName,
			Config: appconfig.WrapAny(&registrymodulev1.Module{}),
		},
	}

	storeKeys := StoreKeyExtractor{}
	appConfig := depinject.Configs(
		appconfig.Compose(&appv1alpha1.Config{
			Modules: moduleConfigs,
		}),
		depinject.Supply(
			logger,
			authority,
			utils.NewAddressCodec(),
		),
		depinject.Provide(
			server.NewGasService,
		),
		storeKeys.config(),
	)
	var (
		appBuilder     *runtime.AppBuilder
		cdc            codec.Codec
		accountKeeper  authkeeper.AccountKeeper
		bankKeeper     bankkeeper.Keeper
		auctionKeeper  *auctionkeeper.Keeper
		bondKeeper     *bondkeeper.Keeper
		registryKeeper registrykeeper.Keeper
	)
	if err := depinject.Inject(
		appConfig,
		&appBuilder,
		&cdc,
		&accountKeeper,
		&bankKeeper,
		&auctionKeeper,
		&bondKeeper,
		&registryKeeper,
	); err != nil {
		return err
	}

	cms := integration.CreateMultiStore(storeKeys, logger)
	newCtx := sdk.NewContext(cms, cmtprototypes.Header{}, true, logger)

	modules := map[string]appmodule.AppModule{
		authtypes.ModuleName:     auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil),
		banktypes.ModuleName:     bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil),
		auctiontypes.ModuleName:  auctionmodule.NewAppModule(cdc, auctionKeeper),
		bondtypes.ModuleName:     bondmodule.NewAppModule(cdc, bondKeeper),
		registrytypes.ModuleName: registrymodule.NewAppModule(cdc, registryKeeper),
	}

	integrationApp := integration.NewIntegrationApp(newCtx, logger, storeKeys, cdc, modules)
	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	auctiontypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), auctionkeeper.NewMsgServerImpl(auctionKeeper))
	auctiontypes.RegisterQueryServer(integrationApp.QueryHelper(), auctionkeeper.NewQueryServerImpl(auctionKeeper))
	bondtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), bondkeeper.NewMsgServerImpl(bondKeeper))
	bondtypes.RegisterQueryServer(integrationApp.QueryHelper(), bondkeeper.NewQueryServerImpl(bondKeeper))
	registrytypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), registrykeeper.NewMsgServerImpl(registryKeeper))
	registrytypes.RegisterQueryServer(integrationApp.QueryHelper(), registrykeeper.NewQueryServerImpl(registryKeeper))

	auctionKeeper.Params.Set(sdkCtx, auctiontypes.DefaultParams())
	bondKeeper.Params.Set(sdkCtx, bondtypes.DefaultParams())
	registryKeeper.Params.Set(sdkCtx, registrytypes.DefaultParams())

	tf.App, tf.SdkCtx, tf.cdc, tf.keys = integrationApp, sdkCtx, cdc, storeKeys
	tf.AccountKeeper, tf.BankKeeper = accountKeeper, bankKeeper
	tf.AuctionKeeper, tf.BondKeeper, tf.RegistryKeeper = auctionKeeper, bondKeeper, registryKeeper

	return nil
}

type BondDenomProvider struct{}

func (bdp BondDenomProvider) BondDenom(ctx context.Context) (string, error) {
	return params.CoinUnit, nil
}

// store keys are uniquely constructed by ProvideKVStoreKey when building the app, so we must
// extract the keys like this rather than create new ones.
type StoreKeyExtractor map[string]*storetypes.KVStoreKey

func (ske *StoreKeyExtractor) ExtractKey(mkey depinject.ModuleKey, storeKey *storetypes.KVStoreKey) {
	(*ske)[mkey.Name()] = storeKey
}

func (ske *StoreKeyExtractor) config() depinject.Config {
	return depinject.Configs(
		depinject.Supply(ske),
		depinject.InvokeInModule(banktypes.ModuleName, (*StoreKeyExtractor).ExtractKey),
		depinject.InvokeInModule(authtypes.ModuleName, (*StoreKeyExtractor).ExtractKey),
		depinject.InvokeInModule(bondtypes.ModuleName, (*StoreKeyExtractor).ExtractKey),
		depinject.InvokeInModule(auctiontypes.ModuleName, (*StoreKeyExtractor).ExtractKey),
		depinject.InvokeInModule(registrytypes.ModuleName, (*StoreKeyExtractor).ExtractKey),
	)
}
