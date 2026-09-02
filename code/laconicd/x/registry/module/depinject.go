package module

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	auth "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	modulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/registry/module/v1"
	"git.vdb.to/cerc-io/laconicd/utils"
	"git.vdb.to/cerc-io/laconicd/x/auction"
	auctionkeeper "git.vdb.to/cerc-io/laconicd/x/auction/keeper"
	"git.vdb.to/cerc-io/laconicd/x/bond"
	bondkeeper "git.vdb.to/cerc-io/laconicd/x/bond/keeper"
	"git.vdb.to/cerc-io/laconicd/x/registry/keeper"
)

var _ appmodule.AppModule = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

func init() {
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	StoreService  store.KVStoreService
	HeaderService header.Service
	EventService  event.Service
	Config        *modulev1.Module
	Cdc           codec.Codec
	Logger        log.Logger

	AccountKeeper auth.AccountKeeper
	BankKeeper    bank.Keeper

	BondKeeper    *bondkeeper.Keeper
	AuctionKeeper *auctionkeeper.Keeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule

	AuctionHooks auction.AuctionHooksWrapper
	BondHooks    bond.BondHooksWrapper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := utils.AddressOrModuleAddress(in.Config.Authority, govtypes.ModuleName)
	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService,

		in.HeaderService,
		in.EventService,
		in.AccountKeeper,
		in.BankKeeper,
		in.BondKeeper,
		in.AuctionKeeper,
		authority,
		in.Logger,
	)
	m := NewAppModule(in.Cdc, k)

	recordKeeper := keeper.NewRecordKeeper(in.Cdc, &k, in.AuctionKeeper, in.Logger)

	return ModuleOutputs{
		Module: m, Keeper: k,
		AuctionHooks: auction.AuctionHooksWrapper{AuctionUsageKeeper: recordKeeper},
		BondHooks:    bond.BondHooksWrapper{BondUsageKeeper: recordKeeper},
	}
}
