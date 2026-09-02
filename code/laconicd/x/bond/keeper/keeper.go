package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	store "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	auth "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	bondtypes "git.vdb.to/cerc-io/laconicd/x/bond"
	nitrokeeper "git.vdb.to/cerc-io/laconicd/x/nitro/keeper"
)

type BondsIndexes struct {
	Owner *indexes.Multi[string, string, bondtypes.Bond]
}

func (b BondsIndexes) IndexesList() []collections.Index[string, bondtypes.Bond] {
	return []collections.Index[string, bondtypes.Bond]{b.Owner}
}

func newBondIndexes(sb *collections.SchemaBuilder) BondsIndexes {
	return BondsIndexes{
		Owner: indexes.NewMulti(
			sb, bondtypes.BondOwnerIndexPrefix, "bonds_by_owner",
			collections.StringKey, collections.StringKey,
			func(_ string, v bondtypes.Bond) (string, error) {
				return v.Owner, nil
			},
		),
	}
}

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	logger       log.Logger
	eventService event.Service
	GasService   gas.Service

	authority types.AccAddress

	// External keepers
	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	nitroKeeper   nitrokeeper.Keeper

	// Track bond usage in other cosmos-sdk modules (more like a usage tracker).
	usageKeepers []bondtypes.BondUsageKeeper

	// State management
	Schema collections.Schema
	Params collections.Item[bondtypes.Params]
	Bonds  *collections.IndexedMap[string, bondtypes.Bond, BondsIndexes]
}

// NewKeeper creates new instances of the bond Keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	eventService event.Service,
	gasService gas.Service,
	accountKeeper auth.AccountKeeper,
	bankKeeper bank.Keeper,
	authority types.AccAddress,
	logger log.Logger,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:           cdc,
		addressCodec:  accountKeeper.AddressCodec(),
		eventService:  eventService,
		GasService:    gasService,
		logger:        logger.With(log.ModuleKey, "x/"+bondtypes.ModuleName),
		authority:     authority,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		Params:        collections.NewItem(sb, bondtypes.ParamsPrefix, "params", codec.CollValue[bondtypes.Params](cdc)),
		Bonds: collections.NewIndexedMap(
			sb, bondtypes.BondsPrefix, "bonds", collections.StringKey, codec.CollValue[bondtypes.Bond](cdc), newBondIndexes(sb),
		),
		usageKeepers: nil,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return &k
}

func (k *Keeper) SetUsageKeepers(usageKeepers []bondtypes.BondUsageKeeper) {
	if k.usageKeepers != nil {
		panic("cannot set bond hooks twice")
	}

	k.usageKeepers = usageKeepers
}

// BondId simplifies generation of bond Ids.
type BondId struct {
	Address  sdk.Address
	AccNum   uint64
	Sequence uint64
}

// Generate creates the bond Id.
func (bondId BondId) Generate(ac address.Codec) string {
	hasher := sha256.New()
	addrStr, err := ac.BytesToString(bondId.Address.Bytes())
	if err != nil {
		panic(err)
	}
	str := fmt.Sprintf("%s:%d:%d", addrStr, bondId.AccNum, bondId.Sequence)
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

// HasBond - checks if a bond by the given Id exists.
func (k Keeper) HasBond(ctx context.Context, id string) (bool, error) {
	has, err := k.Bonds.Has(ctx, id)
	if err != nil {
		return false, err
	}

	return has, nil
}

// SaveBond - saves a bond to the store.
func (k Keeper) SaveBond(ctx context.Context, bond *bondtypes.Bond) error {
	fmt.Println(">>> saving bond:", bond)
	return k.Bonds.Set(ctx, bond.Id, *bond)
}

// DeleteBond - deletes the bond.
func (k Keeper) DeleteBond(ctx context.Context, bond bondtypes.Bond) error {
	return k.Bonds.Remove(ctx, bond.Id)
}

// ListBonds - get all bonds.
func (k Keeper) ListBonds(ctx context.Context) ([]*bondtypes.Bond, error) {
	var bonds []*bondtypes.Bond

	iter, err := k.Bonds.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}

	for ; iter.Valid(); iter.Next() {
		bond, err := iter.Value()
		if err != nil {
			return nil, err
		}

		bonds = append(bonds, &bond)
	}

	return bonds, nil
}

func (k Keeper) GetBondById(ctx context.Context, id string) (bondtypes.Bond, error) {
	bond, err := k.Bonds.Get(ctx, id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			err = errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found")
		}
		return bondtypes.Bond{}, err
	}

	return bond, nil
}

func (k Keeper) GetBondsByOwner(ctx context.Context, owner string) ([]bondtypes.Bond, error) {
	iter, err := k.Bonds.Indexes.Owner.MatchExact(ctx, owner)
	if err != nil {
		return []bondtypes.Bond{}, err
	}

	return indexes.CollectValues(ctx, k.Bonds, iter)
}

// GetBondModuleBalances gets the bond module account(s) balances.
func (k Keeper) GetBondModuleBalances(ctx context.Context) sdk.Coins {
	moduleAddress := k.accountKeeper.GetModuleAddress(bondtypes.ModuleName)
	balances := k.bankKeeper.GetAllBalances(ctx, moduleAddress)

	return balances
}

// CreateBond creates a new bond.
func (k Keeper) CreateBond(ctx context.Context, ownerAddress sdk.AccAddress, coins sdk.Coins) (*bondtypes.Bond, error) {
	// Check if account has funds.
	for _, coin := range coins {
		balance := k.bankKeeper.HasBalance(ctx, ownerAddress, coin)
		if !balance {
			return nil, errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "failed to create bond")
		}
	}

	// Generate bond Id.
	account := k.accountKeeper.GetAccount(ctx, ownerAddress)
	bondId := BondId{
		Address:  ownerAddress,
		AccNum:   account.GetAccountNumber(),
		Sequence: account.GetSequence(),
	}.Generate(k.addressCodec)

	maxBondAmount, err := k.getMaxBondAmount(ctx)
	if err != nil {
		return nil, err
	}

	addrStr, err := k.addressCodec.BytesToString(ownerAddress)
	if err != nil {
		return nil, err
	}
	bond := bondtypes.Bond{Id: bondId, Owner: addrStr, Balance: coins}
	if bond.Balance.IsAnyGT(maxBondAmount) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Max bond amount exceeded")
	}

	// Move funds into the bond account module.
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAddress, bondtypes.ModuleName, bond.Balance)
	if err != nil {
		return nil, err
	}

	// use a payment channel
	// err = k.nitroKeeper.EnsurePaymentChannel(ownerAddress, bondtypes.ModuleAccountAddress, bond.Balance)
	// err = k.nitroKeeper.Pay(ownerAddress, bondtypes.ModuleAccountAddress, bond.Balance)

	// Save bond in store.
	err = k.SaveBond(ctx, &bond)
	if err != nil {
		return nil, err
	}

	return &bond, nil
}

func (k Keeper) RefillBond(ctx context.Context, id string, ownerAddress sdk.AccAddress, coins sdk.Coins) (*bondtypes.Bond, error) {
	if has, err := k.HasBond(ctx, id); !has {
		if err != nil {
			return nil, err
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found")
	}

	bond, err := k.GetBondById(ctx, id)
	if err != nil {
		return nil, err
	}

	addrStr, err := k.addressCodec.BytesToString(ownerAddress)
	if err != nil {
		return nil, err
	}
	if bond.Owner != addrStr {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch")
	}

	// Check if account has funds.
	for _, coin := range coins {
		if !k.bankKeeper.HasBalance(ctx, ownerAddress, coin) {
			return nil, errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "Insufficient funds")
		}
	}

	maxBondAmount, err := k.getMaxBondAmount(ctx)
	if err != nil {
		return nil, err
	}

	updatedBalance := bond.Balance.Add(coins...)
	if updatedBalance.IsAnyGT(maxBondAmount) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Max bond amount exceeded")
	}

	// Nitro: create voucher on existing payment channel with new balance
	// Move funds into the bond account module.
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAddress, bondtypes.ModuleName, coins)
	if err != nil {
		return nil, err
	}

	// Update bond balance and save.
	bond.Balance = updatedBalance
	err = k.SaveBond(ctx, &bond)
	if err != nil {
		return nil, err
	}

	return &bond, nil
}

func (k Keeper) WithdrawBond(ctx context.Context, id string, ownerAddress sdk.AccAddress, coins sdk.Coins) (*bondtypes.Bond, error) {
	if has, err := k.HasBond(ctx, id); !has {
		if err != nil {
			return nil, err
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found")
	}

	bond, err := k.GetBondById(ctx, id)
	if err != nil {
		return nil, err
	}

	addrStr, err := k.addressCodec.BytesToString(ownerAddress)
	if err != nil {
		return nil, err
	}
	if bond.Owner != addrStr {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch")
	}

	updatedBalance, isNeg := bond.Balance.SafeSub(coins...)
	if isNeg {
		return nil, errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "Insufficient bond balance")
	}

	// Nitro: create voucher with lower balance (?)
	// Move funds from the bond into the account.
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, bondtypes.ModuleName, ownerAddress, coins)
	if err != nil {
		return nil, err
	}

	// Update bond balance and save.
	bond.Balance = updatedBalance
	err = k.SaveBond(ctx, &bond)
	if err != nil {
		return nil, err
	}

	return &bond, nil
}

func (k Keeper) CancelBond(ctx context.Context, id string, ownerAddress sdk.AccAddress) (*bondtypes.Bond, error) {
	if has, err := k.HasBond(ctx, id); !has {
		if err != nil {
			return nil, err
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found")
	}

	bond, err := k.GetBondById(ctx, id)
	if err != nil {
		return nil, err
	}

	addrStr, err := k.addressCodec.BytesToString(ownerAddress)
	if err != nil {
		return nil, err
	}
	if bond.Owner != addrStr {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch")
	}

	// Check if bond is used in other modules.
	for _, usageKeeper := range k.usageKeepers {
		if usageKeeper.UsesBond(ctx, id) {
			return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("Bond in use by the '%s' module", usageKeeper.ModuleName()))
		}
	}

	// Nitro: close payment channel, remaining funds just not sent
	// Move funds from the bond into the account.
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, bondtypes.ModuleName, ownerAddress, bond.Balance)
	if err != nil {
		return nil, err
	}

	// Remove bond from store.
	err = k.DeleteBond(ctx, bond)
	if err != nil {
		return nil, err
	}

	return &bond, nil
}

// GetParams gets the bond module's parameters.
func (k Keeper) GetParams(ctx context.Context) (*bondtypes.Params, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &params, nil
}

// SetParams sets the x/bond module parameters.
func (k Keeper) SetParams(ctx context.Context, params bondtypes.Params) error {
	return k.Params.Set(ctx, params)
}

func (k Keeper) getMaxBondAmount(ctx context.Context) (sdk.Coins, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	maxBondAmount := params.MaxBondAmount
	return sdk.NewCoins(maxBondAmount), nil
}

// TransferRentToModuleAccount sends rent from a bond to a module account.
func (k Keeper) TransferRentToModuleAccount(ctx context.Context, id, moduleAccount string, rent sdk.Coins) error {
	bond, err := k.GetBondById(ctx, id)
	if err != nil {
		return err
	}

	// Deduct rent from bond.
	updatedBalance, isNeg := bond.Balance.SafeSub(rent...)
	if isNeg {
		// Check if bond has sufficient funds.
		return errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "Insufficient funds")
	}

	// Nitro: payment on module-module channel
	// Move funds from bond module to record rent module.
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, bondtypes.ModuleName, moduleAccount, rent)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Error transferring funds")
	}

	// Update bond balance.
	bond.Balance = updatedBalance
	err = k.SaveBond(ctx, &bond)

	return err
}
