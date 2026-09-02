package auction

import "context"

// AuctionUsageKeeper keep track of auction usage in other modules.
// Used to, for example, prevent deletion of a auction that's in use.
type AuctionUsageKeeper interface {
	ModuleName() string
	UsesAuction(ctx context.Context, auctionId string) bool

	OnAuctionWinnerSelected(ctx context.Context, auctionId string, currentHeight uint64)
}

// AuctionHooksWrapper is a wrapper for modules to inject AuctionUsageKeeper using depinject.
// Reference: https://github.com/cosmos/cosmos-sdk/tree/v0.50.3/core/appmodule#resolving-circular-dependencies
type AuctionHooksWrapper struct{ AuctionUsageKeeper }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AuctionHooksWrapper) IsOnePerModuleType() {}
