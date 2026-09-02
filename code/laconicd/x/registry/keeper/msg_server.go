package keeper

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/utils"
	registrytypes "git.vdb.to/cerc-io/laconicd/x/registry"
)

var _ registrytypes.MsgServer = msgServer{}

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) registrytypes.MsgServer {
	return &msgServer{k: keeper}
}

func (ms msgServer) SetRecord(ctx context.Context, msg *registrytypes.MsgSetRecord) (*registrytypes.MsgSetRecordResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	record, err := ms.k.SetRecord(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeSetRecord,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.GetSigner()},
		event.Attribute{Key: registrytypes.AttributeKeyBondId, Value: msg.GetBondId()},
		event.Attribute{Key: registrytypes.AttributeKeyPayload, Value: msg.Payload.Record.Id},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeSetRecord)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgSetRecordResponse{Id: record.Id}, nil
}

// nolint: all
func (ms msgServer) SetName(ctx context.Context, msg *registrytypes.MsgSetName) (*registrytypes.MsgSetNameResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.SetName(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeSetRecord,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyLRN, Value: msg.Lrn},
		event.Attribute{Key: registrytypes.AttributeKeyCID, Value: msg.Cid},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeSetRecord)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgSetNameResponse{}, nil
}

func (ms msgServer) ReserveAuthority(ctx context.Context, msg *registrytypes.MsgReserveAuthority) (*registrytypes.MsgReserveAuthorityResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}
	_, err = addrCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, err
	}

	err = ms.k.ReserveAuthority(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeReserveAuthority,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyName, Value: msg.Name},
		event.Attribute{Key: registrytypes.AttributeKeyOwner, Value: msg.Owner},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeReserveAuthority)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgReserveAuthorityResponse{}, nil
}

// nolint: all
func (ms msgServer) SetAuthorityBond(ctx context.Context, msg *registrytypes.MsgSetAuthorityBond) (*registrytypes.MsgSetAuthorityBondResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.SetAuthorityBond(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeAuthorityBond,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyName, Value: msg.Name},
		event.Attribute{Key: registrytypes.AttributeKeyBondId, Value: msg.BondId},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeAuthorityBond)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgSetAuthorityBondResponse{}, nil
}

func (ms msgServer) DeleteName(ctx context.Context, msg *registrytypes.MsgDeleteName) (*registrytypes.MsgDeleteNameResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.DeleteName(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeDeleteName,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyLRN, Value: msg.Lrn},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeDeleteName)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgDeleteNameResponse{}, nil
}

func (ms msgServer) RenewRecord(ctx context.Context, msg *registrytypes.MsgRenewRecord) (*registrytypes.MsgRenewRecordResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.RenewRecord(ctx, *msg, ms.k.headerService.GetHeaderInfo(ctx).Time)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeRenewRecord,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyRecordId, Value: msg.RecordId},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeRenewRecord)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgRenewRecordResponse{}, nil
}

// nolint: all
func (ms msgServer) AssociateBond(ctx context.Context, msg *registrytypes.MsgAssociateBond) (*registrytypes.MsgAssociateBondResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.AssociateBond(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeAssociateBond,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyRecordId, Value: msg.RecordId},
		event.Attribute{Key: registrytypes.AttributeKeyBondId, Value: msg.BondId},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeAssociateBond)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgAssociateBondResponse{}, nil
}

func (ms msgServer) DissociateBond(ctx context.Context, msg *registrytypes.MsgDissociateBond) (*registrytypes.MsgDissociateBondResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.DissociateBond(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeDissociateBond,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyRecordId, Value: msg.RecordId},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeDissociateBond)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgDissociateBondResponse{}, nil
}

func (ms msgServer) DissociateRecords(
	ctx context.Context,
	msg *registrytypes.MsgDissociateRecords,
) (*registrytypes.MsgDissociateRecordsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.DissociateRecords(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeDissociateRecords,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyBondId, Value: msg.BondId},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeDissociateRecords)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgDissociateRecordsResponse{}, nil
}

func (ms msgServer) ReassociateRecords(ctx context.Context, msg *registrytypes.MsgReassociateRecords) (*registrytypes.MsgReassociateRecordsResponse, error) { //nolint: all
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Registry module doesn't use gas tracking

	addrCodec := utils.NewAddressCodec()
	_, err := addrCodec.StringToBytes(msg.Signer)
	if err != nil {
		return nil, err
	}

	err = ms.k.ReassociateRecords(ctx, *msg)
	if err != nil {
		return nil, err
	}

	eventManager := ms.k.eventService.EventManager(ctx)
	if err := eventManager.EmitKV(
		ctx,
		registrytypes.EventTypeReassociateRecords,
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
		event.Attribute{Key: registrytypes.AttributeKeyOldBondId, Value: msg.OldBondId},
		event.Attribute{Key: registrytypes.AttributeKeyNewBondId, Value: msg.NewBondId},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", registrytypes.EventTypeReassociateRecords)
	}
	if err := eventManager.EmitKV(
		ctx,
		sdk.EventTypeMessage,
		event.Attribute{Key: sdk.AttributeKeyModule, Value: registrytypes.AttributeValueCategory},
		event.Attribute{Key: registrytypes.AttributeKeySigner, Value: msg.Signer},
	); err != nil {
		return nil, errors.Wrapf(err, "failed to emit event: %s", sdk.EventTypeMessage)
	}

	return &registrytypes.MsgReassociateRecordsResponse{}, nil
}

// UpdateParams defines a method to perform updation of module params.
func (ms msgServer) UpdateParams(ctx context.Context, msg *registrytypes.MsgUpdateParams) (*registrytypes.MsgUpdateParamsResponse, error) {
	if err := utils.CheckAuthorityAddress(ms.k.addressCodec, ms.k.authority, msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &registrytypes.MsgUpdateParamsResponse{}, nil
}
