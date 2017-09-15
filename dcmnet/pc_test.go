package dcmnet

import (
	"reflect"
	"testing"

	"github.com/jeremyhuiskamp/dcm/dcm"
)

func TestRole(t *testing.T) {
	expectRole := func(r Role, scu, scp bool) {
		if r.IsSCU() != scu {
			t.Errorf("%s expected scu %t", r, scu)
		}
		if r.IsSCP() != scp {
			t.Errorf("%s expected scp %t", r, scp)
		}
	}

	expectRole(DefaultRole, true, false)

	var role Role
	expectRole(role, false, false)

	role.SetSCP(true)
	expectRole(role, false, true)

	role.SetSCU(true)
	expectRole(role, true, true)
}

// simple case with single, matching, accepted contexts
func simplePcs() PresentationContexts {
	return PresentationContexts{
		Requested: []PresentationContext{{
			AbstractSyntax: "1.2.3",
			ID:             1,
			TransferSyntaxes: []dcm.TransferSyntax{
				dcm.ImplicitVRLittleEndian,
			},
		}},
		Accepted: []PresentationContext{{
			AbstractSyntax: "1.2.3",
			ID:             1,
			Result:         PCAcceptance,
			TransferSyntaxes: []dcm.TransferSyntax{
				dcm.ImplicitVRLittleEndian,
			},
		}},
	}
}

// matches simplePcs above
var matchingTc = NewTransferCapability("1.2.3", dcm.ImplicitVRLittleEndian)

func expectMatchingTc(t *testing.T, got *TransferCapability) {
	if !reflect.DeepEqual(&matchingTc, got) {
		t.Errorf("unexpected transfer capability: %s", got)
	}
}

func TestPCsSingleAcceptance(t *testing.T) {
	pcs := simplePcs()
	if got := pcs.FindAcceptedTCap(PCID(1)); !reflect.DeepEqual(&matchingTc, got) {
		t.Errorf("unknown tc for pcid 1: %s", got)
	}
	expectMatchingTc(t, pcs.FindAcceptedTCap(PCID(1)))

	if got := pcs.FindAcceptedTCap(PCID(2)); got != nil {
		t.Errorf("expected no tc for pcid 2 but got: %s", got)
	}

	if got := pcs.FindAcceptedPCID(matchingTc); got == nil || PCID(1) != *got {
		t.Errorf("unexpected pcid: %d", got)
	}

	nonMatchingTc := matchingTc
	nonMatchingTc.AbstractSyntax = "1.2.3.4"
	if got := pcs.FindAcceptedPCID(nonMatchingTc); got != nil {
		t.Errorf("expected no pcid but got: %d", got)
	}
}

func TestPCsMultipleProposedTransferSyntaxes(t *testing.T) {
	pcs := simplePcs()
	pcs.Requested[0].TransferSyntaxes = []dcm.TransferSyntax{
		// add a new one first, make sure it gets skipped:
		dcm.ExplicitVRLittleEndian,
		pcs.Requested[0].TransferSyntaxes[0],
	}

	expectMatchingTc(t, pcs.FindAcceptedTCap(PCID(1)))
}

func TestPCsNoOverlappingTransferSyntaxes(t *testing.T) {
	pcs := simplePcs()
	// this probably isn't technically legal in dicom, but the code should
	// handle it:
	pcs.Accepted[0].TransferSyntaxes = []dcm.TransferSyntax{
		dcm.ExplicitVRBigEndian,
	}

	if got := pcs.FindAcceptedPCID(matchingTc); got != nil {
		t.Errorf("expected no pcid but got: %d", got)
	}
}

func TestPCsSingleUnacceptedContext(t *testing.T) {
	pcs := simplePcs()
	pcs.Accepted[0].Result = PCUserRejection

	if got := pcs.FindAcceptedTCap(PCID(1)); got != nil {
		t.Errorf("expected no tcap but got: %d", got)
	}
	if got := pcs.FindAcceptedPCID(matchingTc); got != nil {
		t.Errorf("expected no pcid but got: %d", got)
	}
}
