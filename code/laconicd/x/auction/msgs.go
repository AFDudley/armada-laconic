package auction

import (
	"fmt"
	time "time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"git.vdb.to/cerc-io/laconicd/utils"
)

var (
	_ sdk.Msg = &MsgCreateAuction{}
	_ sdk.Msg = &MsgCommitBid{}
	_ sdk.Msg = &MsgRevealBid{}
)

// NewMsgCreateAuction is the constructor function for MsgCreateAuction.
func NewMsgCreateAuction(
	kind string,
	commitsDuration time.Duration,
	revealsDuration time.Duration,
	commitFee sdk.Coin,
	revealFee sdk.Coin,
	minimumBid sdk.Coin,
	maxPrice sdk.Coin,
	numProviders int32,
	signer sdk.AccAddress,
) MsgCreateAuction {
	signerStr, err := utils.NewAddressCodec().BytesToString(signer)
	if err != nil {
		panic(err)
	}
	return MsgCreateAuction{
		CommitsDuration: commitsDuration,
		RevealsDuration: revealsDuration,
		CommitFee:       commitFee,
		RevealFee:       revealFee,
		MinimumBid:      minimumBid,
		MaxPrice:        maxPrice,
		Kind:            kind,
		NumProviders:    numProviders,
		Signer:          signerStr,
	}
}

// NewMsgCommitBid is the constructor function for MsgCommitBid.
func NewMsgCommitBid(auctionId string, commitHash string, signer sdk.AccAddress) MsgCommitBid {
	signerStr, err := utils.NewAddressCodec().BytesToString(signer)
	if err != nil {
		panic(err)
	}
	return MsgCommitBid{
		AuctionId:  auctionId,
		CommitHash: commitHash,
		Signer:     signerStr,
	}
}

// NewMsgRevealBid is the constructor function for MsgRevealBid.
func NewMsgRevealBid(auctionId string, reveal string, signer sdk.AccAddress) MsgRevealBid {
	signerStr, err := utils.NewAddressCodec().BytesToString(signer)
	if err != nil {
		panic(err)
	}
	return MsgRevealBid{
		AuctionId: auctionId,
		Reveal:    reveal,
		Signer:    signerStr,
	}
}

// NewMsgReleaseFunds is the constructor function for MsgReleaseFunds.
func NewMsgReleaseFunds(auctionId string, signer sdk.AccAddress) MsgReleaseFunds {
	signerStr, err := utils.NewAddressCodec().BytesToString(signer)
	if err != nil {
		panic(err)
	}
	return MsgReleaseFunds{
		AuctionId: auctionId,
		Signer:    signerStr,
	}
}

// ValidateBasic Implements Msg.
func (msg MsgCreateAuction) ValidateBasic() error {
	if msg.Signer == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}

	if msg.Kind != AuctionKindVickrey && msg.Kind != AuctionKindProvider {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("auction kind should be one of %s | %s", AuctionKindVickrey, AuctionKindProvider))
	}

	if msg.CommitsDuration <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commit phase duration invalid")
	}

	if msg.RevealsDuration <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "reveal phase duration invalid")
	}

	if msg.Kind == AuctionKindVickrey && !msg.MinimumBid.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("minimum bid should be greater than zero for %s auction", AuctionKindVickrey))
	}

	if msg.Kind == AuctionKindProvider {
		if !msg.MaxPrice.IsPositive() {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("max price should be greater than zero for %s auction", AuctionKindProvider))
		}

		if msg.NumProviders <= 0 {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("num providers should be greater than zero for %s auction", AuctionKindProvider))
		}
	}

	return nil
}

// ValidateBasic Implements Msg.
func (msg MsgCommitBid) ValidateBasic() error {
	if msg.Signer == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid signer address")
	}

	if msg.AuctionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid auction id")
	}

	if msg.CommitHash == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid commit hash")
	}

	return nil
}

// ValidateBasic Implements Msg.
func (msg MsgRevealBid) ValidateBasic() error {
	if msg.Signer == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid signer address")
	}

	if msg.AuctionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid auction id")
	}

	if msg.Reveal == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid reveal data")
	}

	return nil
}

// ValidateBasic Implements Msg.
func (msg MsgReleaseFunds) ValidateBasic() error {
	if msg.Signer == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid signer address")
	}

	if msg.AuctionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid auction id")
	}

	return nil
}
