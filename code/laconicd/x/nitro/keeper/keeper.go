package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"git.vdb.to/cerc-io/laconicd/x/nitro/types"
	"git.vdb.to/cerc-io/laconicd/x/nitro/types/v1"
)

type Keeper struct {
	addressCodec address.Codec
	logger       log.Logger
	eventService event.Service

	// use interface from bank module
	accountKeeper banktypes.AccountKeeper
	// TODO: the way to do this is probably the service + context key pattern
	// distsigManager *distsig.Manager

	LedgerChannels  collections.Map[string, v1.MsgOpenChannel]
	PaymentChannels collections.Map[string, v1.PaymentChannel]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	eventService event.Service,
	accountKeeper banktypes.AccountKeeper,
	logger log.Logger,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Keeper{
		addressCodec:  accountKeeper.AddressCodec(),
		eventService:  eventService,
		logger:        logger.With(log.ModuleKey, "x/"+types.ModuleName),
		accountKeeper: accountKeeper,

		LedgerChannels: collections.NewMap(
			sb, types.ChannelProposalsPrefix, "channel_proposals",
			collections.StringKey, codec.CollValue[v1.MsgOpenChannel](cdc),
		),
		PaymentChannels: collections.NewMap(
			sb, types.PaymentChannelsPrefix, "payment_channels",
			collections.StringKey, codec.CollValue[v1.PaymentChannel](cdc),
		),
	}
}

func (k Keeper) OpenChannel(ctx context.Context, req *v1.MsgOpenChannel) error {
	k.logger.Info("opening channel", "from", req.NitroAddress)

	k.logger.Info("setting open channel", "req", *req)
	return k.LedgerChannels.Set(ctx, req.ChannelId, *req)
}

func (k Keeper) CloseChannel(ctx context.Context, req *v1.MsgCloseChannel) error {
	k.logger.Info("closing channel", "channel_id", req.ChannelId, "is_challenge", req.IsChallenge)

	// Verify the channel exists
	_, err := k.LedgerChannels.Get(ctx, req.ChannelId)
	if err != nil {
		return err
	}

	// Remove the channel from storage upon close request
	// In a full implementation, we might track close requests separately
	// and only remove after actual nitro objective completion
	k.logger.Info("removing closed channel", "channel_id", req.ChannelId)
	return k.LedgerChannels.Remove(ctx, req.ChannelId)
}

func (k Keeper) CreatePaymentChannel(ctx context.Context, req *v1.MsgCreatePaymentChannel) error {
	k.logger.Info("creating payment channel",
		"from", req.NitroAddress,
		"counterparty", req.Counterparty,
		"intermediaries", req.Intermediaries,
		"challenge_duration", req.ChallengeDuration,
		"funds", req.Funds,
		"channel_id", req.ChannelId)

	// Store the payment channel
	paymentChannel := &v1.PaymentChannel{
		ChannelId: req.ChannelId,
	}

	return k.PaymentChannels.Set(ctx, req.ChannelId, *paymentChannel)
}

func (k Keeper) ClosePaymentChannel(ctx context.Context, req *v1.MsgClosePaymentChannel) error {
	k.logger.Info("closing payment channel", "channel_id", req.ChannelId)

	// Verify the payment channel exists
	_, err := k.PaymentChannels.Get(ctx, req.ChannelId)
	if err != nil {
		return err
	}

	// Remove the payment channel from storage upon close request
	k.logger.Info("removing closed payment channel", "channel_id", req.ChannelId)
	return k.PaymentChannels.Remove(ctx, req.ChannelId)
}

func (k Keeper) Pay(ctx context.Context, req *v1.MsgPay) error {
	k.logger.Info("processing payment",
		"channel_id", req.ChannelId,
		"amount", req.Amount)

	if req.Amount.Amount.IsZero() || req.Amount.Amount.IsNegative() {
		return fmt.Errorf("payment amount must be positive, got %s", req.Amount.Amount.String())
	}

	// Verify the payment channel exists
	_, err := k.PaymentChannels.Get(ctx, req.ChannelId)
	if err != nil {
		return fmt.Errorf("payment channel not found: %w", err)
	}

	// No-op implementation for now
	// In a full implementation, this would:
	// - Verify the payment channel has sufficient balance
	// - Verify the requester has permission to make payments from this channel
	// - Create and send the payment voucher through the nitro node
	// - Update channel state/balances
	// - Emit payment events for tracking

	k.logger.Info("payment processed successfully",
		"channel_id", req.ChannelId,
		"amount", req.Amount.String())

	return nil
}
