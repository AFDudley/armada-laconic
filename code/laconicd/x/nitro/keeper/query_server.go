package keeper

import types "git.vdb.to/cerc-io/laconicd/x/nitro/types/v1"

type queryHandlers struct {
	k Keeper
}

func NewQueryServer(k Keeper) types.QueryServer {
	return queryHandlers{k}
}
