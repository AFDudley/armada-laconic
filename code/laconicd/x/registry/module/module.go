package module

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/client/v2/autocli"
	appmodule "cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	registry "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
	"git.vdb.to/cerc-io/laconicd/x/registry/client/cli"
	"git.vdb.to/cerc-io/laconicd/x/registry/keeper"
)

var (
	_ module.AppModuleBasic      = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasServices         = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}

	_ autocli.HasCustomTxCommand = AppModule{}
)

// ConsensusVersion defines the current module consensus version.
const ConsensusVersion = 1

type AppModule struct {
	cdc    codec.Codec
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		keeper: keeper,
	}
}

// module.AppModule

// Name returns the registry module's name.
func (AppModule) Name() string { return registrytypes.ModuleName }

// module.HasGRPCGateway

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the registry module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := registrytypes.RegisterQueryHandlerClient(context.Background(), mux, registrytypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// appmodule.HasRegisterInterfaces

// RegisterInterfaces registers interfaces and implementations of the registry module.
func (AppModule) RegisterInterfaces(registry registry.InterfaceRegistry) {
	registrytypes.RegisterInterfaces(registry)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// appmodule.HasGenesis

// DefaultGenesis returns default genesis state as raw bytes for the module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(registrytypes.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the registry module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data registrytypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", registrytypes.ModuleName, err)
	}

	return data.Validate()
}

// InitGenesis performs genesis initialization for the registry module.
// It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState registrytypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	if err := am.keeper.InitGenesis(ctx, &genesisState); err != nil {
		panic(fmt.Sprintf("failed to initialize %s genesis state: %v", registrytypes.ModuleName, err))
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the registry
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to export %s genesis state: %w", registrytypes.ModuleName, err))
	}

	return cdc.MustMarshalJSON(gs)
}

// module.HasServices

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register servers
	registrytypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	registrytypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))

	// Register in place module state migration migrations
	// m := keeper.NewMigrator(am.keeper)
	// if err := cfg.RegisterMigration(registrytypes.ModuleName, 1, m.Migrate1to2); err != nil {
	//     panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", registrytypes.ModuleName, err))
	// }
}

// appmodule.HasEndBlocker

func (am AppModule) EndBlock(ctx context.Context) error {
	return EndBlocker(ctx, am.keeper)
}

// autocli.HasCustomTxCommand

// RegisterLegacyAminoCodec registers the registry module's types on the LegacyAmino codec.
// New modules do not need to support Amino.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// Get the root tx command of this module
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}
