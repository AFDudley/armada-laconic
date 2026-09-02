package server

import (
	"context"

	"cosmossdk.io/core/gas"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ gas.Service = gasService{}

// Placeholder gas service, in anticipation of future use
// https://github.com/cosmos/cosmos-sdk/pull/16310
type gasService struct{}

func NewGasService() gas.Service { return gasService{} }

func (gasService) GetGasMeter(ctx context.Context) gas.Meter {
	c := sdk.UnwrapSDKContext(ctx)
	return c.GasMeter()
}

func (gasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	c := sdk.UnwrapSDKContext(ctx)
	return c.WithGasMeter(meter)
}

// deprecated https://github.com/cosmos/cosmos-sdk/issues/19793
func (gasService) GetBlockGasMeter(context.Context) gas.Meter {
	return nil
}

// deprecated https://github.com/cosmos/cosmos-sdk/issues/19793
func (gasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return ctx
}
