package distsig

import (
	"fmt"

	clientdss "git.vdb.to/cerc-io/chain-signatures/ethdss"
	"git.vdb.to/cerc-io/chain-signatures/ethschnorr"
)

type SigMessages struct {
	runID    SigRunID
	messages sigMessage
}

type sigStep int

const (
	sig_partial sigStep = iota
	sig_complete
)

type sigRun struct {
	*clientdss.DSS
	msgBuffer sigMessage

	dkgID DkgRunID
	step  sigStep

	sig ethschnorr.Signature
}

type sigMessage = *clientdss.PartialSig

func (d *sigRun) SigRunID() SigRunID {
	return SigRunID(fmt.Sprintf("%020d-%x", d.dkgID, d.SessionID()[:8]))
}

func (d *sigRun) prepareMessages() error {
	var err error
	d.msgBuffer, err = d.PartialSig()
	return err
}

func (d *sigRun) flushMessages() sigMessage {
	buf := d.msgBuffer
	d.msgBuffer = nil
	return buf
}

func (d *sigRun) processMessage(in sigMessage) error {
	err := d.ProcessPartialSig(in)
	if err != nil {
		return err
	}
	// return d.EnoughPartialSig(), nil
	if d.EnoughPartialSig() {
		d.sig, err = d.Signature()
		return err
	}
	return nil
}
