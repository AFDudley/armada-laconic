package app

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"git.vdb.to/cerc-io/laconicd/server/distsig"
)

type (
	// VoteExtensionHandler defines the vote extension handler for LaconicApp.
	VoteExtensionHandler struct {
		app *LaconicApp
	}

	// VoteExtension defines the structure used to create a vote extension.
	VoteExtension struct {
		Hash   []byte
		Height int64
		Data   []byte
	}
)

func NewVoteExtensionHandler() *VoteExtensionHandler {
	return &VoteExtensionHandler{}
}

func (app *LaconicApp) NewVoteExtensionHandler() *VoteExtensionHandler {
	return &VoteExtensionHandler{app: app}
}

func (h *VoteExtensionHandler) SetHandlers(bApp *baseapp.BaseApp) {
	if h.app != nil {
		// Use the real laconic handlers when app is available
		bApp.SetExtendVoteHandler(h.ExtendVoteWithLaconic())
		bApp.SetVerifyVoteExtensionHandler(h.VerifyVoteExtensionWithLaconic())
		bApp.SetPrepareProposal(h.PrepareProposal())
		bApp.SetProcessProposal(h.ProcessProposal())
	} else {
		// Use dummy handlers for testing
		bApp.SetExtendVoteHandler(h.ExtendVote())
		bApp.SetVerifyVoteExtensionHandler(h.VerifyVoteExtension())
	}
}

func (h *VoteExtensionHandler) ExtendVote() sdk.ExtendVoteHandler {
	return func(_ sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		buf := make([]byte, 1024)

		_, err := rand.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random vote extension data: %w", err)
		}

		ve := VoteExtension{
			Hash:   req.Hash,
			Height: req.Height,
			Data:   buf,
		}

		bz, err := json.Marshal(ve)
		if err != nil {
			return nil, fmt.Errorf("failed to encode vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		var ve VoteExtension

		if err := json.Unmarshal(req.VoteExtension, &ve); err != nil {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		switch {
		case req.Height != ve.Height:
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil

		case !bytes.Equal(req.Hash, ve.Hash):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil

		case len(ve.Data) != 1024:
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

// PrepareProposal is responsible for:
// - collecting Nitro state updates and applying them as inserted transactions in the block proposal
func (h *VoteExtensionHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {

		// paramsResp, err := h.app.ConsensusParamsKeeper.Params(ctx, &consensustypes.QueryParamsRequest{})
		// if err != nil {
		// 	return nil, err
		// }

		// var maxBlockGas uint64
		// if b := paramsResp.GetParams().Block; b != nil {
		// 	maxBlockGas = uint64(b.MaxGas)
		// }

		// Decode transactions from bytes
		txs := make([]sdk.Tx, 0, len(req.Txs))
		for _, txBz := range req.Txs {
			tx, err := h.app.txConfig.TxDecoder()(txBz)
			if err != nil {
				continue // Skip invalid transactions
			}
			txs = append(txs, tx)
		}

		// For now, just return the transactions as-is
		// TODO: Add custom transaction selection logic here
		return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
	}
}

func (h *VoteExtensionHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		// For now, accept all proposals
		// TODO: Add custom proposal validation logic here
		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// ExtendVoteWithLaconic implements the real Laconic-specific vote extension logic
// - renews distsig key (DKG) if validator set has changed
// - broadcasts prepared DKG and distsig messages
func (h *VoteExtensionHandler) ExtendVoteWithLaconic() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		// Check if we have a distributed signature manager
		// Since we can't access DistSigManager() directly in v1, we'll implement a simpler version
		// TODO: Integrate with actual distributed signature manager when available

		// Check if we are a participant
		if h.app.OnboardingKeeper != nil {
			// For now, return empty vote extension
			// TODO: Implement the full distsig logic here
			return &abci.ResponseExtendVote{VoteExtension: []byte{}}, nil
		}

		return &abci.ResponseExtendVote{VoteExtension: []byte{}}, nil
	}
}

// VerifyVoteExtensionWithLaconic implements the real Laconic-specific vote extension verification
func (h *VoteExtensionHandler) VerifyVoteExtensionWithLaconic() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		if len(req.VoteExtension) == 0 {
			// Empty vote extension is acceptable
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
		}

		// Decode and verify the vote extension
		dec := gob.NewDecoder(bytes.NewReader(req.VoteExtension))
		var messages distsig.PeerMessages
		if err := dec.Decode(&messages); err != nil {
			// If we can't decode, reject
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		// TODO: Actually verify the messages with the distributed signature manager
		// For now, just accept
		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
