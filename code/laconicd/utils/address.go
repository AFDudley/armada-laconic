package utils

import (
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"git.vdb.to/cerc-io/laconicd/app/params"
)

type addressCodec struct {
	address.Codec
}

func (ac addressCodec) StringToBytes(text string) ([]byte, error) {
	bz, err := ac.Codec.StringToBytes(text)
	if err != nil {
		return nil, err
	}
	if len(bz) != 20 && len(bz) != 32 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress,
			"address length must be 20 or 32 bz, got %d", len(bz))
	}
	return bz, nil
}

func NewAddressCodec() address.Codec {
	return addressCodec{
		Codec: addresscodec.NewBech32Codec(params.Bech32PrefixAccAddr),
	}
}

func NewValidatorAddressCodec() runtime.ValidatorAddressCodec {
	return addressCodec{
		Codec: addresscodec.NewBech32Codec(params.Bech32PrefixValAddr),
	}
}

func NewConsensusAddressCodec() runtime.ConsensusAddressCodec {
	return addressCodec{
		Codec: addresscodec.NewBech32Codec(params.Bech32PrefixConsAddr),
	}
}

func MustBytesToString(ac address.Codec, bz []byte) string {
	str, err := ac.BytesToString(bz)
	if err != nil {
		panic(err)
	}
	return str
}
