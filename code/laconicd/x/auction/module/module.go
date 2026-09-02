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

	"git.vdb.to/cerc-io/laconicd/x/auction"
	"git.vdb.to/cerc-io/laconicd/x/auction/client/cli"
	"git.vdb.to/cerc-io/laconicd/x/auction/keeper"
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

// ConsensusVersion defines the current module consensus version
const ConsensusVersion = 1

type AppModule struct {
	cdc    codec.Codec
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		keeper: keeper,
	}
}

// module.AppModule

// Name returns the auction module's name.
func (AppModule) Name() string { return auction.ModuleName }

// module.HasGRPCGateway

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the auction module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := auction.RegisterQueryHandlerClient(context.Background(), mux, auction.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// appmodule.HasRegisterInterfaces

// RegisterInterfaces registers interfaces and implementations of the auction module.
func (AppModule) RegisterInterfaces(registry registry.InterfaceRegistry) {
	auction.RegisterInterfaces(registry)
}

// appmodule.HasConsensusVersion

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// module.HasGenesis

// DefaultGenesis returns default genesis state as raw bytes for the module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(auction.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the auction module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data auction.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", auction.ModuleName, err)
	}

	return data.Validate()
}

// InitGenesis performs genesis initialization for the auction module.
// It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState auction.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	if err := am.keeper.InitGenesis(ctx, &genesisState); err != nil {
		panic(fmt.Sprintf("failed to initialize %s genesis state: %v", auction.ModuleName, err))
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the auction
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to export %s genesis state: %w", auction.ModuleName, err))
	}

	return cdc.MustMarshalJSON(gs)
}

// module.HasServices

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register servers
	auction.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	auction.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))

	// Register in place module state migration migrations
	// m := keeper.NewMigrator(am.keeper)
	// if err := cfg.RegisterMigration(auction.ModuleName, 1, m.Migrate1to2); err != nil {
	//     panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", auction.ModuleName, err))
	// }
}

// appmodule.HasEndBlocker

func (am AppModule) EndBlock(ctx context.Context) error {
	return EndBlocker(ctx, am.keeper)
}

// autocli.HasCustomTxCommand

// RegisterLegacyAminoCodec registers the auction module's types on the LegacyAmino codec.
// New modules do not need to support Amino.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// Get the root tx command of this module
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}
