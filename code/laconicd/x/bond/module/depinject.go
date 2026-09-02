package module

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	auth "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	modulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/bond/module/v1"
	"git.vdb.to/cerc-io/laconicd/utils"
	"git.vdb.to/cerc-io/laconicd/x/bond"
	"git.vdb.to/cerc-io/laconicd/x/bond/keeper"
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
		appconfig.Invoke(InvokeSetBondHooks),
	)
}

type ModuleInputs struct {
	depinject.In

	StoreService store.KVStoreService
	EventService event.Service
	GasService   gas.Service
	Config       *modulev1.Module
	Cdc          codec.Codec

	AccountKeeper auth.AccountKeeper
	BankKeeper    bank.Keeper

	Logger log.Logger
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := utils.AddressOrModuleAddress(in.Config.Authority, govtypes.ModuleName)
	k := keeper.NewKeeper(in.Cdc, in.StoreService, in.EventService, in.GasService, in.AccountKeeper, in.BankKeeper, authority, in.Logger)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k}
}

func InvokeSetBondHooks(
	config *modulev1.Module,
	keeper *keeper.Keeper,
	bondHooks map[string]bond.BondHooksWrapper,
) error {
	// All arguments to invokers are optional
	if keeper == nil || config == nil {
		return nil
	}

	usageKeepers := make([]bond.BondUsageKeeper, 0, len(bondHooks))

	for _, hook := range bondHooks {
		usageKeepers = append(usageKeepers, hook)
	}

	keeper.SetUsageKeepers(usageKeepers)

	return nil
}
