package registry

const (
	EventTypeSetRecord          = "set"
	EventTypeDeleteName         = "delete-name"
	EventTypeReserveAuthority   = "reserve-authority"
	EventTypeAuthorityBond      = "authority-bond"
	EventTypeRenewRecord        = "renew-record"
	EventTypeAssociateBond      = "associate-bond"
	EventTypeDissociateBond     = "dissociate-bond"
	EventTypeDissociateRecords  = "dissociate-record"
	EventTypeReassociateRecords = "re-associate-records"

	AttributeKeySigner     = "signer"
	AttributeKeyOwner      = "owner"
	AttributeKeyBondId     = "bond-id"
	AttributeKeyPayload    = "payload"
	AttributeKeyOldBondId  = "old-bond-id"
	AttributeKeyNewBondId  = "new-bond-id"
	AttributeKeyCID        = "cid"
	AttributeKeyName       = "name"
	AttributeKeyLRN        = "lrn"
	AttributeKeyRecordId   = "record-id"
	AttributeValueCategory = ModuleName
)
