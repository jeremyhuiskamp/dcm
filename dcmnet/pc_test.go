package dcmnet

import (
	"testing"

	"github.com/jeremyhuiskamp/dcm/dcm"
	. "github.com/onsi/gomega"
)

func TestRole(t *testing.T) {
	RegisterTestingT(t)

	Expect(DefaultRole.IsSCU()).To(BeTrue())
	Expect(DefaultRole.IsSCP()).To(BeFalse())

	var role Role

	Expect(role.IsSCU()).To(BeFalse())
	Expect(role.IsSCP()).To(BeFalse())

	role.SetSCP(true)
	Expect(role.IsSCU()).To(BeFalse())
	Expect(role.IsSCP()).To(BeTrue())

	role.SetSCU(true)
	Expect(role.IsSCU()).To(BeTrue())
	Expect(role.IsSCP()).To(BeTrue())
}

// simple case with single, matching, accepted contexts
var simplePcs = PresentationContexts{
	Requested: []PresentationContext{
		{
			AbstractSyntax: "1.2.3",
			Id:             1,
			TransferSyntaxes: []dcm.TransferSyntax{
				dcm.ImplicitVRLittleEndian,
			},
		},
	},
	Accepted: []PresentationContext{
		{
			AbstractSyntax: "1.2.3",
			Id:             1,
			Result:         PCAcceptance,
			TransferSyntaxes: []dcm.TransferSyntax{
				dcm.ImplicitVRLittleEndian,
			},
		},
	},
}

// matches simplePcs above
var matchingTc = NewTransferCapability("1.2.3", dcm.ImplicitVRLittleEndian)

func TestPCsSingleAcceptance(t *testing.T) {
	RegisterTestingT(t)

	Expect(simplePcs.FindAcceptedTCap(PCID(1))).To(Equal(&matchingTc))

	Expect(simplePcs.FindAcceptedTCap(PCID(2))).To(BeNil())

	pcid1 := PCID(1)
	Expect(simplePcs.FindAcceptedPCID(matchingTc)).To(Equal(&pcid1))

	nonMatchingTc := matchingTc
	nonMatchingTc.AbstractSyntax = "1.2.3.4"
	Expect(simplePcs.FindAcceptedPCID(nonMatchingTc)).To(BeNil())
}

func TestPCsMultipleProposedTransferSyntaxes(t *testing.T) {
	RegisterTestingT(t)

	pcs := simplePcs
	pcs.Requested[0].TransferSyntaxes = []dcm.TransferSyntax{
		// add a new one first, make sure it gets skipped:
		dcm.ExplicitVRLittleEndian,
		pcs.Requested[0].TransferSyntaxes[0],
	}

	Expect(pcs.FindAcceptedTCap(PCID(1))).To(Equal(&matchingTc))
}

func TestPCsNoOverlappingTransferSyntaxes(t *testing.T) {
	RegisterTestingT(t)

	pcs := simplePcs
	// this probably isn't technically legal in dicom, but the code should
	// handle it:
	pcs.Accepted[0].TransferSyntaxes = []dcm.TransferSyntax{
		dcm.ExplicitVRBigEndian,
	}

	Expect(pcs.FindAcceptedPCID(matchingTc)).To(BeNil())
}

func TestPCsSingleUnacceptedContext(t *testing.T) {
	RegisterTestingT(t)

	pcs := simplePcs
	pcs.Accepted[0].Result = PCUserRejection

	Expect(pcs.FindAcceptedTCap(PCID(1))).To(BeNil())
	Expect(pcs.FindAcceptedPCID(matchingTc)).To(BeNil())
}
