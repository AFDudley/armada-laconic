package keeper

import (
	"context"

	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
)

// GetParams - Get all parameters as types.Params.
func (k Keeper) GetParams(ctx context.Context) (*registrytypes.Params, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &params, nil
}
