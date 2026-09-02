package onboarding

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var PERMITTED_ROLES = []string{"participant", "validator"}

var _ sdk.Msg = &MsgOnboardParticipant{}

func (msg MsgOnboardParticipant) ValidateBasic() error {
	if len(msg.Participant) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrorInvalidSigner, msg.Participant)
	}

	if len(msg.EthPayload.Address) != 42 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Participant)
	}

	if len(msg.EthSignature) != 132 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid Eth signature")
	}

	if len(msg.KycId) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Empty KYC ID")
	}

	isRoleValid := false
	for _, v := range PERMITTED_ROLES {
		if msg.Role == v {
			isRoleValid = true
			break
		}
	}

	if !isRoleValid {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("Participant role has to be one of: %v", PERMITTED_ROLES))
	}

	return nil
}
