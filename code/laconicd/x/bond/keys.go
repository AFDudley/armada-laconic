package bond

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	ModuleName = "bond"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

// Store prefixes
var (
	ParamsPrefix = collections.NewPrefix(0)

	BondsPrefix          = collections.NewPrefix(1)
	BondOwnerIndexPrefix = collections.NewPrefix(2)
)

// ModuleAccountAddress defines the x/bond module address.
var ModuleAccountAddress = address.Module(ModuleName)
