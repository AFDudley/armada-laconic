package bond

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"git.vdb.to/cerc-io/laconicd/utils"
)

var _ sdk.Msg = &MsgCreateBond{}

// NewMsgCreateBond is the constructor function for MsgCreateBond.
func NewMsgCreateBond(coins sdk.Coins, signer sdk.AccAddress) MsgCreateBond {
	signerStr, err := utils.NewAddressCodec().BytesToString(signer)
	if err != nil {
		panic(err)
	}
	return MsgCreateBond{
		Coins:  coins,
		Signer: signerStr,
	}
}

func (msg MsgCreateBond) ValidateBasic() error {
	if len(msg.Signer) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}
	if len(msg.Coins) == 0 || !msg.Coins.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	return nil
}

func (msg MsgRefillBond) ValidateBasic() error {
	if len(msg.Id) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, msg.Id)
	}
	if len(msg.Signer) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}
	if len(msg.Coins) == 0 || !msg.Coins.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	return nil
}

func (msg MsgWithdrawBond) ValidateBasic() error {
	if len(msg.Id) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, msg.Id)
	}
	if len(msg.Signer) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}
	if len(msg.Coins) == 0 || !msg.Coins.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}
	return nil
}

func (msg MsgCancelBond) ValidateBasic() error {
	if len(msg.Id) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, msg.Id)
	}
	if len(msg.Signer) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}
	return nil
}
