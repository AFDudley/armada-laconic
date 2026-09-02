package module

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	modulev1 "git.vdb.to/cerc-io/laconicd/api/cerc/onboarding/module/v1"
	"git.vdb.to/cerc-io/laconicd/utils"
	"git.vdb.to/cerc-io/laconicd/x/onboarding/keeper"
)

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

	Cdc          codec.Codec
	StoreService store.KVStoreService
	AddressCodec address.Codec
	EventService event.Service
	GasService   gas.Service
	Logger       log.Logger

	Config *modulev1.Module
}

type ModuleOutputs struct {
	depinject.Out

	Module appmodule.AppModule
	Keeper *keeper.Keeper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance as authority if not provided
	authority := utils.AddressOrModuleAddress(in.Config.Authority, govtypes.ModuleName)
	k := keeper.NewKeeper(in.Cdc, in.AddressCodec, in.StoreService, in.EventService, in.GasService, authority, in.Logger)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k}
}
