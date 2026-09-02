package registry

import (
	fmt "fmt"
	time "time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/app/params"
)

// Default parameter values.
var (
	// DefaultRecordRent is the default record rent for 1 time period (see expiry time).
	DefaultRecordRent = sdkmath.NewInt(1000000) // 10^6

	// DefaultRecordExpiryTime is the default record expiry time (1 year).
	DefaultRecordExpiryTime = time.Hour * 24 * 365

	DefaultAuthorityRent        = sdkmath.NewInt(1000000) // 10^6
	DefaultAuthorityExpiryTime  = time.Hour * 24 * 365
	DefaultAuthorityGracePeriod = time.Hour * 24 * 2

	DefaultAuthorityAuctionEnabled = false
	DefaultCommitsDuration         = time.Hour * 24
	DefaultRevealsDuration         = time.Hour * 24
	DefaultCommitFee               = sdkmath.NewInt(1000000) // 10^6
	DefaultRevealFee               = sdkmath.NewInt(1000000) // 10^6
	DefaultMinimumBid              = sdkmath.NewInt(5000000) // 5 * 10^6
)

// NewParams creates a new Params instance
func NewParams(
	recordRent sdk.Coin,
	recordRentDuration time.Duration,
	authorityRent sdk.Coin,
	authorityRentDuration time.Duration,
	authorityGracePeriod time.Duration,
	authorityAuctionEnabled bool,
	commitsDuration time.Duration, revealsDuration time.Duration,
	commitFee sdk.Coin, revealFee sdk.Coin,
	minimumBid sdk.Coin,
) Params {
	return Params{
		RecordRent:         recordRent,
		RecordRentDuration: recordRentDuration,

		AuthorityRent:         authorityRent,
		AuthorityRentDuration: authorityRentDuration,
		AuthorityGracePeriod:  authorityGracePeriod,

		AuthorityAuctionEnabled:         authorityAuctionEnabled,
		AuthorityAuctionCommitsDuration: commitsDuration,
		AuthorityAuctionRevealsDuration: revealsDuration,
		AuthorityAuctionCommitFee:       commitFee,
		AuthorityAuctionRevealFee:       revealFee,
		AuthorityAuctionMinimumBid:      minimumBid,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		sdk.NewCoin(params.CoinUnit, DefaultRecordRent), DefaultRecordExpiryTime,
		sdk.NewCoin(params.CoinUnit, DefaultAuthorityRent),
		DefaultAuthorityExpiryTime, DefaultAuthorityGracePeriod, DefaultAuthorityAuctionEnabled, DefaultCommitsDuration,
		DefaultRevealsDuration,
		sdk.NewCoin(params.CoinUnit, DefaultCommitFee),
		sdk.NewCoin(params.CoinUnit, DefaultRevealFee),
		sdk.NewCoin(params.CoinUnit, DefaultMinimumBid),
	)
}

// Validate a set of params.
func (p Params) Validate() error {
	if err := validateRecordRent(p.RecordRent); err != nil {
		return err
	}

	if err := validateRecordRentDuration(p.RecordRentDuration); err != nil {
		return err
	}

	if err := validateAuthorityRent(p.AuthorityRent); err != nil {
		return err
	}

	if err := validateAuthorityRentDuration(p.AuthorityRentDuration); err != nil {
		return err
	}

	if err := validateAuthorityGracePeriod(p.AuthorityGracePeriod); err != nil {
		return err
	}

	if err := validateAuthorityAuctionEnabled(p.AuthorityAuctionEnabled); err != nil {
		return err
	}

	if err := validateCommitsDuration(p.AuthorityAuctionCommitsDuration); err != nil {
		return err
	}

	if err := validateRevealsDuration(p.AuthorityAuctionRevealsDuration); err != nil {
		return err
	}

	if err := validateCommitFee(p.AuthorityAuctionCommitFee); err != nil {
		return err
	}

	if err := validateRevealFee(p.AuthorityAuctionRevealFee); err != nil {
		return err
	}

	if err := validateMinimumBid(p.AuthorityAuctionMinimumBid); err != nil {
		return err
	}

	return nil
}

func validateRecordRent(i interface{}) error {
	return validateAmount("RecordRent", i)
}

func validateRecordRentDuration(i interface{}) error {
	return validateDuration("RecordRentDuration", i)
}

func validateAuthorityRent(i interface{}) error {
	return validateAmount("AuthorityRent", i)
}

func validateAuthorityRentDuration(i interface{}) error {
	return validateDuration("AuthorityRentDuration", i)
}

func validateAuthorityGracePeriod(i interface{}) error {
	return validateDuration("AuthorityGracePeriod", i)
}

func validateAuthorityAuctionEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("%s invalid parameter type: %T", "AuthorityAuctionEnabled", i)
	}

	return nil
}

func validateCommitsDuration(i interface{}) error {
	return validateDuration("AuthorityCommitsDuration", i)
}

func validateRevealsDuration(i interface{}) error {
	return validateDuration("AuthorityRevealsDuration", i)
}

func validateCommitFee(i interface{}) error {
	return validateAmount("AuthorityCommitFee", i)
}

func validateRevealFee(i interface{}) error {
	return validateAmount("AuthorityRevealFee", i)
}

func validateMinimumBid(i interface{}) error {
	return validateAmount("AuthorityMinimumBid", i)
}

func validateAmount(name string, i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("%s invalid parameter type: %T", name, i)
	}

	if v.Amount.IsNegative() {
		return fmt.Errorf("%s can't be negative", name)
	}

	return nil
}

func validateDuration(name string, i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("%s invalid parameter type: %T", name, i)
	}

	if v <= 0 {
		return fmt.Errorf("%s must be a positive integer", name)
	}

	return nil
}
