package bond

import context "context"

// BondUsageKeeper keep track of bond usage in other modules.
// Used to, for example, prevent deletion of a bond that's in use.
type BondUsageKeeper interface {
	ModuleName() string
	UsesBond(ctx context.Context, bondId string) bool
}

// BondHooksWrapper is a wrapper for modules to inject BondUsageKeeper using depinject.
// Reference: https://github.com/cosmos/cosmos-sdk/tree/v0.50.3/core/appmodule#resolving-circular-dependencies
type BondHooksWrapper struct{ BondUsageKeeper }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (BondHooksWrapper) IsOnePerModuleType() {}
