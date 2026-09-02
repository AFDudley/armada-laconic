package keeper

import (
	"context"

	"git.vdb.to/cerc-io/laconicd/x/auction"
)

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *auction.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	// Save auctions in store.
	for _, auction := range data.Auctions.Auctions {
		if err := k.Auctions.Set(ctx, auction.Id, auction); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*auction.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	auctions, err := k.ListAuctions(ctx)
	if err != nil {
		return nil, err
	}

	return &auction.GenesisState{
		Params:   params,
		Auctions: &auction.Auctions{Auctions: auctions},
	}, nil
}
