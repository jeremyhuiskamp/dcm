package dcmnet

import (
	"bytes"
	"github.com/jeremyhuiskamp/dcm/dcm"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func TestParseAssocRQCEcho(t *testing.T) {
	RegisterTestingT(t)

	assocrq := readAssocRQAC("testdata/assocrq_cecho.bin", PDUAssociateRQ)

	Expect(assocrq.ProtocolVersion).To(Equal(int16(1)))
	Expect(assocrq.CalledAE).To(Equal("RIALTO"))
	Expect(assocrq.CallingAE).To(Equal("DCMECHO"))
	Expect(assocrq.ApplicationContext).To(Equal("1.2.840.10008.3.1.1.1"))

	Expect(assocrq.PresentationContexts).To(HaveLen(1))
	pc1 := assocrq.PresentationContexts[0]
	Expect(pc1.Id).To(Equal(PCID(1)))
	Expect(pc1.AbstractSyntax).To(Equal("1.2.840.10008.1.1"))
	Expect(pc1.Result).To(Equal(PCAcceptance)) // should always be 0 in a request
	Expect(pc1.TransferSyntaxes).To(HaveLen(1))
	Expect(pc1.TransferSyntaxes[0]).To(Equal(dcm.ImplicitVRLittleEndian))

	Expect(assocrq.MaxPDULength).To(Equal(uint32(16384)))
	Expect(assocrq.ImplementationClassUID).To(Equal("1.2.40.0.13.1.1"))

	// hmm, not a good test, since these are the defaults:
	Expect(assocrq.MaxOperationsInvoked).To(Equal(uint16(0)))
	Expect(assocrq.MaxOperationsPerformed).To(Equal(uint16(0)))

	Expect(assocrq.ImplementationVersion).To(Equal("dcm4che-2.0"))
}

func TestParseAssocACCEcho(t *testing.T) {
	RegisterTestingT(t)

	assocac := readAssocRQAC("testdata/assocac_cecho.bin", PDUAssociateAC)

	Expect(assocac.ProtocolVersion).To(Equal(int16(1)))
	Expect(assocac.CalledAE).To(Equal("RIALTO"))
	Expect(assocac.CallingAE).To(Equal("DCMECHO"))
	Expect(assocac.ApplicationContext).To(Equal("1.2.840.10008.3.1.1.1"))

	Expect(assocac.PresentationContexts).To(HaveLen(1))
	pc1 := assocac.PresentationContexts[0]
	Expect(pc1.Id).To(Equal(PCID(1)))
	Expect(pc1.AbstractSyntax).To(HaveLen(0))
	Expect(pc1.Result).To(Equal(PCAcceptance))
	Expect(pc1.TransferSyntaxes).To(HaveLen(1))
	Expect(pc1.TransferSyntaxes[0]).To(Equal(dcm.ImplicitVRLittleEndian))

	Expect(assocac.MaxPDULength).To(Equal(uint32(16384)))
	Expect(assocac.ImplementationClassUID).To(Equal("1.2.40.0.13.1.1"))

	// hmm, not a good test, since these are the defaults:
	Expect(assocac.MaxOperationsInvoked).To(Equal(uint16(0)))
	Expect(assocac.MaxOperationsPerformed).To(Equal(uint16(0)))

	Expect(assocac.ImplementationVersion).To(Equal("dcm4che-2.0"))
}

func readAssocRQAC(file string, pduType PDUType) AssociateRQAC {
	b, err := ioutil.ReadFile(file)
	Expect(err).To(BeNil())

	var buf bytes.Buffer
	buf.Write(b)
	buf.WriteString("don't read me bro")

	pduDecoder := NewPDUDecoder(&buf)
	pdu, err := pduDecoder.NextPDU()
	Expect(err).To(BeNil())
	Expect(pdu).ToNot(BeNil())

	Expect(pdu.Type).To(Equal(pduType))
	var rqac AssociateRQAC
	rqac.Read(pdu.Data)

	// ensure that we stopped at the right spot:
	Expect(buf.String()).To(Equal("don't read me bro"))

	return rqac
}
