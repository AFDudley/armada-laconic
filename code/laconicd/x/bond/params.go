package bond

import (
	"errors"
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
	"git.vdb.to/cerc-io/laconicd/app/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultMaxBondAmountTokens are the default parameter values.
var DefaultMaxBondAmountTokens = sdkmath.NewInt(1000000000000) // 10^12

func NewParams(maxBondAmount sdk.Coin) Params {
	return Params{MaxBondAmount: maxBondAmount}
}

// DefaultParams returns default module parameters
func DefaultParams() Params {
	return NewParams(sdk.NewCoin(params.CoinUnit, DefaultMaxBondAmountTokens))
}

// Validate checks that the parameters have valid values
func (p Params) Validate() error {
	if err := validateMaxBondAmount(p.MaxBondAmount); err != nil {
		return err
	}

	return nil
}

func validateMaxBondAmount(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Amount.IsNegative() {
		return errors.New("max bond amount must be positive")
	}

	return nil
}
