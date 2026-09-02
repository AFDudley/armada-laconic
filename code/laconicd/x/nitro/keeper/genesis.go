package keeper

import (
	"context"

	"git.vdb.to/cerc-io/laconicd/x/nitro/types"
)

// TODO

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	return &types.GenesisState{}, nil
}
