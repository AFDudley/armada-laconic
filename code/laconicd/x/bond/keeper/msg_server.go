package keeper

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/utils"
	"git.vdb.to/cerc-io/laconicd/x/bond"
)

type msgServer struct {
	k *Keeper
}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper *Keeper) msgServer {
	return msgServer{k: keeper}
}

func (ms msgServer) CreateBond(ctx context.Context, msg *bond.MsgCreateBond) (*bond.MsgCreateBondResponse, error) {
	ms.k.logger.Debug("handlers.CreateBond", "msg", msg)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.GasService)
	defer logGas(ms.k.logger, "CreateBond")

	addrCodec := utils.NewAddressCodec()
	signerAddress, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}
	resp, err := ms.k.CreateBond(ctx, signerAddress, msg.Coins)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		bond.EventTypeCreateBond,
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: sdk.AttributeKeyAmount, Value: msg.Coins.String()},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", bond.EventTypeCreateBond)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: bond.AttributeValueCategory},
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &bond.MsgCreateBondResponse{Id: resp.Id}, nil
}

// RefillBond implements bond.MsgServer.
func (ms msgServer) RefillBond(ctx context.Context, msg *bond.MsgRefillBond) (*bond.MsgRefillBondResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.GasService)
	defer logGas(ms.k.logger, "RefillBond")

	addrCodec := utils.NewAddressCodec()
	signerAddress, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	_, err = ms.k.RefillBond(ctx, msg.Id, signerAddress, msg.Coins)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		bond.EventTypeRefillBond,
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: bond.AttributeKeyBondId, Value: msg.Id},
		event.Attribute{Key: sdk.AttributeKeyAmount, Value: msg.Coins.String()},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", bond.EventTypeRefillBond)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: bond.AttributeValueCategory},
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &bond.MsgRefillBondResponse{}, nil
}

// WithdrawBond implements bond.MsgServer.
func (ms msgServer) WithdrawBond(ctx context.Context, msg *bond.MsgWithdrawBond) (*bond.MsgWithdrawBondResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.GasService)
	defer logGas(ms.k.logger, "WithdrawBond")

	addrCodec := utils.NewAddressCodec()
	signerAddress, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	_, err = ms.k.WithdrawBond(ctx, msg.Id, signerAddress, msg.Coins)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		bond.EventTypeWithdrawBond,
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: bond.AttributeKeyBondId, Value: msg.Id},
		event.Attribute{Key: sdk.AttributeKeyAmount, Value: msg.Coins.String()},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", bond.EventTypeWithdrawBond)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: bond.AttributeValueCategory},
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &bond.MsgWithdrawBondResponse{}, nil
}

// CancelBond implements bond.MsgServer.
func (ms msgServer) CancelBond(ctx context.Context, msg *bond.MsgCancelBond) (*bond.MsgCancelBondResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx, logGas := utils.WithCustomGasConfig(ctx, ms.k.GasService)
	defer logGas(ms.k.logger, "CancelBond")

	addrCodec := utils.NewAddressCodec()
	signerAddress, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	_, err = ms.k.CancelBond(ctx, msg.Id, signerAddress)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		bond.EventTypeCancelBond,
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: bond.AttributeKeyBondId, Value: msg.Id},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", bond.EventTypeCancelBond)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: bond.AttributeValueCategory},
		event.Attribute{Key: bond.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &bond.MsgCancelBondResponse{}, nil
}

// UpdateParams defines a method to perform updation of module params.
func (ms msgServer) UpdateParams(ctx context.Context, msg *bond.MsgUpdateParams) (*bond.MsgUpdateParamsResponse, error) {
	if err := utils.CheckAuthorityAddress(ms.k.addressCodec, ms.k.authority, msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &bond.MsgUpdateParamsResponse{}, nil
}
