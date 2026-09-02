package keeper

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	store "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	auth "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"git.vdb.to/cerc-io/laconicd/utils"
	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
)

// CompletedAuctionDeleteTimeout => Completed auctions are deleted after this timeout (after reveals end time).
const CompletedAuctionDeleteTimeout = 365 * 24 * time.Hour // 1 year

type AuctionsIndexes struct {
	Owner *indexes.Multi[string, string, auctiontypes.Auction]
}

func (a AuctionsIndexes) IndexesList() []collections.Index[string, auctiontypes.Auction] {
	return []collections.Index[string, auctiontypes.Auction]{a.Owner}
}

func newAuctionIndexes(sb *collections.SchemaBuilder) AuctionsIndexes {
	return AuctionsIndexes{
		Owner: indexes.NewMulti(
			sb, auctiontypes.AuctionOwnerIndexPrefix, "auctions_by_owner",
			collections.StringKey, collections.StringKey,
			func(_ string, v auctiontypes.Auction) (string, error) {
				return v.OwnerAddress, nil
			},
		),
	}
}

type BidsIndexes struct {
	Bidder *indexes.ReversePair[string, string, auctiontypes.Bid]
}

func (b BidsIndexes) IndexesList() []collections.Index[collections.Pair[string, string], auctiontypes.Bid] {
	return []collections.Index[collections.Pair[string, string], auctiontypes.Bid]{b.Bidder}
}

func newBidsIndexes(sb *collections.SchemaBuilder) BidsIndexes {
	return BidsIndexes{
		Bidder: indexes.NewReversePair[auctiontypes.Bid](
			sb, auctiontypes.BidderAuctionIdIndexPrefix, "auction_id_by_bidder",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
		),
	}
}

type Keeper struct {
	cdc           codec.BinaryCodec
	logger        log.Logger
	addressCodec  address.Codec
	headerService header.Service
	eventService  event.Service
	gasService    gas.Service

	authority types.AccAddress

	// External keepers
	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper

	// Track auction usage in other cosmos-sdk modules (more like a usage tracker).
	usageKeepers []auctiontypes.AuctionUsageKeeper

	// state management
	Schema   collections.Schema
	Params   collections.Item[auctiontypes.Params]
	Auctions *collections.IndexedMap[
		string, auctiontypes.Auction, AuctionsIndexes,
	] // map: auctionId -> Auction, index: owner -> Auctions
	Bids *collections.IndexedMap[
		collections.Pair[string, string], auctiontypes.Bid, BidsIndexes,
	] // map: (auctionId, bidder) -> Bid, index: bidder -> auctionId
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	headerService header.Service,
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
		headerService: headerService,
		eventService:  eventService,
		gasService:    gasService,
		logger:        logger.With(log.ModuleKey, "x/"+auctiontypes.ModuleName),
		authority:     authority,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		Params:        collections.NewItem(sb, auctiontypes.ParamsPrefix, "params", codec.CollValue[auctiontypes.Params](cdc)),
		Auctions: collections.NewIndexedMap(
			sb, auctiontypes.AuctionsPrefix, "auctions", collections.StringKey, codec.CollValue[auctiontypes.Auction](cdc), newAuctionIndexes(sb),
		),
		Bids: collections.NewIndexedMap(
			sb, auctiontypes.BidsPrefix, "bids",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[auctiontypes.Bid](cdc),
			newBidsIndexes(sb),
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

func (k *Keeper) SetUsageKeepers(usageKeepers []auctiontypes.AuctionUsageKeeper) {
	if k.usageKeepers != nil {
		panic("cannot set auction hooks twice")
	}

	k.usageKeepers = usageKeepers
}

// SaveAuction - saves a auction to the store.
func (k Keeper) SaveAuction(ctx context.Context, auction *auctiontypes.Auction) error {
	return k.Auctions.Set(ctx, auction.Id, *auction)
}

// DeleteAuction - deletes the auction.
func (k Keeper) DeleteAuction(ctx context.Context, auction auctiontypes.Auction) error {
	// Delete all bids first.
	bids, err := k.GetBids(ctx, auction.Id)
	if err != nil {
		return err
	}

	for _, bid := range bids {
		if err := k.DeleteBid(ctx, *bid); err != nil {
			return err
		}
	}

	return k.Auctions.Remove(ctx, auction.Id)
}

func (k Keeper) HasAuction(ctx context.Context, id string) (bool, error) {
	has, err := k.Auctions.Has(ctx, id)
	if err != nil {
		return false, err
	}

	return has, nil
}

func (k Keeper) SaveBid(ctx context.Context, bid *auctiontypes.Bid) error {
	key := collections.Join(bid.AuctionId, bid.BidderAddress)
	return k.Bids.Set(ctx, key, *bid)
}

func (k Keeper) DeleteBid(ctx context.Context, bid auctiontypes.Bid) error {
	key := collections.Join(bid.AuctionId, bid.BidderAddress)
	return k.Bids.Remove(ctx, key)
}

func (k Keeper) HasBid(ctx context.Context, id string, bidder string) (bool, error) {
	key := collections.Join(id, bidder)
	has, err := k.Bids.Has(ctx, key)
	if err != nil {
		return false, err
	}

	return has, nil
}

func (k Keeper) GetBid(ctx context.Context, id string, bidder string) (auctiontypes.Bid, error) {
	key := collections.Join(id, bidder)
	bid, err := k.Bids.Get(ctx, key)
	if err != nil {
		return auctiontypes.Bid{}, err
	}

	return bid, nil
}

// GetBids gets the auction bids.
func (k Keeper) GetBids(ctx context.Context, id string) ([]*auctiontypes.Bid, error) {
	var bids []*auctiontypes.Bid

	err := k.Bids.Walk(ctx,
		collections.NewPrefixedPairRange[string, string](id),
		func(
			key collections.Pair[string, string],
			value auctiontypes.Bid) (stop bool, err error,
		) {
			bids = append(bids, &value)
			return false, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return bids, nil
}

// ListAuctions - get all auctions.
func (k Keeper) ListAuctions(ctx context.Context) ([]auctiontypes.Auction, error) {
	iter, err := k.Auctions.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}

	return iter.Values()
}

// MatchAuctions - get all matching auctions.
func (k Keeper) MatchAuctions(ctx context.Context, matchFn func(*auctiontypes.Auction) (bool, error)) ([]*auctiontypes.Auction, error) {
	var auctions []*auctiontypes.Auction

	err := k.Auctions.Walk(ctx, nil, func(key string, value auctiontypes.Auction) (bool, error) {
		auctionMatched, err := matchFn(&value)
		if err != nil {
			return true, err
		}

		if auctionMatched {
			auctions = append(auctions, &value)
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return auctions, nil
}

// GetAuction - gets a record from the store.
func (k Keeper) GetAuctionById(ctx context.Context, id string) (auctiontypes.Auction, error) {
	auction, err := k.Auctions.Get(ctx, id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return auctiontypes.Auction{}, errorsmod.Wrap(sdkerrors.ErrNotFound, "Auction not found")
		}
		return auctiontypes.Auction{}, err
	}

	return auction, nil
}

func (k Keeper) GetAuctionsByOwner(ctx context.Context, owner string) ([]auctiontypes.Auction, error) {
	iter, err := k.Auctions.Indexes.Owner.MatchExact(ctx, owner)
	if err != nil {
		return nil, err
	}

	return indexes.CollectValues(ctx, k.Auctions, iter)
}

// QueryAuctionsByBidder - query auctions by bidder
func (k Keeper) QueryAuctionsByBidder(ctx context.Context, bidderAddress string) ([]auctiontypes.Auction, error) {
	auctions := []auctiontypes.Auction{}

	iter, err := k.Bids.Indexes.Bidder.MatchExact(ctx, bidderAddress)
	if err != nil {
		return nil, err
	}

	for ; iter.Valid(); iter.Next() {
		keyPair, err := iter.PrimaryKey()
		if err != nil {
			return nil, err
		}

		auction, err := k.GetAuctionById(ctx, keyPair.K1())
		if err != nil {
			return nil, err
		}

		auctions = append(auctions, auction)
	}

	return auctions, nil
}

// CreateAuction creates a new auction.
func (k Keeper) CreateAuction(ctx context.Context, msg auctiontypes.MsgCreateAuction) (*auctiontypes.Auction, error) {
	// Might be called from another module directly, always validate.
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	signerAddress, err := utils.NewAddressCodec().StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Generate auction Id.
	account := k.accountKeeper.GetAccount(ctx, signerAddress)
	if account == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, msg.Signer)
	}

	auctionId := auctiontypes.AuctionId{
		Address:  sdk.AccAddress(signerAddress),
		AccNum:   account.GetAccountNumber(),
		Sequence: account.GetSequence(),
	}.Generate(k.addressCodec)

	// Compute timestamps.
	now := k.headerService.GetHeaderInfo(ctx).Time
	commitsEndTime := now.Add(msg.CommitsDuration)
	revealsEndTime := now.Add(msg.CommitsDuration + msg.RevealsDuration)

	if msg.Kind == auctiontypes.AuctionKindProvider {
		totalLockedAmount := sdk.NewCoin(msg.MaxPrice.Denom, msg.MaxPrice.Amount.MulRaw(int64(msg.NumProviders)))

		sdkErr := k.bankKeeper.SendCoinsFromAccountToModule(ctx, signerAddress, auctiontypes.ModuleName, sdk.NewCoins(totalLockedAmount))
		if sdkErr != nil {
			return nil, errorsmod.Wrap(sdkErr, "Error transferring maximum price amount")
		}
	}

	auction := auctiontypes.Auction{
		Id:             auctionId,
		Kind:           msg.Kind,
		Status:         auctiontypes.AuctionStatusCommitPhase,
		OwnerAddress:   msg.Signer,
		CreateTime:     now,
		CommitsEndTime: commitsEndTime,
		RevealsEndTime: revealsEndTime,
		CommitFee:      msg.CommitFee,
		RevealFee:      msg.RevealFee,
		MinimumBid:     msg.MinimumBid,
		MaxPrice:       msg.MaxPrice,
		NumProviders:   msg.NumProviders,
	}

	// Save auction in store.
	if err = k.SaveAuction(ctx, &auction); err != nil {
		return nil, err
	}

	return &auction, nil
}

func (k Keeper) CommitBid(ctx context.Context, msg auctiontypes.MsgCommitBid) (*auctiontypes.Bid, error) {
	if has, err := k.HasAuction(ctx, msg.AuctionId); !has {
		if err != nil {
			return nil, err
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "Auction not found")
	}

	auction, err := k.GetAuctionById(ctx, msg.AuctionId)
	if err != nil {
		return nil, err
	}

	if auction.Status != auctiontypes.AuctionStatusCommitPhase {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Auction is not in commit phase")
	}

	addrCodec := utils.NewAddressCodec()
	signerAddress, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Take auction fees from account.
	totalFee := auction.CommitFee.Add(auction.RevealFee)
	sdkErr := k.bankKeeper.SendCoinsFromAccountToModule(ctx, signerAddress, auctiontypes.ModuleName, sdk.NewCoins(totalFee))
	if sdkErr != nil {
		return nil, sdkErr
	}

	// Check if an old bid already exists, if so, return old bids auction fee (update bid scenario).
	bidExists, err := k.HasBid(ctx, msg.AuctionId, msg.Signer)
	if err != nil {
		return nil, err
	}

	if bidExists {
		oldBid, err := k.GetBid(ctx, msg.AuctionId, msg.Signer)
		if err != nil {
			return nil, err
		}

		oldTotalFee := oldBid.CommitFee.Add(oldBid.RevealFee)
		sdkErr := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, auctiontypes.ModuleName, signerAddress, sdk.NewCoins(oldTotalFee))
		if sdkErr != nil {
			return nil, sdkErr
		}
	}

	// Save new bid.
	bid := auctiontypes.Bid{
		AuctionId:     msg.AuctionId,
		BidderAddress: msg.Signer,
		Status:        auctiontypes.BidStatusCommitted,
		CommitHash:    msg.CommitHash,
		CommitTime:    k.headerService.GetHeaderInfo(ctx).Time,
		CommitFee:     auction.CommitFee,
		RevealFee:     auction.RevealFee,
	}

	if err = k.SaveBid(ctx, &bid); err != nil {
		return nil, err
	}

	return &bid, nil
}

func (k Keeper) RevealBid(ctx context.Context, msg auctiontypes.MsgRevealBid) (*auctiontypes.Auction, error) {
	if has, err := k.HasAuction(ctx, msg.AuctionId); !has {
		if err != nil {
			return nil, err
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "Auction not found")
	}

	auction, err := k.GetAuctionById(ctx, msg.AuctionId)
	if err != nil {
		return nil, err
	}

	if auction.Status != auctiontypes.AuctionStatusRevealPhase {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Auction is not in reveal phase")
	}

	addrCodec := utils.NewAddressCodec()
	signerAddress, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	bidExists, err := k.HasBid(ctx, msg.AuctionId, msg.Signer)
	if err != nil {
		return nil, err
	}

	if !bidExists {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotFound, "Bid not found")
	}

	bid, err := k.GetBid(ctx, msg.AuctionId, msg.Signer)
	if err != nil {
		return nil, err
	}

	if bid.Status != auctiontypes.BidStatusCommitted {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Bid not in committed state")
	}

	revealBytes, err := hex.DecodeString(msg.Reveal)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal string")
	}

	cid, err := utils.CIDFromJSONBytes(revealBytes)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal JSON")
	}

	if bid.CommitHash != cid {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Commit hash mismatch")
	}

	var reveal map[string]interface{}
	err = json.Unmarshal(revealBytes, &reveal)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Reveal JSON unmarshal error")
	}

	headerInfo := k.headerService.GetHeaderInfo(ctx)
	chainId, err := utils.GetAttributeAsString(reveal, "chainId")
	if err != nil || chainId != headerInfo.ChainID {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal chainID")
	}

	auctionId, err := utils.GetAttributeAsString(reveal, "auctionId")
	if err != nil || auctionId != msg.AuctionId {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal auction Id")
	}

	bidderAddress, err := utils.GetAttributeAsString(reveal, "bidderAddress")
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal bid address")
	}

	if bidderAddress != msg.Signer {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Reveal bid address mismatch")
	}

	bidAmountStr, err := utils.GetAttributeAsString(reveal, "bidAmount")
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal bid amount")
	}

	bidAmount, err := sdk.ParseCoinNormalized(bidAmountStr)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal bid amount")
	}

	if auction.Kind == auctiontypes.AuctionKindVickrey && bidAmount.IsLT(auction.MinimumBid) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Bid is lower than minimum bid")
	}

	if auction.Kind == auctiontypes.AuctionKindProvider && auction.MaxPrice.IsLT(bidAmount) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Bid is higher than max price")
	}

	// Lock bid amount.
	if auction.Kind == auctiontypes.AuctionKindVickrey {
		sdkErr := k.bankKeeper.SendCoinsFromAccountToModule(ctx, signerAddress, auctiontypes.ModuleName, sdk.NewCoins(bidAmount))
		if sdkErr != nil {
			return nil, sdkErr
		}
	}

	// Update bid.
	bid.BidAmount = bidAmount
	bid.RevealTime = headerInfo.Time
	bid.Status = auctiontypes.BidStatusRevealed
	if err = k.SaveBid(ctx, &bid); err != nil {
		return nil, err
	}

	return &auction, nil
}

// GetParams gets the auction module's parameters.
func (k Keeper) GetParams(ctx context.Context) (*auctiontypes.Params, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &params, nil
}

// SetParams sets the x/auction module parameters.
func (k Keeper) SetParams(ctx context.Context, params auctiontypes.Params) error {
	return k.Params.Set(ctx, params)
}

// GetAuctionModuleBalances gets the auction module account(s) balances.
func (k Keeper) GetAuctionModuleBalances(ctx context.Context) sdk.Coins {
	moduleAddress := k.accountKeeper.GetModuleAddress(auctiontypes.ModuleName)
	balances := k.bankKeeper.GetAllBalances(ctx, moduleAddress)

	return balances
}

func (k Keeper) EndBlockerProcessAuctions(ctx context.Context) error {
	// Transition auction state (commit, reveal, expired, completed).
	if err := k.processAuctionPhases(ctx); err != nil {
		return err
	}

	// Delete stale auctions.
	return k.deleteCompletedAuctions(ctx)
}

func (k Keeper) processAuctionPhases(ctx context.Context) error {
	auctions, err := k.MatchAuctions(ctx, func(_ *auctiontypes.Auction) (bool, error) {
		return true, nil
	})
	if err != nil {
		return err
	}

	now := k.headerService.GetHeaderInfo(ctx).Time
	for _, auction := range auctions {
		// Commit -> Reveal state.
		if auction.Status == auctiontypes.AuctionStatusCommitPhase && now.After(auction.CommitsEndTime) {
			auction.Status = auctiontypes.AuctionStatusRevealPhase
			if err = k.SaveAuction(ctx, auction); err != nil {
				return err
			}

			k.logger.Info("Moved auction to reveal phase", "id", auction.Id)
		}

		// Reveal -> Expired state.
		if auction.Status == auctiontypes.AuctionStatusRevealPhase && now.After(auction.RevealsEndTime) {
			auction.Status = auctiontypes.AuctionStatusExpired
			if err = k.SaveAuction(ctx, auction); err != nil {
				return err
			}

			k.logger.Info("Moved auction to expired state", "id", auction.Id)
		}

		// If auction has expired, pick a winner from revealed bids.
		if auction.Status == auctiontypes.AuctionStatusExpired {
			if auction.Kind == auctiontypes.AuctionKindVickrey {
				if err = k.pickAuctionWinner(ctx, auction); err != nil {
					return err
				}
			} else {
				if err = k.pickProviderAuctionWinners(ctx, auction); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Delete completed stale auctions.
func (k Keeper) deleteCompletedAuctions(ctx context.Context) error {
	auctions, err := k.MatchAuctions(ctx, func(auction *auctiontypes.Auction) (bool, error) {
		deleteTime := auction.RevealsEndTime.Add(CompletedAuctionDeleteTimeout)
		return auction.Status == auctiontypes.AuctionStatusCompleted && k.headerService.GetHeaderInfo(ctx).Time.After(deleteTime), nil
	})
	if err != nil {
		return err
	}

	for _, auction := range auctions {
		k.logger.Info("Deleting completed auction after timeout", "id", auction.Id)
		if err := k.DeleteAuction(ctx, *auction); err != nil {
			return err
		}
	}

	return nil
}

// Pick winner for vickrey auction
func (k Keeper) pickAuctionWinner(ctx context.Context, auction *auctiontypes.Auction) error {
	logger := k.logger.With("id", auction.Id)
	logger.Info("Picking auction winner")

	var highestBid *auctiontypes.Bid
	var secondHighestBid *auctiontypes.Bid

	bids, err := k.GetBids(ctx, auction.Id)
	if err != nil {
		return err
	}

	for _, bid := range bids {
		logger.Info("Processing bid", "address", bid.BidderAddress, "amount", bid.BidAmount)

		// Only consider revealed bids.
		if bid.Status != auctiontypes.BidStatusRevealed {
			logger.Info("Ignoring unrevealed bid", "address", bid.BidderAddress, "amount", bid.BidAmount)
			continue
		}

		// Init highest bid.
		if highestBid == nil {
			highestBid = bid
			logger.Info("Initializing 1st bid", "address", bid.BidderAddress, "amount", bid.BidAmount)
			continue
		}

		//nolint: all
		if highestBid.BidAmount.IsLT(bid.BidAmount) {
			logger.Info("New highest bid", "address", bid.BidderAddress, "amount", bid.BidAmount)

			secondHighestBid = highestBid
			highestBid = bid

			logger.Info("Updated 1st bid", "address", highestBid.BidderAddress, "amount", highestBid.BidAmount)
			logger.Info("Updated 2nd bid", "address", secondHighestBid.BidderAddress, "amount", secondHighestBid.BidAmount)

		} else if secondHighestBid == nil || secondHighestBid.BidAmount.IsLT(bid.BidAmount) {
			logger.Info("New 2nd highest bid", "address", bid.BidderAddress, "amount", bid.BidAmount)

			secondHighestBid = bid
			logger.Info("Updated 2nd bid", "address", secondHighestBid.BidderAddress, "amount", secondHighestBid.BidAmount)
		} else {
			logger.Info("Ignoring bid as it doesn't affect 1st/2nd price", "address", bid.BidderAddress, "amount", bid.BidAmount)
		}
	}

	// Highest bid is the winner, but pays second highest bid price.
	auction.Status = auctiontypes.AuctionStatusCompleted

	if highestBid != nil {
		auction.WinningBids = []sdk.Coin{highestBid.BidAmount}
		auction.WinnerAddresses = []string{highestBid.BidderAddress}

		// Winner pays 2nd price, if a 2nd price exists.
		auction.WinningPrice = highestBid.BidAmount
		if secondHighestBid != nil {
			auction.WinningPrice = secondHighestBid.BidAmount
		}
		logger.Info("Auction winner chosen",
			"address", auction.WinnerAddresses[0],
			"bid", auction.WinningBids[0],
			"price", auction.WinningPrice)
	} else {
		logger.Info("Auction has no valid revealed bids (no winner)")
	}

	if err := k.SaveAuction(ctx, auction); err != nil {
		return err
	}

	addrCodec := utils.NewAddressCodec()

	for _, bid := range bids {
		bidderAddress, err := addrCodec.StringToBytes(bid.BidderAddress)
		if err != nil {
			logger.Error("Invalid bidderAddress address", "error", err)
			panic("Invalid bidder address")
		}

		if bid.Status == auctiontypes.BidStatusRevealed {
			// Send reveal fee back to bidders that've revealed the bid.
			sdkErr := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, auctiontypes.ModuleName, bidderAddress, sdk.NewCoins(bid.RevealFee))
			if sdkErr != nil {
				logger.Error("Error returning reveal fee", "error", sdkErr)
				panic(sdkErr)
			}
		}

		// Send back locked bid amount to all bidders.
		sdkErr := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, auctiontypes.ModuleName, bidderAddress, sdk.NewCoins(bid.BidAmount))
		if sdkErr != nil {
			logger.Error("Error returning bid amount", "error", sdkErr)
			panic(sdkErr)
		}
	}

	// Process winner account (if nobody bids, there won't be a winner).
	if len(auction.WinnerAddresses) != 0 {
		winnerAddress, err := addrCodec.StringToBytes(auction.WinnerAddresses[0])
		if err != nil {
			logger.Error("Invalid winner address", "error", err)
			panic("Invalid winner address")
		}

		// Take 2nd price from winner.
		sdkErr := k.bankKeeper.SendCoinsFromAccountToModule(ctx, winnerAddress, auctiontypes.ModuleName, sdk.NewCoins(auction.WinningPrice))
		if sdkErr != nil {
			logger.Error("Error taking funds from winner", "error", sdkErr)
			panic(sdkErr)
		}

		// Burn anything over the min. bid amount.
		amountToBurn := auction.WinningPrice.Sub(auction.MinimumBid)
		if amountToBurn.IsNegative() {
			logger.Error("Auction coins to burn cannot be negative")
			panic("Auction coins to burn cannot be negative")
		}

		// Use auction burn module account instead of actually burning coins to better keep track of supply.
		sdkErr = k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			auctiontypes.ModuleName,
			auctiontypes.AuctionBurnModuleAccountName,
			sdk.NewCoins(amountToBurn),
		)
		if sdkErr != nil {
			logger.Error("Error burning coins", "error", sdkErr)
			panic(sdkErr)
		}
	}

	// Notify other modules (hook).
	// k.logger.Info(fmt.Sprintf("Auction notifying %d modules", len(k.usageKeepers)))
	for _, keeper := range k.usageKeepers {
		logger.Info("Auction notifying module", "module", keeper.ModuleName())
		keeper.OnAuctionWinnerSelected(ctx, auction.Id, uint64(k.headerService.GetHeaderInfo(ctx).Height))
	}

	return nil
}

// Pick winner for provider auction
func (k Keeper) pickProviderAuctionWinners(ctx context.Context, auction *auctiontypes.Auction) error {
	logger := k.logger.With("id", auction.Id)
	logger.Info("Picking auction winners", auction.Id)

	bids, err := k.GetBids(ctx, auction.Id)
	if err != nil {
		return err
	}

	revealedBids := make([]*auctiontypes.Bid, 0, len(bids))
	for _, bid := range bids {
		logger.Info("Processing bid", "address", bid.BidderAddress, "amount", bid.BidAmount)

		// Only consider revealed bids.
		if bid.Status != auctiontypes.BidStatusRevealed {
			logger.Info("Ignoring unrevealed bid", "address", bid.BidderAddress, "amount", bid.BidAmount)
			continue
		}

		revealedBids = append(revealedBids, bid)
	}

	// Sort the valid bids
	slices.SortStableFunc(revealedBids, func(a, b *auctiontypes.Bid) int {
		if a.BidAmount.Amount.LT(b.BidAmount.Amount) {
			return -1
		} else if a.BidAmount.Amount.GT(b.BidAmount.Amount) {
			return 1
		}
		return 0
	})

	// Take best min(len(revealedBids), auction.NumProviders) bids
	numWinners := int(auction.NumProviders)
	if len(revealedBids) < numWinners {
		numWinners = len(revealedBids)
	}
	winnerBids := revealedBids[:numWinners]

	auction.Status = auctiontypes.AuctionStatusCompleted

	if len(winnerBids) > 0 {
		winnerAddresses := make([]string, len(winnerBids))
		winningBids := make([]sdk.Coin, len(winnerBids))
		for i, bid := range winnerBids {
			winnerAddresses[i] = bid.BidderAddress
			winningBids[i] = bid.BidAmount
		}

		auction.WinnerAddresses = winnerAddresses
		auction.WinningBids = winningBids

		// The last best bid is the winning price
		auction.WinningPrice = winnerBids[len(winnerBids)-1].BidAmount

		for _, bid := range winnerBids {
			logger.Info("Auction winner", "address", bid.BidderAddress, "bid", bid.BidAmount)
		}
		logger.Info("Auction winning price", "price", auction.WinningPrice)
	} else {
		logger.Info("Auction has no valid revealed bids (no winner)")
	}

	if err := k.SaveAuction(ctx, auction); err != nil {
		return err
	}

	addrCodec := utils.NewAddressCodec()

	for _, bid := range bids {
		bidderAddress, err := addrCodec.StringToBytes(bid.BidderAddress)
		if err != nil {
			logger.Error("Invalid bidderAddress address", "error", err)
			panic("Invalid bidder address")
		}

		if bid.Status == auctiontypes.BidStatusRevealed {
			// Send reveal fee back to bidders that've revealed the bid.
			sdkErr := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, auctiontypes.ModuleName, bidderAddress, sdk.NewCoins(bid.RevealFee))
			if sdkErr != nil {
				logger.Error("Error returning reveal fee", "error", sdkErr)
				panic(sdkErr)
			}
		}
	}

	// Send back any leftover locked amount to auction creator
	// All of it in case of no winners
	totalLockedAmount := auction.MaxPrice.Amount.Mul(math.NewInt(int64(auction.NumProviders)))
	totalAmountPaid := auction.WinningPrice.Amount.Mul(math.NewInt(int64(len(auction.WinnerAddresses))))
	creatorLeftOverAmount := sdk.NewCoin(auction.MaxPrice.Denom, totalLockedAmount.Sub(totalAmountPaid))

	ownerAccAddress, err := addrCodec.StringToBytes(auction.OwnerAddress)
	if err != nil {
		logger.Error("Invalid auction owner address", "error", err)
		panic("Invalid auction owner address")
	}

	sdkErr := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		auctiontypes.ModuleName,
		ownerAccAddress,
		sdk.NewCoins(creatorLeftOverAmount),
	)
	if sdkErr != nil {
		logger.Error("Error returning leftover locked amount", "error", sdkErr)
		panic(sdkErr)
	}

	// Notify other modules (hook).
	for _, keeper := range k.usageKeepers {
		logger.Info("Auction notifying module", "module", keeper.ModuleName())
		keeper.OnAuctionWinnerSelected(ctx, auction.Id, uint64(k.headerService.GetHeaderInfo(ctx).Height))
	}

	return nil
}

func (k Keeper) ReleaseFunds(ctx context.Context, msg auctiontypes.MsgReleaseFunds) (*auctiontypes.Auction, error) {
	auction, err := k.GetAuctionById(ctx, msg.AuctionId)
	if err != nil {
		return nil, err
	}

	if auction.Kind != auctiontypes.AuctionKindProvider {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Auction kind must be provider")
	}

	// Only the auction owner can release funds.
	if msg.Signer != auction.OwnerAddress {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "Only auction owner can release funds")
	}

	if auction.Status != auctiontypes.AuctionStatusCompleted {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Auction is not completed")
	}

	if auction.FundsReleased {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Auction funds already released")
	}

	// Mark funds as released in the stored auction
	auction.FundsReleased = true
	if err = k.SaveAuction(ctx, &auction); err != nil {
		return nil, err
	}

	addrCodec := utils.NewAddressCodec()
	logger := k.logger.With("id", auction.Id)

	// Process winner accounts.
	for _, winner := range auction.WinnerAddresses {
		winnerAddress, err := addrCodec.StringToBytes(winner)
		if err != nil {
			logger.Error("Invalid winner address", "error", err)
			panic("Invalid winner address")
		}

		// Send winning price to winning bidders
		sdkErr := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			auctiontypes.ModuleName,
			winnerAddress,
			sdk.NewCoins(auction.WinningPrice),
		)
		if sdkErr != nil {
			logger.Error("Error sending funds to winner", "error", sdkErr)
			panic(sdkErr)
		}
	}

	return &auction, err
}
