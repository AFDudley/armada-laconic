package registry

import "cosmossdk.io/collections"

const (
	// ModuleName is the name of the registry module
	ModuleName = "registry"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RecordRentModuleAccountName is the name of the module account that keeps track of record rents paid.
	RecordRentModuleAccountName = "record_rent"

	// AuthorityRentModuleAccountName is the name of the module account that keeps track of authority rents paid.
	AuthorityRentModuleAccountName = "authority_rent"
)

// Store prefixes
var (
	// ParamsKey is the prefix for params key
	ParamsPrefix = collections.NewPrefix(0)

	RecordsPrefix              = collections.NewPrefix(1)
	RecordsByBondIdIndexPrefix = collections.NewPrefix(2)

	AuthoritiesPrefix                 = collections.NewPrefix(3)
	AuthoritiesByAuctionIdIndexPrefix = collections.NewPrefix(4)
	AuthoritiesByBondIdIndexPrefix    = collections.NewPrefix(5)

	NameRecordsPrefix           = collections.NewPrefix(6)
	NameRecordsByCidIndexPrefix = collections.NewPrefix(7)

	RecordExpiryQueuePrefix    = collections.NewPrefix(8)
	AuthorityExpiryQueuePrefix = collections.NewPrefix(9)

	AttributesMapPrefix = collections.NewPrefix(10)
)
