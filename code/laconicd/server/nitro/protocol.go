package nitro

import (
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/statechannels/go-nitro/channel/state/outcome"
	"github.com/statechannels/go-nitro/protocols"
	nitrotypes "github.com/statechannels/go-nitro/types"

	"git.vdb.to/cerc-io/laconicd/x/nitro/types/v1"
)

type MsgWrapper struct {
	Msg         sdk.Msg
	Objective   protocols.ObjectiveId
	SignerField *string
}

func (s *Server) OpenLedgerChannel(funds nitrotypes.Funds, counterparty nitrotypes.Participant) (*MsgWrapper, error) {
	// TODO: handle multiple assets
	if len(funds) != 1 {
		return nil, errors.New("only one asset is supported")
	}
	fund, exit := makeOutcome(s.ethAddress, counterparty, funds)

	r, err := s.node.CreateLedgerChannel(counterparty, 0, exit)
	if err != nil {
		return nil, err
	}
	s.logger.Info("Initiated objective", "id", r.Id)

	// Convert nitrotypes.Funds back to sdk.Coins for the message
	// Use "token" as a default denomination for now
	// TODO: nitro will store a mapping of denoms to token addresses
	// an entry will be added for new denominations upon confirmation of ledger deposit
	coins := sdk.Coins{sdk.NewCoin("token", math.NewIntFromBigInt(fund.Amount))}

	msg := &v1.MsgOpenChannel{
		Funds:        coins,
		NitroAddress: s.ethAddress.String(),
		Counterparty: counterparty.String(),
		ChannelId:    r.ChannelId.String(),
	}
	return &MsgWrapper{msg, r.Id, &msg.Signer}, nil
}

func (s *Server) CloseLedgerChannel(channelId nitrotypes.ChannelID, isChallenge bool) (*MsgWrapper, error) {
	objective, err := s.node.CloseLedgerChannel(channelId, isChallenge)
	if err != nil {
		return nil, err
	}
	s.logger.Info("Initiated objective", "id", objective)

	msg := &v1.MsgCloseChannel{
		ChannelId:   channelId.String(),
		IsChallenge: isChallenge,
	}
	return &MsgWrapper{msg, objective, &msg.Signer}, nil
}

func (s *Server) CreatePaymentChannel(intermediaries []string, counterpartyString string, challengeDuration uint32, funds nitrotypes.Funds) (*MsgWrapper, error) {
	// Parse intermediaries
	var intermediaryParticipants []nitrotypes.Participant
	for _, addr := range intermediaries {
		participant, err := nitrotypes.ParseParticipant(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse intermediary %s: %w", addr, err)
		}
		intermediaryParticipants = append(intermediaryParticipants, participant)
	}

	// Parse counterparty
	counterparty, err := nitrotypes.ParseParticipant(counterpartyString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse counterparty: %w", err)
	}

	// TODO: handle multiple assets
	if len(funds) != 1 {
		return nil, errors.New("only one asset is supported")
	}
	if len(funds) != 1 {
		return nil, errors.New("only one asset is supported")
	}
	fund, exit := makeOutcome(s.ethAddress, counterparty, funds)

	// Create payment channel
	r, err := s.node.CreatePaymentChannel(intermediaryParticipants, counterparty, challengeDuration, exit)
	if err != nil {
		return nil, err
	}
	s.logger.Info("Initiated payment channel objective", "id", r.Id)

	// Convert nitrotypes.Funds back to sdk.Coins for the message
	coins := sdk.Coins{sdk.NewCoin("token", math.NewIntFromBigInt(fund.Amount))}

	msg := &v1.MsgCreatePaymentChannel{
		NitroAddress:      s.ethAddress.String(),
		Counterparty:      counterpartyString,
		Intermediaries:    intermediaries,
		ChallengeDuration: challengeDuration,
		Funds:             coins,
		ChannelId:         r.ChannelId.String(),
	}
	return &MsgWrapper{msg, r.Id, &msg.Signer}, nil
}

func (s *Server) ClosePaymentChannel(channelId nitrotypes.ChannelID) (*MsgWrapper, error) {
	objective, err := s.node.ClosePaymentChannel(channelId)
	if err != nil {
		return nil, err
	}
	s.logger.Info("Initiated payment channel close objective", "id", objective)

	msg := &v1.MsgClosePaymentChannel{
		ChannelId: channelId.String(),
	}
	return &MsgWrapper{msg, objective, &msg.Signer}, nil
}

func (s *Server) Pay(channelId nitrotypes.ChannelID, amount sdk.Coin) (*MsgWrapper, error) {
	err := s.node.Pay(channelId, amount.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	nodeAddrStr, err := s.addressCodec.BytesToString(s.nodeAddress())
	if err != nil {
		return nil, err
	}

	msg := &v1.MsgPay{
		ChannelId: channelId.String(),
		Amount:    amount,
		Signer:    nodeAddrStr,
	}
	return &MsgWrapper{msg, "", &msg.Signer}, nil
}

func makeOutcome(ethAddress, counterparty nitrotypes.Participant, funds nitrotypes.Funds) (nitrotypes.Fund, outcome.Exit) {
	var fund nitrotypes.Fund
	for asset, amount := range funds {
		fund = nitrotypes.Fund{asset, amount}
	}

	exit := outcome.Exit{
		outcome.SingleAssetExit{
			Asset: fund.Asset,
			Allocations: []outcome.Allocation{
				{
					Destination: ethAddress.ToDestination(),
					Amount:      fund.Amount,
				},
				{
					Destination: counterparty.ToDestination(),
					Amount:      big.NewInt(0),
				},
			},
		},
	}
	return fund, exit
}
