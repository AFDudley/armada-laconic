package utils

import (
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	addresstypes "github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AddressOrModuleAddress returns with this precedence:
// - the preferred address if it is valid, or
// - a module address with the preferred name, or
// - a module address with the alt name
func AddressOrModuleAddress(preferred, alt string) types.AccAddress {
	if preferred != "" {
		addrCodec := NewAddressCodec()
		if addr, err := addrCodec.StringToBytes(preferred); err == nil {
			return addr
		}
		return addresstypes.Module(preferred)
	}
	return addresstypes.Module(alt)
}

func CheckAuthorityAddress(ac address.Codec, expected sdk.AccAddress, msgaddr string) error {
	authority, err := ac.StringToBytes(msgaddr)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msgaddr)
	}
	if !expected.Equals(types.AccAddress(authority)) {
		return errorsmod.Wrapf(govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s", MustBytesToString(ac, expected), msgaddr)
	}
	return nil
}
