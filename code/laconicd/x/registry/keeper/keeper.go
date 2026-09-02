package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	store "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	auth "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/gibson042/canonicaljson-go"
	cid "github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"

	auctionkeeper "git.vdb.to/cerc-io/laconicd/x/auction/keeper"
	bondkeeper "git.vdb.to/cerc-io/laconicd/x/bond/keeper"
	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
	"git.vdb.to/cerc-io/laconicd/x/registry/helpers"
)

type RecordsIndexes struct {
	BondId *indexes.Multi[string, string, registrytypes.Record]
}

func (b RecordsIndexes) IndexesList() []collections.Index[string, registrytypes.Record] {
	return []collections.Index[string, registrytypes.Record]{b.BondId}
}

func newRecordIndexes(sb *collections.SchemaBuilder) RecordsIndexes {
	return RecordsIndexes{
		BondId: indexes.NewMulti(
			sb, registrytypes.RecordsByBondIdIndexPrefix, "records_by_bond_id",
			collections.StringKey, collections.StringKey,
			func(_ string, v registrytypes.Record) (string, error) {
				return v.BondId, nil
			},
		),
	}
}

type AuthoritiesIndexes struct {
	AuctionId *indexes.Multi[string, string, registrytypes.NameAuthority]
}

func (a AuthoritiesIndexes) IndexesList() []collections.Index[string, registrytypes.NameAuthority] {
	return []collections.Index[string, registrytypes.NameAuthority]{a.AuctionId}
}

func newAuthorityIndexes(sb *collections.SchemaBuilder) AuthoritiesIndexes {
	return AuthoritiesIndexes{
		AuctionId: indexes.NewMulti(
			sb, registrytypes.AuthoritiesByAuctionIdIndexPrefix, "authorities_by_auction_id",
			collections.StringKey, collections.StringKey,
			func(name string, v registrytypes.NameAuthority) (string, error) {
				return v.AuctionId, nil
			},
		),
	}
}

type NameRecordsIndexes struct {
	Cid *indexes.Multi[string, string, registrytypes.NameRecord]
}

func (b NameRecordsIndexes) IndexesList() []collections.Index[string, registrytypes.NameRecord] {
	return []collections.Index[string, registrytypes.NameRecord]{b.Cid}
}

func newNameRecordIndexes(sb *collections.SchemaBuilder) NameRecordsIndexes {
	return NameRecordsIndexes{
		Cid: indexes.NewMulti(
			sb, registrytypes.NameRecordsByCidIndexPrefix, "name_records_by_cid",
			collections.StringKey, collections.StringKey,
			func(_ string, v registrytypes.NameRecord) (string, error) {
				return v.Latest.Id, nil
			},
		),
	}
}

type Keeper struct {
	cdc           codec.BinaryCodec
	addressCodec  address.Codec
	logger        log.Logger
	headerService header.Service
	eventService  event.Service

	authority types.AccAddress

	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	bondKeeper    *bondkeeper.Keeper
	auctionKeeper *auctionkeeper.Keeper

	// state management
	Schema               collections.Schema
	Params               collections.Item[registrytypes.Params]
	Records              *collections.IndexedMap[string, registrytypes.Record, RecordsIndexes]
	Authorities          *collections.IndexedMap[string, registrytypes.NameAuthority, AuthoritiesIndexes]
	NameRecords          *collections.IndexedMap[string, registrytypes.NameRecord, NameRecordsIndexes]
	RecordExpiryQueue    collections.Map[time.Time, registrytypes.ExpiryQueue]
	AuthorityExpiryQueue collections.Map[time.Time, registrytypes.ExpiryQueue]
	AttributesMap        collections.Map[collections.Pair[string, string], registrytypes.RecordsList]
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	headerService header.Service,
	eventService event.Service,
	accountKeeper auth.AccountKeeper,
	bankKeeper bank.Keeper,
	bondKeeper *bondkeeper.Keeper,
	auctionKeeper *auctionkeeper.Keeper,
	authority types.AccAddress,
	logger log.Logger,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:           cdc,
		addressCodec:  accountKeeper.AddressCodec(),
		headerService: headerService,
		eventService:  eventService,
		logger:        logger.With(log.ModuleKey, "x/"+registrytypes.ModuleName),
		authority:     authority,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		bondKeeper:    bondKeeper,
		auctionKeeper: auctionKeeper,
		Params:        collections.NewItem(sb, registrytypes.ParamsPrefix, "params", codec.CollValue[registrytypes.Params](cdc)),
		Records: collections.NewIndexedMap(
			sb, registrytypes.RecordsPrefix, "records",
			collections.StringKey, codec.CollValue[registrytypes.Record](cdc),
			newRecordIndexes(sb),
		),
		Authorities: collections.NewIndexedMap(
			sb, registrytypes.AuthoritiesPrefix, "authorities",
			collections.StringKey, codec.CollValue[registrytypes.NameAuthority](cdc),
			newAuthorityIndexes(sb),
		),
		NameRecords: collections.NewIndexedMap(
			sb, registrytypes.NameRecordsPrefix, "name_records",
			collections.StringKey, codec.CollValue[registrytypes.NameRecord](cdc),
			newNameRecordIndexes(sb),
		),
		RecordExpiryQueue: collections.NewMap(
			sb, registrytypes.RecordExpiryQueuePrefix, "record_expiry_queue",
			sdk.TimeKey, codec.CollValue[registrytypes.ExpiryQueue](cdc),
		),
		AuthorityExpiryQueue: collections.NewMap(
			sb, registrytypes.AuthorityExpiryQueuePrefix, "authority_expiry_queue",
			sdk.TimeKey, codec.CollValue[registrytypes.ExpiryQueue](cdc),
		),
		AttributesMap: collections.NewMap(
			sb, registrytypes.AttributesMapPrefix, "attributes_map",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey), codec.CollValue[registrytypes.RecordsList](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

// SetParams sets the x/registry module parameters.
func (k Keeper) SetParams(ctx context.Context, params registrytypes.Params) error {
	return k.Params.Set(ctx, params)
}

// HasRecord - checks if a record by the given id exists.
func (k Keeper) HasRecord(ctx context.Context, id string) (bool, error) {
	has, err := k.Records.Has(ctx, id)
	if err != nil {
		return false, err
	}

	return has, nil
}

// PaginatedListRecords - get all records with optional pagination.
func (k Keeper) PaginatedListRecords(ctx context.Context, pagination *query.PageRequest) ([]registrytypes.Record, *query.PageResponse, error) {
	var records []registrytypes.Record
	var pageResp *query.PageResponse

	if pagination == nil {
		err := k.Records.Walk(ctx, nil, func(key string, value registrytypes.Record) (bool, error) {
			if err := k.populateRecordNames(ctx, &value); err != nil {
				return true, err
			}
			records = append(records, value)

			return false, nil
		})
		if err != nil {
			return nil, nil, err
		}
	} else {
		var err error
		records, pageResp, err = query.CollectionPaginate(
			ctx,
			k.Records,
			pagination,
			func(key string, value registrytypes.Record) (registrytypes.Record, error) {
				if err := k.populateRecordNames(ctx, &value); err != nil {
					return registrytypes.Record{}, err
				}

				return value, nil
			})
		if err != nil {
			return nil, nil, err
		}
	}

	return records, pageResp, nil
}

// GetRecordById - gets a record from the store.
func (k Keeper) GetRecordById(ctx context.Context, id string) (registrytypes.Record, error) {
	record, err := k.Records.Get(ctx, id)
	if err != nil {
		return registrytypes.Record{}, err
	}

	if err := k.populateRecordNames(ctx, &record); err != nil {
		return registrytypes.Record{}, err
	}

	return record, nil
}

// GetRecordsByBondId - gets a record from the store.
func (k Keeper) GetRecordsByBondId(ctx context.Context, bondId string) ([]registrytypes.Record, error) {
	var records []registrytypes.Record

	err := k.Records.Indexes.BondId.Walk(ctx, collections.NewPrefixedPairRange[string, string](bondId), func(bondId string, id string) (bool, error) {
		record, err := k.Records.Get(ctx, id)
		if err != nil {
			return true, err
		}

		if err := k.populateRecordNames(ctx, &record); err != nil {
			return true, err
		}
		records = append(records, record)

		return false, nil
	})
	if err != nil {
		return []registrytypes.Record{}, err
	}

	return records, nil
}

// PaginatedRecordsFromAttributes gets a list of records whose attributes match all provided values
// with optional pagination.
func (k Keeper) PaginatedRecordsFromAttributes(
	ctx context.Context,
	attributes []*registrytypes.QueryRecordsRequest_KeyValueInput,
	all bool,
	pagination *query.PageRequest,
) ([]registrytypes.Record, *query.PageResponse, error) {
	var resultRecordIds []string
	var pageResp *query.PageResponse

	filteredRecordIds := []string{}
	for i, attr := range attributes {
		suffix, err := QueryValueToJSON(attr.Value)
		if err != nil {
			return nil, nil, err
		}
		mapKey := collections.Join(attr.Key, string(suffix))
		recordIds, err := k.getAttributeMapping(ctx, mapKey)
		if err != nil {
			return nil, nil, err
		}

		if i == 0 {
			filteredRecordIds = recordIds
		} else {
			filteredRecordIds = getIntersection(recordIds, filteredRecordIds)
		}
	}

	if pagination != nil {
		resultRecordIds, pageResp = paginate(filteredRecordIds, pagination)
	} else {
		resultRecordIds = filteredRecordIds
	}

	records := []registrytypes.Record{}
	for _, id := range resultRecordIds {
		record, err := k.GetRecordById(ctx, id)
		if err != nil {
			return nil, nil, err
		}
		if record.Deleted {
			continue
		}
		if !all && len(record.Names) == 0 {
			continue
		}
		records = append(records, record)
	}

	return records, pageResp, nil
}

// TODO not recursive, and only should be if we want to support querying with whole sub-objects,
// which seems unnecessary.
func QueryValueToJSON(input *registrytypes.QueryRecordsRequest_ValueInput) ([]byte, error) {
	np := basicnode.Prototype.Any
	nb := np.NewBuilder()

	switch value := input.GetValue().(type) {
	case *registrytypes.QueryRecordsRequest_ValueInput_String_:
		err := nb.AssignString(value.String_)
		if err != nil {
			return nil, err
		}
	case *registrytypes.QueryRecordsRequest_ValueInput_Int:
		err := nb.AssignInt(value.Int)
		if err != nil {
			return nil, err
		}
	case *registrytypes.QueryRecordsRequest_ValueInput_Float:
		err := nb.AssignFloat(value.Float)
		if err != nil {
			return nil, err
		}
	case *registrytypes.QueryRecordsRequest_ValueInput_Boolean:
		err := nb.AssignBool(value.Boolean)
		if err != nil {
			return nil, err
		}
	case *registrytypes.QueryRecordsRequest_ValueInput_Link:
		link := cidlink.Link{Cid: cid.MustParse(value.Link)}
		err := nb.AssignLink(link)
		if err != nil {
			return nil, err
		}
	case *registrytypes.QueryRecordsRequest_ValueInput_Array:
		return nil, fmt.Errorf("recursive query values are not supported")
	case *registrytypes.QueryRecordsRequest_ValueInput_Map:
		return nil, fmt.Errorf("recursive query values are not supported")
	default:
		return nil, fmt.Errorf("value has unexpected type %T", value)
	}

	n := nb.Build()
	var buf bytes.Buffer
	if err := dagjson.Encode(n, &buf); err != nil {
		return nil, fmt.Errorf("encoding value to JSON failed: %w", err)
	}
	return buf.Bytes(), nil
}

// PutRecord - saves a record to the store.
func (k Keeper) SaveRecord(ctx context.Context, record registrytypes.Record) error {
	return k.Records.Set(ctx, record.Id, record)
}

// ProcessSetRecord creates a record.
func (k Keeper) SetRecord(ctx context.Context, msg registrytypes.MsgSetRecord) (*registrytypes.ReadableRecord, error) {
	payload := msg.Payload.ToReadablePayload()
	record := registrytypes.ReadableRecord{Attributes: payload.RecordAttributes, BondId: msg.BondId}

	// Check signatures.
	resourceSignBytes, _ := record.GetSignBytes()
	cid, err := record.GetCid()
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid record JSON")
	}

	record.Id = cid

	has, err := k.HasRecord(ctx, record.Id)
	if err != nil {
		return nil, err
	}
	if has {
		// Immutable record already exists. No-op.
		return &record, nil
	}

	record.Owners = []string{}
	for _, sig := range payload.Signatures {
		pubKey, err := legacy.PubKeyFromBytes(helpers.BytesFromBase64(sig.PubKey))
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprint("Error decoding pubKey from bytes: ", err))
		}

		sigOK := pubKey.VerifySignature(resourceSignBytes, helpers.BytesFromBase64(sig.Sig))
		if !sigOK {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprint("Signature mismatch: ", sig.PubKey))
		}
		record.Owners = append(record.Owners, pubKey.Address().String())
	}

	// Sort owners list.
	sort.Strings(record.Owners)
	sdkErr := k.processRecord(ctx, &record)
	if sdkErr != nil {
		return nil, sdkErr
	}

	return &record, nil
}

func (k Keeper) processRecord(ctx context.Context, record *registrytypes.ReadableRecord) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	rent := params.RecordRent
	if err = k.bondKeeper.TransferRentToModuleAccount(
		ctx, record.BondId, registrytypes.RecordRentModuleAccountName, sdk.NewCoins(rent),
	); err != nil {
		return err
	}

	// TODO use HeaderService
	createTime := k.headerService.GetHeaderInfo(ctx).Time
	record.CreateTime = createTime.Format(time.RFC3339)
	record.ExpiryTime = createTime.Add(params.RecordRentDuration).Format(time.RFC3339)
	record.Deleted = false

	recordObj, err := record.ToRecordObj()
	if err != nil {
		return err
	}

	// Save record in store.
	if err = k.SaveRecord(ctx, recordObj); err != nil {
		return err
	}

	// TODO look up/validate record type here

	if err := k.processAttributes(ctx, record.Attributes, record.Id); err != nil {
		return err
	}

	return k.insertRecordExpiryQueue(ctx, recordObj)
}

func (k Keeper) processAttributes(ctx context.Context, attrs registrytypes.AttributeMap, id string) error {
	np := basicnode.Prototype.Map
	nb := np.NewBuilder()
	encAttrs, err := canonicaljson.Marshal(attrs)
	if err != nil {
		return err
	}
	if len(attrs) == 0 {
		encAttrs = []byte("{}")
	}
	err = dagjson.Decode(nb, bytes.NewReader(encAttrs))
	if err != nil {
		return fmt.Errorf("failed to decode attributes: %w", err)
	}
	n := nb.Build()
	if n.Kind() != ipld.Kind_Map {
		return fmt.Errorf("record attributes must be a map, not %T", n.Kind())
	}

	return k.processAttributeMap(ctx, n, id, "")
}

func (k Keeper) processAttributeMap(ctx context.Context, n ipld.Node, id string, prefix string) error {
	for it := n.MapIterator(); !it.Done(); {
		//nolint:misspell
		keynode, valuenode, err := it.Next()
		if err != nil {
			return err
		}
		key, err := keynode.AsString()
		if err != nil {
			return err
		}

		if valuenode.Kind() == ipld.Kind_Map {
			err := k.processAttributeMap(ctx, valuenode, id, key)
			if err != nil {
				return err
			}
		} else {
			var buf bytes.Buffer
			if err := dagjson.Encode(valuenode, &buf); err != nil {
				return err
			}

			value := buf.Bytes()
			mapKey := collections.Join(prefix+key, string(value))
			if err := k.setAttributeMapping(ctx, mapKey, id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (k Keeper) setAttributeMapping(ctx context.Context, key collections.Pair[string, string], recordId string) error {
	var recordIds []string

	has, err := k.AttributesMap.Has(ctx, key)
	if err != nil {
		return err
	}
	if has {
		value, err := k.AttributesMap.Get(ctx, key)
		if err != nil {
			return err
		}
		recordIds = value.Value
	}

	recordIds = append(recordIds, recordId)

	return k.AttributesMap.Set(ctx, key, registrytypes.RecordsList{Value: recordIds})
}

func (k Keeper) getAttributeMapping(ctx context.Context, key collections.Pair[string, string]) ([]string, error) {
	if has, err := k.AttributesMap.Has(ctx, key); !has {
		if err != nil {
			return []string{}, err
		}

		k.logger.Debug(fmt.Sprintf("store doesn't have key: %v", key))
		return []string{}, nil
	}

	value, err := k.AttributesMap.Get(ctx, key)
	if err != nil {
		return []string{}, err
	}

	return value.Value, nil
}

func (k Keeper) populateRecordNames(ctx context.Context, record *registrytypes.Record) error {
	iter, err := k.NameRecords.Indexes.Cid.MatchExact(ctx, record.Id)
	if err != nil {
		return err
	}

	names, err := iter.PrimaryKeys()
	if err != nil {
		return err
	}
	record.Names = names

	return nil
}

// GetModuleBalances gets the registry module account(s) balances.
func (k Keeper) GetModuleBalances(ctx context.Context) []*registrytypes.AccountBalance {
	var balances []*registrytypes.AccountBalance
	accountNames := []string{
		registrytypes.RecordRentModuleAccountName,
		registrytypes.AuthorityRentModuleAccountName,
	}

	for _, accountName := range accountNames {
		moduleAddress := k.accountKeeper.GetModuleAddress(accountName)

		moduleAccount := k.accountKeeper.GetAccount(ctx, moduleAddress)
		if moduleAccount != nil {
			accountBalance := k.bankKeeper.GetAllBalances(ctx, moduleAddress)
			balances = append(balances, &registrytypes.AccountBalance{
				AccountName: accountName,
				Balance:     accountBalance,
			})
		}
	}

	return balances
}

// ProcessRecordExpiryQueue tries to renew expiring records (by collecting rent) else marks them as deleted.
func (k Keeper) ProcessRecordExpiryQueue(ctx context.Context) error {
	cids, err := k.getAllExpiredRecords(ctx, k.headerService.GetHeaderInfo(ctx).Time)
	if err != nil {
		return err
	}

	for _, cid := range cids {
		record, err := k.GetRecordById(ctx, cid)
		if err != nil {
			return err
		}

		bondExists := false
		if record.BondId != "" {
			bondExists, err = k.bondKeeper.HasBond(ctx, record.BondId)
			if err != nil {
				return err
			}
		}

		// If record doesn't have an associated bond or if bond no longer exists, mark it deleted.
		if !bondExists {
			record.Deleted = true
			if err := k.SaveRecord(ctx, record); err != nil {
				return err
			}

			if err := k.deleteRecordExpiryQueue(ctx, record); err != nil {
				return err
			}
		}

		// Try to renew the record by taking rent.
		if err := k.tryTakeRecordRent(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

// getAllExpiredRecords returns a concatenated list of all the timeslices before currTime.
func (k Keeper) getAllExpiredRecords(ctx context.Context, currTime time.Time) ([]string, error) {
	var expiredRecordCIDs []string

	// Get all the records with expiry time until currTime
	rng := new(collections.Range[time.Time]).EndInclusive(currTime)
	err := k.RecordExpiryQueue.Walk(ctx, rng, func(key time.Time, value registrytypes.ExpiryQueue) (stop bool, err error) {
		expiredRecordCIDs = append(expiredRecordCIDs, value.Value...)
		return false, nil
	})
	if err != nil {
		return []string{}, err
	}

	return expiredRecordCIDs, nil
}

// insertRecordExpiryQueue inserts a record CID to the appropriate timeslice in the record expiry queue.
func (k Keeper) insertRecordExpiryQueue(ctx context.Context, record registrytypes.Record) error {
	expiryTime, err := time.Parse(time.RFC3339, record.ExpiryTime)
	if err != nil {
		return err
	}

	existingRecordsList, err := k.RecordExpiryQueue.Get(ctx, expiryTime)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			existingRecordsList = registrytypes.ExpiryQueue{
				Id:    expiryTime.String(),
				Value: []string{},
			}
		} else {
			return err
		}
	}

	existingRecordsList.Value = append(existingRecordsList.Value, record.Id)

	return k.RecordExpiryQueue.Set(ctx, expiryTime, existingRecordsList)
}

// deleteRecordExpiryQueue deletes a record CID from the record expiry queue.
func (k Keeper) deleteRecordExpiryQueue(ctx context.Context, record registrytypes.Record) error {
	expiryTime, err := time.Parse(time.RFC3339, record.ExpiryTime)
	if err != nil {
		return err
	}

	existingRecordsList, err := k.RecordExpiryQueue.Get(ctx, expiryTime)
	if err != nil {
		return err
	}

	newRecordsSlice := []string{}
	for _, id := range existingRecordsList.Value {
		if id != record.Id {
			newRecordsSlice = append(newRecordsSlice, id)
		}
	}

	if len(existingRecordsList.Value) == 0 {
		return k.RecordExpiryQueue.Remove(ctx, expiryTime)
	} else {
		existingRecordsList.Value = newRecordsSlice
		return k.RecordExpiryQueue.Set(ctx, expiryTime, existingRecordsList)
	}
}

// tryTakeRecordRent tries to take rent from the record bond.
func (k Keeper) tryTakeRecordRent(ctx context.Context, record registrytypes.Record) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	rent := params.RecordRent
	sdkErr := k.bondKeeper.TransferRentToModuleAccount(ctx, record.BondId, registrytypes.RecordRentModuleAccountName, sdk.NewCoins(rent))
	if sdkErr != nil {
		// Insufficient funds, mark record as deleted.
		record.Deleted = true
		if err := k.SaveRecord(ctx, record); err != nil {
			return err
		}

		return k.deleteRecordExpiryQueue(ctx, record)
	}

	// Delete old expiry queue entry, create new one.
	if err := k.deleteRecordExpiryQueue(ctx, record); err != nil {
		return err
	}

	record.ExpiryTime = k.headerService.GetHeaderInfo(ctx).Time.Add(params.RecordRentDuration).Format(time.RFC3339)
	if err := k.insertRecordExpiryQueue(ctx, record); err != nil {
		return err
	}

	// Save record.
	record.Deleted = false
	return k.SaveRecord(ctx, record)
}

// paginate implements basic pagination over a list of objects
func paginate[T any](data []T, pagination *query.PageRequest) ([]T, *query.PageResponse) {
	pageReq := initPageRequestDefaults(pagination)

	offset := pageReq.Offset
	limit := pageReq.Limit
	countTotal := pageReq.CountTotal

	totalItems := uint64(len(data))
	start := offset
	end := offset + limit

	if start > totalItems {
		if countTotal {
			return []T{}, &query.PageResponse{Total: 0}
		} else {
			return []T{}, nil
		}
	}
	if end > totalItems {
		end = totalItems
	}

	paginatedItems := data[start:end]

	if countTotal {
		return paginatedItems, &query.PageResponse{Total: totalItems}
	} else {
		return paginatedItems, nil
	}
}

func getIntersection(a []string, b []string) []string {
	result := []string{}
	if len(a) < len(b) {
		for _, str := range a {
			if contains(b, str) {
				result = append(result, str)
			}
		}
	} else {
		for _, str := range b {
			if contains(a, str) {
				result = append(result, str)
			}
		}
	}
	return result
}

func contains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// https://github.com/cosmos/cosmos-sdk/blob/v0.50.3/types/query/pagination.go#L141
// initPageRequestDefaults initializes a PageRequest's defaults when those are not set.
func initPageRequestDefaults(pageRequest *query.PageRequest) *query.PageRequest {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}

	pageRequestCopy := *pageRequest
	if len(pageRequestCopy.Key) == 0 {
		pageRequestCopy.Key = nil
	}

	if pageRequestCopy.Limit == 0 {
		pageRequestCopy.Limit = query.DefaultLimit

		// count total results when the limit is zero/not supplied
		pageRequestCopy.CountTotal = true
	}

	return &pageRequestCopy
}
