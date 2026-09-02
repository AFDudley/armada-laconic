package module

import (
	"context"

	"git.vdb.to/cerc-io/laconicd/x/auction/keeper"
)

// EndBlocker is called every block
func EndBlocker(ctx context.Context, k *keeper.Keeper) error {
	return k.EndBlockerProcessAuctions(ctx)
}
