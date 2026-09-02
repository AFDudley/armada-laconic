package utils

import (
	"context"
	"fmt"

	"cosmossdk.io/core/gas"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RefundModuleGasDescriptor = "laconic module gas refund"

// WithCustomGasConfig returns a context with a zero-cost gas config, and a function to log all consumed gas.
//
// Note: as planned in server/v2, the gas meter will not be available on the STF-provided context
// during tx execution, and must be accessed via gas.Service instead. This arg is a placeholder
// to ensure modules using this include the gas service in anticipation of that change.
// https://github.com/cosmos/cosmos-sdk/pull/16310
func WithCustomGasConfig(ctx context.Context, _ gas.Service) (context.Context, func(log.Logger, string)) {
	c := sdk.UnwrapSDKContext(ctx)
	meter := c.GasMeter()

	startGas := meter.GasConsumed()
	logGas := func(logger log.Logger, method string) {
		endGas := meter.GasConsumed()
		meter.RefundGas(endGas-startGas, RefundModuleGasDescriptor)
		LogTxGasConsumed(meter, logger, method)
	}

	// Set a zero-cost gas config on the returned context
	return c.WithKVGasConfig(storetypes.GasConfig{}), logGas
}

func LogTxGasConsumed(gm gas.Meter, logger log.Logger, method string) {
	gasConsumed := gm.GasConsumed()
	logger.Info("tx executed", "method", method, "gas_consumed", fmt.Sprintf("%d", gasConsumed))
}
