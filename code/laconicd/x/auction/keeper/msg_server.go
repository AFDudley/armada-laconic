package keeper

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/utils"
	auctiontypes "git.vdb.to/cerc-io/laconicd/x/auction"
)

var _ auctiontypes.MsgServer = msgServer{}

type msgServer struct {
	k *Keeper
}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper *Keeper) auctiontypes.MsgServer {
	return &msgServer{k: keeper}
}

func (ms msgServer) CreateAuction(ctx context.Context, msg *auctiontypes.MsgCreateAuction) (*auctiontypes.MsgCreateAuctionResponse, error) {
	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.gasService)
	defer logGas(ms.k.logger, "CreateAuction")

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	resp, err := ms.k.CreateAuction(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		auctiontypes.EventTypeCreateAuction,
		event.Attribute{Key: auctiontypes.AttributeKeyCommitsDuration, Value: msg.CommitsDuration.String()},
		event.Attribute{Key: auctiontypes.AttributeKeyCommitFee, Value: msg.CommitFee.String()},
		event.Attribute{Key: auctiontypes.AttributeKeyRevealFee, Value: msg.RevealFee.String()},
		event.Attribute{Key: auctiontypes.AttributeKeyMinimumBid, Value: msg.MinimumBid.String()},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", auctiontypes.EventTypeCreateAuction)
	}
	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: auctiontypes.AttributeValueCategory},
		event.Attribute{Key: auctiontypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &auctiontypes.MsgCreateAuctionResponse{Auction: resp}, nil
}

// CommitBid is the command for committing a bid
func (ms msgServer) CommitBid(ctx context.Context, msg *auctiontypes.MsgCommitBid) (*auctiontypes.MsgCommitBidResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.gasService)
	defer logGas(ms.k.logger, "CommitBid")

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	resp, err := ms.k.CommitBid(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		auctiontypes.EventTypeCommitBid,
		event.Attribute{Key: auctiontypes.AttributeKeyAuctionId, Value: msg.AuctionId},
		event.Attribute{Key: auctiontypes.AttributeKeyCommitHash, Value: msg.CommitHash},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", auctiontypes.EventTypeCommitBid)
	}
	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: auctiontypes.AttributeValueCategory},
		event.Attribute{Key: auctiontypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &auctiontypes.MsgCommitBidResponse{Bid: resp}, nil
}

// RevealBid is the command for revealing a bid
func (ms msgServer) RevealBid(ctx context.Context, msg *auctiontypes.MsgRevealBid) (*auctiontypes.MsgRevealBidResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.gasService)
	defer logGas(ms.k.logger, "RevealBid")

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	resp, err := ms.k.RevealBid(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		auctiontypes.EventTypeRevealBid,
		event.Attribute{Key: auctiontypes.AttributeKeyAuctionId, Value: msg.AuctionId},
		event.Attribute{Key: auctiontypes.AttributeKeyReveal, Value: msg.Reveal},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", auctiontypes.EventTypeRevealBid)
	}
	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: auctiontypes.AttributeValueCategory},
		event.Attribute{Key: auctiontypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &auctiontypes.MsgRevealBidResponse{Auction: resp}, nil
}

// UpdateParams defines a method to perform updation of module params.
func (ms msgServer) UpdateParams(ctx context.Context, msg *auctiontypes.MsgUpdateParams) (*auctiontypes.MsgUpdateParamsResponse, error) {
	if err := utils.CheckAuthorityAddress(ms.k.addressCodec, ms.k.authority, msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &auctiontypes.MsgUpdateParamsResponse{}, nil
}

// ReleaseFunds is the command to pay the winning amounts to provider auction winners
func (ms msgServer) ReleaseFunds(ctx context.Context, msg *auctiontypes.MsgReleaseFunds) (*auctiontypes.MsgReleaseFundsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.gasService)
	defer logGas(ms.k.logger, "ReleaseFunds")

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	resp, err := ms.k.ReleaseFunds(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		auctiontypes.EventTypeReleaseFunds,
		event.Attribute{Key: auctiontypes.AttributeKeyAuctionId, Value: msg.AuctionId},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", auctiontypes.EventTypeReleaseFunds)
	}
	if err := ms.k.eventService.EventManager(ctx).EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: auctiontypes.AttributeValueCategory},
		event.Attribute{Key: auctiontypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &auctiontypes.MsgReleaseFundsResponse{Auction: resp}, nil
}
