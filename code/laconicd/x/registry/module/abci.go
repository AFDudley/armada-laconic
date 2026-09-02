package module

import (
	"context"

	"git.vdb.to/cerc-io/laconicd/x/registry/keeper"
)

// EndBlocker is called every block
func EndBlocker(ctx context.Context, k keeper.Keeper) error {
	if err := k.ProcessRecordExpiryQueue(ctx); err != nil {
		return err
	}

	if err := k.ProcessAuthorityExpiryQueue(ctx); err != nil {
		return err
	}

	return nil
}
