package distsig

import (
	"fmt"

	dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
)

// DkgMessages collects local DKG messages that are pending broadcast.
type DkgMessages struct {
	runID    DkgRunID
	messages []dkgMessage
}

// DKG steps (actions) that are executed by the DKG protocol
type dkgStep int

const (
	dkg_deal dkgStep = iota
	dkg_certify
	dkg_commit
	dkg_finish
	dkg_done
)

type dkgRun struct {
	*dkg.DistKeyGenerator
	// runID RunID
	msgBuffer []dkgMessage
	step      dkgStep
	share     *dkg.DistKeyShare
}

type dkgMessage interface {
	send(*dkgRun) (dkgMessage, error)
	from() uint32
	target() uint32
}

// prepareMessages produces DKG state transitions that don't rely on peer inputs, i.e. the deal or
// commit steps.  This is called during ExtendVote by all validators.
func (d *dkgRun) prepareMessages() error {
	out, err := d.handleMessage(nil)
	if err != nil {
		return err
	}
	if out != nil {
		d.msgBuffer = append(d.msgBuffer, out...)
	}
	return nil
}

func (d *dkgRun) flushMessages() []dkgMessage {
	buf := d.msgBuffer
	d.msgBuffer = nil
	return buf
}

// processMessages processes incoming DKG messages and produces outgoing messages.  This is called
// during VerifyVoteExtension by all validators (TODO: verify).
func (d *dkgRun) processMessages(in []dkgMessage) error {
	for _, msg := range in {
		out, err := d.handleMessage(msg)
		if err != nil {
			return err
		}
		if out != nil {
			d.msgBuffer = append(d.msgBuffer, out...)
		}
	}
	return nil
}

func (d *dkgRun) handleMessage(in dkgMessage) ([]dkgMessage, error) {
	// fmt.Printf("[%d, %7s] handle: %T", d.index, d.step, in)
	// if in != nil {
	// 	fmt.Printf(" (from %d to %+v)", in.from(), in.target())
	// }
	// fmt.Println()

	// nil input means we try to execute a step requiring no peer input
	if in == nil {
		var out []dkgMessage
		switch d.step {
		case dkg_deal:
			deals, err := d.Deals()
			if err != nil {
				return nil, err
			}
			for index, deal := range deals {
				out = append(out, msgDeal{index: uint32(index), msg: deal})
			}
		case dkg_commit:
			sc, err := d.SecretCommits()
			if err != nil {
				return nil, err
			}
			out = []dkgMessage{msgSecretCommits{msg: sc}}
		case dkg_done:
			dks, err := d.DistKeyShare()
			if err != nil {
				return nil, err
			}
			d.share = dks
		default:
			return nil, fmt.Errorf("unexpected stage: %v", d.step)
		}
		d.step = nextStep(d.step)
		return out, nil
	}

	var out []dkgMessage
	if msg, err := in.send(d); err != nil {
		return nil, err
	} else if msg != nil {
		out = []dkgMessage{msg}
	}
	if (d.step == dkg_certify && d.Certified()) ||
		(d.step == dkg_finish && d.Finished()) {
		d.step = nextStep(d.step)
		msgs, err := d.handleMessage(nil)
		if err != nil {
			return nil, err
		}
		out = append(out, msgs...)
	}
	return out, nil
}

func nextStep(step dkgStep) dkgStep {
	if step == dkg_done {
		return dkg_done
	}
	return step + 1
}

type (
	msgDeal struct {
		msg   *dkg.Deal
		index uint32 // recipient index
	}
	msgResponse           struct{ msg *dkg.Response }
	msgJustification      struct{ msg *dkg.Justification }
	msgSecretCommits      struct{ msg *dkg.SecretCommits }
	msgComplaintCommits   struct{ msg *dkg.ComplaintCommits }
	msgReconstructCommits struct{ msg *dkg.ReconstructCommits }
)

func (m msgDeal) send(d *dkgRun) (dkgMessage, error) {
	if m.index != uint32(d.Index()) {
		return nil, nil
	}
	out, err := d.ProcessDeal(m.msg)
	if err != nil {
		return nil, err
	}
	return msgResponse{out}, nil
}

func (m msgResponse) send(d *dkgRun) (dkgMessage, error) {
	if out, err := d.ProcessResponse(m.msg); err != nil {
		return nil, err
	} else if out != nil {
		return msgJustification{out}, nil
	}
	return nil, nil
}

func (m msgJustification) send(d *dkgRun) (dkgMessage, error) {
	return nil, d.ProcessJustification(m.msg)
}

func (m msgSecretCommits) send(d *dkgRun) (dkgMessage, error) {
	if out, err := d.ProcessSecretCommits(m.msg); err != nil {
		return nil, err
	} else if out != nil {
		return msgComplaintCommits{out}, nil
	}
	return nil, nil
}

func (m msgComplaintCommits) send(d *dkgRun) (dkgMessage, error) {
	out, err := d.ProcessComplaintCommits(m.msg)
	if err != nil {
		return nil, err
	}
	return msgReconstructCommits{out}, nil
}

func (m msgReconstructCommits) send(d *dkgRun) (dkgMessage, error) {
	return nil, d.ProcessReconstructCommits(m.msg)
}

func (d dkgStep) String() string {
	switch d {
	case dkg_deal:
		return "deal"
	case dkg_certify:
		return "certify"
	case dkg_commit:
		return "commit"
	case dkg_finish:
		return "finish"
	case dkg_done:
		return "done"
	default:
		return "unknown"
	}
}

// for debugging
func (m msgDeal) from() uint32               { return m.msg.Index }
func (m msgResponse) from() uint32           { return m.msg.Response.Index }
func (m msgJustification) from() uint32      { return m.msg.Justification.Index }
func (m msgSecretCommits) from() uint32      { return m.msg.Index }
func (m msgComplaintCommits) from() uint32   { return m.msg.Index }
func (m msgReconstructCommits) from() uint32 { return m.msg.Index }

func (m msgDeal) target() uint32               { return m.index }
func (m msgResponse) target() uint32           { return m.msg.Index }
func (m msgJustification) target() uint32      { return m.msg.Index }
func (m msgSecretCommits) target() uint32      { return 0 }
func (m msgComplaintCommits) target() uint32   { return m.msg.DealerIndex }
func (m msgReconstructCommits) target() uint32 { return m.msg.DealerIndex }
