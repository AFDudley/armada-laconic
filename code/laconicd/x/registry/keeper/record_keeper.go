package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
	auctionkeeper "git.vdb.to/cerc-io/laconicd/x/auction/keeper"
	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
)

// Record keeper implements the bond usage keeper interface.
var (
	_ auctiontypes.AuctionUsageKeeper = RecordKeeper{}
	_ bondtypes.BondUsageKeeper       = RecordKeeper{}
)

// RecordKeeper exposes the bare minimal read-only API for other modules.
type RecordKeeper struct {
	cdc           codec.BinaryCodec // The wire codec for binary encoding/decoding.
	k             *Keeper
	auctionKeeper *auctionkeeper.Keeper
	logger        log.Logger
}

// NewRecordKeeper creates new instances of the registry RecordKeeper
func NewRecordKeeper(
	cdc codec.BinaryCodec,
	k *Keeper,
	auctionKeeper *auctionkeeper.Keeper,
	logger log.Logger,
) RecordKeeper {
	return RecordKeeper{
		cdc:           cdc,
		k:             k,
		auctionKeeper: auctionKeeper,
		logger:        logger.With(log.ModuleKey, "x/"+registrytypes.ModuleName),
	}
}

// ModuleName returns the module name.
func (rk RecordKeeper) ModuleName() string {
	return registrytypes.ModuleName
}

func (rk RecordKeeper) UsesAuction(ctx context.Context, auctionId string) bool {
	iter, err := rk.k.Authorities.Indexes.AuctionId.MatchExact(ctx, auctionId)
	if err != nil {
		panic(err)
	}

	return iter.Valid()
}

func (rk RecordKeeper) OnAuctionWinnerSelected(ctx context.Context, auctionId string, blockHeight uint64) {
	// Update authority status based on auction status/winner.
	iter, err := rk.k.Authorities.Indexes.AuctionId.MatchExact(ctx, auctionId)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	names, err := iter.PrimaryKeys()
	if err != nil {
		panic(err)
	}

	if len(names) == 0 {
		// We don't know about this auction, ignore.
		rk.logger.Info(fmt.Sprintf("Ignoring auction notification, name mapping not found: %s", auctionId))
		return
	}

	// Take the first one as an auction (non-empty) will map to only one name
	// MultiIndex being used as there can be multiple entries with empty auction id ("")
	name := names[0]
	if has, err := rk.k.HasNameAuthority(ctx, name); !has {
		if err != nil {
			panic(err)
		}

		// We don't know about this authority, ignore.
		rk.logger.Info(fmt.Sprintf("Ignoring auction notification, authority not found: %s", auctionId))
		return
	}

	authority, err := rk.k.GetNameAuthority(ctx, name)
	if err != nil {
		panic(err)
	}

	auctionObj, err := rk.auctionKeeper.GetAuctionById(ctx, auctionId)
	if err != nil {
		panic(err)
	}

	if auctionObj.Status == auctiontypes.AuctionStatusCompleted {
		if len(auctionObj.WinnerAddresses) != 0 {
			// Mark authority owner and change status to active.
			authority.OwnerAddress = auctionObj.WinnerAddresses[0]
			authority.Status = registrytypes.AuthorityActive

			// Reset bond id if required, as owner has changed.
			authority.BondId = ""

			// Update height for updated/changed authority (owner).
			// Can be used to check if names are older than the authority itself (stale names).
			authority.Height = blockHeight

			rk.logger.Info(fmt.Sprintf("Winner selected, marking authority as active: %s", name))
		} else {
			// Mark as expired.
			authority.Status = registrytypes.AuthorityExpired
			rk.logger.Info(fmt.Sprintf("No winner, marking authority as expired: %s", name))

			rk.logger.Info(fmt.Sprintf("Deleting the expiry queue entry: %s", name))
			if err = rk.k.deleteAuthorityExpiryQueue(ctx, name, authority); err != nil {
				rk.logger.Error("Unable to delete expiry queue entry", err)
			}
		}

		// Forget about this auction now, we no longer need it.
		authority.AuctionId = ""

		if err = rk.k.SaveNameAuthority(ctx, name, &authority); err != nil {
			panic(err)
		}
	} else {
		rk.logger.Info(fmt.Sprintf("Ignoring auction notification, status: %s", auctionObj.Status))
	}
}

// UsesBond returns true if the bond has associated records.
func (rk RecordKeeper) UsesBond(ctx context.Context, bondId string) bool {
	iter, err := rk.k.Records.Indexes.BondId.MatchExact(ctx, bondId)
	if err != nil {
		panic(err)
	}

	return iter.Valid()
}

// RenewRecord renews a record.
func (k Keeper) RenewRecord(ctx context.Context, msg registrytypes.MsgRenewRecord, now time.Time) error {
	if has, err := k.HasRecord(ctx, msg.RecordId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "record not found")
	}

	// Check if renewal is required (i.e. expired record marked as deleted).
	record, err := k.GetRecordById(ctx, msg.RecordId)
	if err != nil {
		return err
	}

	expiryTime, err := time.Parse(time.RFC3339, record.ExpiryTime)
	if err != nil {
		return err
	}

	if !record.Deleted || expiryTime.After(now) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Renewal not required")
	}

	readableRecord := record.ToReadableRecord()
	return k.processRecord(ctx, &readableRecord)
}

// AssociateBond associates a record with a bond.
func (k Keeper) AssociateBond(ctx context.Context, msg registrytypes.MsgAssociateBond) error {
	if has, err := k.HasRecord(ctx, msg.RecordId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "record not found")
	}

	if has, err := k.bondKeeper.HasBond(ctx, msg.BondId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "bond not found")
	}

	// Check if already associated with a bond.
	record, err := k.GetRecordById(ctx, msg.RecordId)
	if err != nil {
		return err
	}

	if len(record.BondId) != 0 {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond already exists")
	}

	// Only the bond owner can associate a record with the bond.
	bond, err := k.bondKeeper.GetBondById(ctx, msg.BondId)
	if err != nil {
		return err
	}
	if msg.Signer != bond.Owner {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch")
	}

	record.BondId = msg.BondId
	if err = k.SaveRecord(ctx, record); err != nil {
		return err
	}

	// Required so that renewal is triggered (with new bond ID) for expired records.
	if record.Deleted {
		return k.insertRecordExpiryQueue(ctx, record)
	}

	return nil
}

// DissociateBond dissociates a record from its bond.
func (k Keeper) DissociateBond(ctx context.Context, msg registrytypes.MsgDissociateBond) error {
	if has, err := k.HasRecord(ctx, msg.RecordId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "record not found")
	}

	record, err := k.GetRecordById(ctx, msg.RecordId)
	if err != nil {
		return err
	}

	// Check if record associated with a bond.
	bondId := record.BondId
	if len(bondId) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "bond not found")
	}

	// Only the bond owner can dissociate a record with the bond.
	bond, err := k.bondKeeper.GetBondById(ctx, bondId)
	if err != nil {
		return err
	}
	if msg.Signer != bond.Owner {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch")
	}

	// Clear bond Id.
	record.BondId = ""
	return k.SaveRecord(ctx, record)
}

// DissociateRecords dissociates all records associated with a given bond.
func (k Keeper) DissociateRecords(ctx context.Context, msg registrytypes.MsgDissociateRecords) error {
	if has, err := k.bondKeeper.HasBond(ctx, msg.BondId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "bond not found")
	}

	// Only the bond owner can dissociate all records from the bond.
	bond, err := k.bondKeeper.GetBondById(ctx, msg.BondId)
	if err != nil {
		return err
	}
	if msg.Signer != bond.Owner {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch")
	}

	// Dissociate all records from the bond.
	records, err := k.GetRecordsByBondId(ctx, msg.BondId)
	if err != nil {
		return err
	}

	for _, record := range records {
		// Clear bond Id.
		record.BondId = ""
		if err = k.SaveRecord(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

// ReassociateRecords switches records from and old to new bond.
func (k Keeper) ReassociateRecords(ctx context.Context, msg registrytypes.MsgReassociateRecords) error {
	if has, err := k.bondKeeper.HasBond(ctx, msg.OldBondId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "old bond not found")
	}

	if has, err := k.bondKeeper.HasBond(ctx, msg.NewBondId); !has {
		if err != nil {
			return err
		}
		return errorsmod.Wrap(sdkerrors.ErrNotFound, "new bond not found")
	}

	// Only the bond owner can re-associate all records.
	oldBond, err := k.bondKeeper.GetBondById(ctx, msg.OldBondId)
	if err != nil {
		return err
	}
	if msg.Signer != oldBond.Owner {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "old bond owner mismatch")
	}

	newBond, err := k.bondKeeper.GetBondById(ctx, msg.NewBondId)
	if err != nil {
		return err
	}
	if msg.Signer != newBond.Owner {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "new bond owner mismatch")
	}

	// Re-associate all records.
	records, err := k.GetRecordsByBondId(ctx, msg.OldBondId)
	if err != nil {
		return err
	}

	for _, record := range records {
		// Switch bond ID.
		record.BondId = msg.NewBondId
		if err = k.SaveRecord(ctx, record); err != nil {
			return err
		}

		// Required so that renewal is triggered (with new bond ID) for expired records.
		if record.Deleted {
			if err = k.insertRecordExpiryQueue(ctx, record); err != nil {
				return err
			}
		}
	}

	return nil
}
