package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
)

type queryHandlers struct {
	k *Keeper
}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k *Keeper) bondtypes.QueryServer {
	return queryHandlers{k}
}

// Params implements bond.QueryServer.
func (qs queryHandlers) Params(ctx context.Context, _ *bondtypes.QueryParamsRequest) (*bondtypes.QueryParamsResponse, error) {
	params, err := qs.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &bondtypes.QueryParamsResponse{Params: params}, nil
}

// Bonds implements bond.QueryServer.
func (qs queryHandlers) Bonds(ctx context.Context, _ *bondtypes.QueryBondsRequest) (*bondtypes.QueryBondsResponse, error) {
	resp, err := qs.k.ListBonds(ctx)
	if err != nil {
		return nil, err
	}

	return &bondtypes.QueryBondsResponse{Bonds: resp}, nil
}

// GetBondById implements bond.QueryServer.
func (qs queryHandlers) GetBondById(ctx context.Context, req *bondtypes.QueryGetBondByIdRequest) (*bondtypes.QueryGetBondByIdResponse, error) {
	bondId := req.GetId()
	if len(bondId) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "bond id required")
	}

	bond, err := qs.k.GetBondById(ctx, bondId)
	if err != nil {
		return nil, err
	}

	return &bondtypes.QueryGetBondByIdResponse{Bond: &bond}, nil
}

// GetBondsByOwner implements bond.QueryServer.
func (qs queryHandlers) GetBondsByOwner(
	ctx context.Context,
	req *bondtypes.QueryGetBondsByOwnerRequest,
) (*bondtypes.QueryGetBondsByOwnerResponse, error) {
	owner := req.GetOwner()
	if len(owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
			"owner required",
		)
	}

	bonds, err := qs.k.GetBondsByOwner(ctx, owner)
	if err != nil {
		return nil, err
	}

	return &bondtypes.QueryGetBondsByOwnerResponse{Bonds: bonds}, nil
}

// GetBondModuleBalance implements bond.QueryServer.
func (qs queryHandlers) GetBondModuleBalance(
	ctx context.Context,
	_ *bondtypes.QueryGetBondModuleBalanceRequest,
) (*bondtypes.QueryGetBondModuleBalanceResponse, error) {
	balances := qs.k.GetBondModuleBalances(ctx)

	return &bondtypes.QueryGetBondModuleBalanceResponse{
		Balance: balances,
	}, nil
}
