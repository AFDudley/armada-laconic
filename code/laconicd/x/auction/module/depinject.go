package module

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	auth "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	modulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/auction/module/v1"
	"git.vdb.to/cerc-io/laconicd/utils"
	"git.vdb.to/cerc-io/laconicd/x/auction"
	"git.vdb.to/cerc-io/laconicd/x/auction/keeper"
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
		appconfig.Invoke(InvokeSetAuctionHooks),
	)
}

type ModuleInputs struct {
	depinject.In

	StoreService  store.KVStoreService
	HeaderService header.Service
	EventService  event.Service
	GasService    gas.Service
	Config        *modulev1.Module
	Cdc           codec.Codec

	AccountKeeper auth.AccountKeeper
	BankKeeper    bank.Keeper

	Logger log.Logger
}

type ModuleOutputs struct {
	depinject.Out

	// Use * as required by InvokeSetAuctionHooks
	// https://github.com/cosmos/cosmos-sdk/tree/v0.50.3/core/appmodule#invoker-invocation-details
	// https://github.com/cosmos/cosmos-sdk/tree/v0.50.3/core/appmodule#regular-golang-types
	Keeper *keeper.Keeper

	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := utils.AddressOrModuleAddress(in.Config.Authority, govtypes.ModuleName)

	k := keeper.NewKeeper(in.Cdc, in.StoreService, in.HeaderService, in.EventService, in.GasService, in.AccountKeeper, in.BankKeeper, authority, in.Logger)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k}
}

func InvokeSetAuctionHooks(
	config *modulev1.Module,
	keeper *keeper.Keeper,
	auctionHooks map[string]auction.AuctionHooksWrapper,
) error {
	// All arguments to invokers are optional
	if keeper == nil || config == nil {
		return nil
	}

	usageKeepers := make([]auction.AuctionUsageKeeper, 0, len(auctionHooks))

	for _, hook := range auctionHooks {
		usageKeepers = append(usageKeepers, hook)
	}

	keeper.SetUsageKeepers(usageKeepers)

	return nil
}
