package v1

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ authtypes.GenesisAccount = (*LockupAccount)(nil)
	_ LockupAccountI           = (*LockupAccount)(nil)
)

// LockupAccountI defines an account interface for lockup account with token distribution
type LockupAccountI interface {
	sdk.ModuleAccountI

	GetDistribution() string
}

// Validate checks for errors on the account fields
func (la LockupAccount) Validate() error {
	if la.BaseAccount == nil {
		return errors.New("uninitialized LockupAccount: BaseAccount is nil")
	}

	return la.BaseAccount.Validate()
}

// HasPermission returns whether or not the account has permission.
// Return false as the lockup account doesn't have any permissions
func (la LockupAccount) HasPermission(permission string) bool {
	return false
}

// GetName returns the name of the holder's module
func (la LockupAccount) GetName() string {
	return la.Name
}

// GetPermissions returns permissions granted to the module account
// Return empty as the lockup account doesn't have any permissions
func (la LockupAccount) GetPermissions() []string {
	return []string{}
}

// GetDistribution returns the total token distribution
func (la LockupAccount) GetDistribution() string {
	return la.Distribution
}
