package module

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"git.vdb.to/cerc-io/laconicd/x/bond"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	registry "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"git.vdb.to/cerc-io/laconicd/x/nitro"
	"git.vdb.to/cerc-io/laconicd/x/nitro/client/cli"
	"git.vdb.to/cerc-io/laconicd/x/nitro/keeper"
	"git.vdb.to/cerc-io/laconicd/x/nitro/types"
	v1 "git.vdb.to/cerc-io/laconicd/x/nitro/types/v1"
)

var (
	_ module.AppModuleBasic      = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasServices         = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// ConsensusVersion defines the current module consensus version
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

// Name returns the bond module's name.
func (AppModule) Name() string { return types.ModuleName }

// module.HasGRPCGateway

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bond module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	// TODO

	// if err := RegisterQueryHandlerClient(context.Background(), mux, NewQueryClient(clientCtx)); err != nil {
	// 	panic(err)
	// }
}

// appmodule.HasRegisterInterfaces

// RegisterInterfaces registers interfaces and implementations of the bond module.
func (AppModule) RegisterInterfaces(registry registry.InterfaceRegistry) {
	nitro.RegisterInterfaces(registry)
}

// appmodule.HasConsensusVersion

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// appmodule.HasGenesis

// DefaultGenesis returns default genesis state as raw bytes for the module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&types.GenesisState{})
}

// ValidateGenesis performs genesis state validation for the bond module.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", bond.ModuleName, err)
	}

	return data.Validate()
}

// InitGenesis performs genesis initialization for the bond module.
// It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	if err := am.keeper.InitGenesis(ctx, &genesisState); err != nil {
		panic(fmt.Sprintf("failed to initialize %s genesis state: %v", bond.ModuleName, err))
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the bond
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to export %s genesis state: %w", bond.ModuleName, err))
	}

	return cdc.MustMarshalJSON(gs)
}

// module.HasServices

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register servers
	v1.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	v1.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))

	// Register in place module state migration migrations
	// m := keeper.NewMigrator(am.keeper)
	// if err := cfg.RegisterMigration(bond.ModuleName, 1, m.Migrate1to2); err != nil {
	//     panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", bond.ModuleName, err))
	// }
}

// RegisterLegacyAminoCodec registers the bond module's types on the LegacyAmino codec.
// New modules do not need to support Amino.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// GetTxCmd returns the root tx command for the nitro module.
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns the root tx command for the nitro module.
func (AppModule) GetQueryCmd() *cobra.Command {
	return cli.NewQueryCmd()
}
