package module

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	auth "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	modulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/nitro/module/v1"
	"git.vdb.to/cerc-io/laconicd/x/nitro/keeper"
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

	StoreService store.KVStoreService
	EventService event.Service
	Config       *modulev1.Module
	Cdc          codec.Codec
	Logger       log.Logger

	AccountKeeper auth.AccountKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// // default to governance authority if not provided
	// authority := utils.AddressOrModuleAddress(in.Config.Authority, govtypes.ModuleName)
	k := keeper.NewKeeper(in.Cdc, in.StoreService, in.EventService, in.AccountKeeper, in.Logger)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k}
}
