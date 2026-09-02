package keeper

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cosmossdk.io/math"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"git.vdb.to/cerc-io/laconicd/app/params"
	"git.vdb.to/cerc-io/laconicd/utils"
	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
	"git.vdb.to/cerc-io/laconicd/x/registry/helpers"
)

// HasNameAuthority - checks if a name/authority exists.
func (k Keeper) HasNameAuthority(ctx context.Context, name string) (bool, error) {
	has, err := k.Authorities.Has(ctx, name)
	if err != nil {
		return false, err
	}

	return has, nil
}

// GetNameAuthority - gets a name authority from the store.
func (k Keeper) GetNameAuthority(ctx context.Context, name string) (registrytypes.NameAuthority, error) {
	authority, err := k.Authorities.Get(ctx, name)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return registrytypes.NameAuthority{}, errorsmod.Wrap(sdkerrors.ErrNotFound, "name authority not found")
		}
		return registrytypes.NameAuthority{}, err
	}

	return authority, nil
}

// ListNameAuthorityRecords - get all name authority records for given owner
// Returns all authorities if owner set to ""
func (k Keeper) ListNameAuthorityRecords(ctx context.Context, owner string) ([]registrytypes.AuthorityEntry, error) {
	authorityEntries := []registrytypes.AuthorityEntry{}

	err := k.Authorities.Walk(ctx, nil, func(key string, value registrytypes.NameAuthority) (bool, error) {
		// If owner is not empty, skip if authority is not owned by owner
		if owner != "" && owner != value.OwnerAddress {
			return false, nil
		}

		authorityEntries = append(authorityEntries, registrytypes.AuthorityEntry{
			Name:  key,
			Entry: &value,
		})

		// Continue the walk
		return false, nil
	})
	if err != nil {
		return []registrytypes.AuthorityEntry{}, err
	}

	return authorityEntries, nil
}

// HasNameRecord - checks if a name record exists.
func (k Keeper) HasNameRecord(ctx context.Context, lrn string) (bool, error) {
	return k.NameRecords.Has(ctx, lrn)
}

// GetNameRecord - gets a name record from the store.
func (k Keeper) GetNameRecord(ctx context.Context, lrn string) (*registrytypes.NameRecord, error) {
	nameRecord, err := k.NameRecords.Get(ctx, lrn)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &nameRecord, nil
}

// LookupNameRecord - gets a name record which is not stale and under active authority.
func (k Keeper) LookupNameRecord(ctx context.Context, lrn string) (*registrytypes.NameRecord, error) {
	_, _, authority, err := k.getAuthority(ctx, lrn)
	if err != nil || authority.Status != registrytypes.AuthorityActive {
		// If authority is not active (or any other error), lookup fails.
		return nil, nil
	}

	nameRecord, err := k.GetNameRecord(ctx, lrn)

	// Name record may not exist.
	if nameRecord == nil {
		return nil, err
	}

	// Name lookup should fail if the name record is stale.
	// i.e. authority was registered later than the name.
	if authority.Height > nameRecord.Latest.Height {
		return nil, nil
	}

	return nameRecord, nil
}

// ListNameRecords - get all name records.
func (k Keeper) ListNameRecords(ctx context.Context) ([]registrytypes.NameEntry, error) {
	var nameEntries []registrytypes.NameEntry

	err := k.NameRecords.Walk(ctx, nil, func(key string, value registrytypes.NameRecord) (stop bool, err error) {
		nameEntries = append(nameEntries, registrytypes.NameEntry{
			Name:  key,
			Entry: &value,
		})

		return false, nil
	})

	return nameEntries, err
}

// SaveNameRecord - sets a name record.
func (k Keeper) SaveNameRecord(ctx context.Context, lrn string, id string) error {
	var nameRecord registrytypes.NameRecord
	existingNameRecord, err := k.GetNameRecord(ctx, lrn)
	if err != nil {
		return err
	}

	if existingNameRecord != nil {
		nameRecord = *existingNameRecord
		nameRecord.History = append(nameRecord.History, nameRecord.Latest)
	}

	nameRecord.Latest = &registrytypes.NameRecordEntry{
		Id:     id,
		Height: uint64(k.headerService.GetHeaderInfo(ctx).Height),
	}

	return k.NameRecords.Set(ctx, lrn, nameRecord)
}

// SetName creates a LRN -> Record ID mapping.
func (k Keeper) SetName(ctx context.Context, msg registrytypes.MsgSetName) error {
	signerAddress, err := utils.NewAddressCodec().StringToBytes(msg.Signer)
	if err != nil {
		return err
	}
	err = k.checkLRNAccess(ctx, signerAddress, msg.Lrn)
	if err != nil {
		return err
	}

	nameRecord, err := k.LookupNameRecord(ctx, msg.Lrn)
	if err != nil {
		return err
	}
	if nameRecord != nil && nameRecord.Latest.Id == msg.Cid {
		return nil
	}

	return k.SaveNameRecord(ctx, msg.Lrn, msg.Cid)
}

// SaveNameAuthority creates the NameAuthority record.
func (k Keeper) SaveNameAuthority(ctx context.Context, name string, authority *registrytypes.NameAuthority) error {
	return k.Authorities.Set(ctx, name, *authority)
}

// ReserveAuthority reserves a name authority.
func (k Keeper) ReserveAuthority(ctx context.Context, msg registrytypes.MsgReserveAuthority) error {
	lrn := fmt.Sprintf("lrn://%s", msg.GetName())
	parsedLrn, err := url.Parse(lrn)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid name")
	}

	name := parsedLrn.Host
	if fmt.Sprintf("lrn://%s", name) != lrn {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid name")
	}

	if strings.Contains(name, ".") {
		return k.ReserveSubAuthority(ctx, name, msg)
	}

	err = k.createAuthority(ctx, name, msg.GetSigner(), true)
	if err != nil {
		return err
	}

	return nil
}

// ReserveSubAuthority reserves a sub-authority.
func (k Keeper) ReserveSubAuthority(ctx context.Context, name string, msg registrytypes.MsgReserveAuthority) error {
	// Get parent authority name.
	names := strings.Split(name, ".")
	parent := strings.Join(names[1:], ".")

	// Check if parent authority exists.
	if has, err := k.HasNameAuthority(ctx, parent); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "parent authority")
	}
	parentAuthority, err := k.GetNameAuthority(ctx, parent)
	if err != nil {
		return err
	}

	// Sub-authority creator needs to be the owner of the parent authority.
	if parentAuthority.OwnerAddress != msg.Signer {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Access denied")
	}

	// Sub-authority owner defaults to parent authority owner.
	subAuthorityOwner := msg.Signer
	if len(msg.Owner) != 0 {
		// Override sub-authority owner if provided in message.
		subAuthorityOwner = msg.Owner
	}

	sdkErr := k.createAuthority(ctx, name, subAuthorityOwner, false)
	if sdkErr != nil {
		return sdkErr
	}

	return nil
}

func (k Keeper) createAuthority(ctx context.Context, name string, owner string, isRoot bool) error {
	moduleParams, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	has, err := k.HasNameAuthority(ctx, name)
	if err != nil {
		return err
	}
	if has {
		authority, err := k.GetNameAuthority(ctx, name)
		if err != nil {
			return err
		}

		if authority.Status != registrytypes.AuthorityExpired {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Name already reserved")
		}
	}

	ownerAddress, err := utils.NewAddressCodec().StringToBytes(owner)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid owner address")
	}
	ownerAccount := k.accountKeeper.GetAccount(ctx, ownerAddress)
	if ownerAccount == nil {
		return errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "Owner account not found")
	}

	headerInfo := k.headerService.GetHeaderInfo(ctx)
	authority := registrytypes.NameAuthority{
		OwnerPublicKey: getAuthorityPubKey(ownerAccount.GetPubKey()),
		OwnerAddress:   owner,
		Height:         uint64(headerInfo.Height),
		Status:         registrytypes.AuthorityActive,
		AuctionId:      "",
		BondId:         "",
		ExpiryTime:     headerInfo.Time.Add(moduleParams.AuthorityGracePeriod),
	}

	if isRoot && moduleParams.AuthorityAuctionEnabled {
		// If auctions are enabled, clear out owner fields. They will be set after a winner is picked.
		authority.OwnerAddress = ""
		authority.OwnerPublicKey = ""

		// Reset bond ID if required.
		authority.BondId = ""

		// Create an auction.
		msg := auctiontypes.NewMsgCreateAuction(
			auctiontypes.AuctionKindVickrey,
			moduleParams.AuthorityAuctionCommitsDuration,
			moduleParams.AuthorityAuctionRevealsDuration,
			moduleParams.AuthorityAuctionCommitFee,
			moduleParams.AuthorityAuctionRevealFee,
			moduleParams.AuthorityAuctionMinimumBid,
			sdk.NewCoin(params.CoinUnit, math.NewInt(0)),
			0,
			ownerAddress,
		)

		auction, sdkErr := k.auctionKeeper.CreateAuction(ctx, msg)
		if sdkErr != nil {
			return sdkErr
		}

		authority.Status = registrytypes.AuthorityUnderAuction
		authority.AuctionId = auction.Id
		authority.ExpiryTime = auction.RevealsEndTime.Add(moduleParams.AuthorityGracePeriod)
	}

	// Save name authority in store.
	if err = k.SaveNameAuthority(ctx, name, &authority); err != nil {
		return err
	}

	return k.insertAuthorityExpiryQueue(ctx, name, authority.ExpiryTime)
}

func (k Keeper) SetAuthorityBond(ctx context.Context, msg registrytypes.MsgSetAuthorityBond) error {
	name := msg.GetName()
	signer := msg.GetSigner()

	if has, err := k.HasNameAuthority(ctx, name); !has {
		if err != nil {
			return err
		}

		return errorsmod.Wrap(sdkerrors.ErrNotFound, "Name authority not found")
	}

	authority, err := k.GetNameAuthority(ctx, name)
	if err != nil {
		return err
	}
	if authority.OwnerAddress != signer {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Access denied")
	}

	if has, err := k.bondKeeper.HasBond(ctx, msg.BondId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "Bond not found")
	}

	bond, err := k.bondKeeper.GetBondById(ctx, msg.BondId)
	if err != nil {
		return err
	}
	if bond.Owner != signer {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch")
	}

	// No-op if bond hasn't changed.
	if authority.BondId == msg.BondId {
		return nil
	}

	// Update bond id and save name authority in store.
	authority.BondId = bond.Id
	if err = k.SaveNameAuthority(ctx, name, &authority); err != nil {
		return err
	}

	return nil
}

// DeleteName removes a LRN -> Record ID mapping.
func (k Keeper) DeleteName(ctx context.Context, msg registrytypes.MsgDeleteName) error {
	signerAddress, err := utils.NewAddressCodec().StringToBytes(msg.Signer)
	if err != nil {
		return err
	}
	err = k.checkLRNAccess(ctx, signerAddress, msg.Lrn)
	if err != nil {
		return err
	}

	lrnExists, err := k.HasNameRecord(ctx, msg.Lrn)
	if err != nil {
		return err
	}
	if !lrnExists {
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "Name not found")
	}

	// Set CID to empty string.
	return k.SaveNameRecord(ctx, msg.Lrn, "")
}

// ResolveLRN resolves a LRN to a record.
func (k Keeper) ResolveLRN(ctx context.Context, lrn string) (*registrytypes.Record, error) {
	_, _, authority, err := k.getAuthority(ctx, lrn)
	if err != nil || authority.Status != registrytypes.AuthorityActive {
		// If authority is not active (or any other error), resolution fails.
		return nil, err
	}

	// Name should not resolve if it's stale.
	// i.e. authority was registered later than the name.
	record, nameRecord, err := k.resolveLRNRecord(ctx, lrn)
	if err != nil {
		return nil, err
	}
	if authority.Height > nameRecord.Latest.Height {
		return nil, nil
	}

	return record, nil
}

func (k Keeper) resolveLRNRecord(ctx context.Context, lrn string) (*registrytypes.Record, *registrytypes.NameRecord, error) {
	nameRecord, err := k.GetNameRecord(ctx, lrn)
	if nameRecord == nil {
		return nil, nil, err
	}

	latestRecordId := nameRecord.Latest.Id
	if latestRecordId == "" {
		return nil, nameRecord, nil
	}

	if has, err := k.HasRecord(ctx, latestRecordId); !has {
		if err != nil {
			return nil, nil, err
		}

		return nil, nameRecord, nil
	}

	record, err := k.GetRecordById(ctx, latestRecordId)
	if err != nil {
		return nil, nil, err
	}

	return &record, nameRecord, nil
}

func (k Keeper) getAuthority(ctx context.Context, lrn string) (string, *url.URL, *registrytypes.NameAuthority, error) {
	parsedLRN, err := url.Parse(lrn)
	if err != nil {
		return "", nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid LRN")
	}

	name := parsedLRN.Host
	if has, err := k.HasNameAuthority(ctx, name); !has {
		if err != nil {
			return "", nil, nil, err
		}

		return "", nil, nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "name authority not found")
	}

	authority, err := k.GetNameAuthority(ctx, name)
	if err != nil {
		return "", nil, nil, err
	}

	return name, parsedLRN, &authority, nil
}

func (k Keeper) checkLRNAccess(ctx context.Context, signer sdk.AccAddress, lrn string) error {
	name, parsedLRN, authority, err := k.getAuthority(ctx, lrn)
	if err != nil {
		return err
	}

	formattedLRN := fmt.Sprintf("lrn://%s%s", name, parsedLRN.RequestURI())
	if formattedLRN != lrn {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid LRN")
	}

	signerAddr, err := k.addressCodec.BytesToString(signer)
	if err != nil {
		return err
	}
	if authority.OwnerAddress != signerAddr {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Access denied")
	}

	if authority.Status != registrytypes.AuthorityActive {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Authority is not active")
	}

	if authority.BondId == "" || len(authority.BondId) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Authority bond not found")
	}

	if authority.OwnerPublicKey == "" {
		// Try to set owner public key if account has it available now.
		ownerAccount := k.accountKeeper.GetAccount(ctx, signer)
		pubKey := ownerAccount.GetPubKey()
		if pubKey != nil {
			// Update public key in authority record.
			authority.OwnerPublicKey = getAuthorityPubKey(pubKey)
			if err = k.SaveNameAuthority(ctx, name, authority); err != nil {
				return err
			}
		}
	}

	return nil
}

// ProcessAuthorityExpiryQueue tries to renew expiring authorities (by collecting rent) else marks them as expired.
func (k Keeper) ProcessAuthorityExpiryQueue(ctx context.Context) error {
	names, err := k.getAllExpiredAuthorities(ctx, k.headerService.GetHeaderInfo(ctx).Time)
	if err != nil {
		return err
	}

	for _, name := range names {
		authority, err := k.GetNameAuthority(ctx, name)
		if err != nil {
			return err
		}

		bondExists := false
		if authority.BondId != "" {
			bondExists, err = k.bondKeeper.HasBond(ctx, authority.BondId)
			if err != nil {
				return err
			}
		}

		// If authority doesn't have an associated bond or if bond no longer exists, mark it expired.
		if !bondExists {
			authority.Status = registrytypes.AuthorityExpired
			if err = k.SaveNameAuthority(ctx, name, &authority); err != nil {
				return err
			}

			if err = k.deleteAuthorityExpiryQueue(ctx, name, authority); err != nil {
				return err
			}

			k.logger.Info(fmt.Sprintf("Marking authority expired as no bond present: %s", name))

			// Continue with the loop
			continue
		}

		// Try to renew the authority by taking rent.
		if err := k.tryTakeAuthorityRent(ctx, name, authority); err != nil {
			return err
		}
	}

	return nil
}

// getAllExpiredAuthorities returns a concatenated list of all the timeslices before currTime.
func (k Keeper) getAllExpiredAuthorities(ctx context.Context, currTime time.Time) ([]string, error) {
	var expiredAuthorityNames []string

	// Get all the authorities with expiry time until currTime
	rng := new(collections.Range[time.Time]).EndInclusive(currTime)
	err := k.AuthorityExpiryQueue.Walk(ctx, rng, func(key time.Time, value registrytypes.ExpiryQueue) (stop bool, err error) {
		expiredAuthorityNames = append(expiredAuthorityNames, value.Value...)
		return false, nil
	})
	if err != nil {
		return []string{}, err
	}

	return expiredAuthorityNames, nil
}

func (k Keeper) insertAuthorityExpiryQueue(ctx context.Context, name string, expiryTime time.Time) error {
	existingNamesList, err := k.AuthorityExpiryQueue.Get(ctx, expiryTime)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			existingNamesList = registrytypes.ExpiryQueue{
				Id:    expiryTime.String(),
				Value: []string{},
			}
		} else {
			return err
		}
	}

	existingNamesList.Value = append(existingNamesList.Value, name)
	return k.AuthorityExpiryQueue.Set(ctx, expiryTime, existingNamesList)
}

// deleteAuthorityExpiryQueue deletes an authority name from the authority expiry queue.
func (k Keeper) deleteAuthorityExpiryQueue(ctx context.Context, name string, authority registrytypes.NameAuthority) error {
	expiryTime := authority.ExpiryTime

	existingNamesList, err := k.AuthorityExpiryQueue.Get(ctx, expiryTime)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil
		}

		return err
	}

	newNamesSlice := []string{}
	for _, id := range existingNamesList.Value {
		if id != name {
			newNamesSlice = append(newNamesSlice, id)
		}
	}

	if len(newNamesSlice) == 0 {
		return k.AuthorityExpiryQueue.Remove(ctx, expiryTime)
	} else {
		existingNamesList.Value = newNamesSlice
		return k.AuthorityExpiryQueue.Set(ctx, expiryTime, existingNamesList)
	}
}

// tryTakeAuthorityRent tries to take rent from the authority bond.
func (k Keeper) tryTakeAuthorityRent(ctx context.Context, name string, authority registrytypes.NameAuthority) error {
	k.logger.Info(fmt.Sprintf("Trying to take rent for authority: %s", name))

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	rent := params.AuthorityRent
	sdkErr := k.bondKeeper.TransferRentToModuleAccount(ctx, authority.BondId, registrytypes.AuthorityRentModuleAccountName, sdk.NewCoins(rent))
	if sdkErr != nil {
		// Insufficient funds, mark authority as expired.
		authority.Status = registrytypes.AuthorityExpired
		if err := k.SaveNameAuthority(ctx, name, &authority); err != nil {
			return err
		}

		k.logger.Info(fmt.Sprintf("Insufficient funds in owner account to pay authority rent, marking as expired: %s", name))

		return k.deleteAuthorityExpiryQueue(ctx, name, authority)
	}

	// Delete old expiry queue entry, create new one.
	if err := k.deleteAuthorityExpiryQueue(ctx, name, authority); err != nil {
		return err
	}

	authority.ExpiryTime = k.headerService.GetHeaderInfo(ctx).Time.Add(params.AuthorityRentDuration)
	if err := k.insertAuthorityExpiryQueue(ctx, name, authority.ExpiryTime); err != nil {
		return err
	}

	// Save authority.
	authority.Status = registrytypes.AuthorityActive

	k.logger.Info(fmt.Sprintf("Authority rent paid successfully: %s", name))

	return k.SaveNameAuthority(ctx, name, &authority)
}

func getAuthorityPubKey(pubKey cryptotypes.PubKey) string {
	if pubKey == nil {
		return ""
	}

	return helpers.BytesToBase64(pubKey.Bytes())
}
