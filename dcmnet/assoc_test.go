package dcmnet

import (
	"bytes"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"testing"
)

func TestAssoc(t *testing.T) {
	Suite(&AssocSuite{})
	TestingT(t)
}

type AssocSuite struct{}

func (s *AssocSuite) TestParseAssocRQCEcho(c *C) {
	assocrq := readAssocRQAC(c, "testdata/assocrq_cecho.bin", PDUAssociateRQ)

	c.Assert(assocrq.ProtocolVersion, Equals, int16(1))
	c.Assert(assocrq.CalledAE, Equals, "RIALTO")
	c.Assert(assocrq.CallingAE, Equals, "DCMECHO")
	c.Assert(assocrq.ApplicationContext, Equals, "1.2.840.10008.3.1.1.1")

	c.Assert(assocrq.PresentationContexts, HasLen, 1)
	pc1 := assocrq.PresentationContexts[0]
	c.Assert(pc1.Id, Equals, uint8(1))
	c.Assert(pc1.AbstractSyntax, Equals, "1.2.840.10008.1.1")
	c.Assert(pc1.Result, Equals, uint32(0)) // not sure if this matters in a req
	c.Assert(pc1.TransferSyntaxes, HasLen, 1)
	c.Assert(pc1.TransferSyntaxes[0], Equals, "1.2.840.10008.1.2")

	c.Assert(assocrq.MaxPDULength, Equals, uint32(16384))
	c.Assert(assocrq.ImplementationClassUID, Equals, "1.2.40.0.13.1.1")

	// hmm, not a good test, since these are the defaults:
	c.Assert(assocrq.MaxOperationsInvoked, Equals, uint16(0))
	c.Assert(assocrq.MaxOperationsPerformed, Equals, uint16(0))

	c.Assert(assocrq.ImplementationVersion, Equals, "dcm4che-2.0")
}

func (s *AssocSuite) TestParseAssocACCEcho(c *C) {
	assocac := readAssocRQAC(c, "testdata/assocac_cecho.bin", PDUAssociateAC)

	c.Assert(assocac.ProtocolVersion, Equals, int16(1))
	c.Assert(assocac.CalledAE, Equals, "RIALTO")
	c.Assert(assocac.CallingAE, Equals, "DCMECHO")
	c.Assert(assocac.ApplicationContext, Equals, "1.2.840.10008.3.1.1.1")

	c.Assert(assocac.PresentationContexts, HasLen, 1)
	pc1 := assocac.PresentationContexts[0]
	c.Assert(pc1.Id, Equals, uint8(1))
	c.Assert(pc1.AbstractSyntax, HasLen, 0)
	c.Assert(pc1.Result, Equals, uint32(0))
	c.Assert(pc1.TransferSyntaxes, HasLen, 1)
	c.Assert(pc1.TransferSyntaxes[0], Equals, "1.2.840.10008.1.2")

	c.Assert(assocac.MaxPDULength, Equals, uint32(16384))
	c.Assert(assocac.ImplementationClassUID, Equals, "1.2.40.0.13.1.1")

	// hmm, not a good test, since these are the defaults:
	c.Assert(assocac.MaxOperationsInvoked, Equals, uint16(0))
	c.Assert(assocac.MaxOperationsPerformed, Equals, uint16(0))

	c.Assert(assocac.ImplementationVersion, Equals, "dcm4che-2.0")
}

func readAssocRQAC(c *C, file string, pduType PDUType) AssociateRQAC {
	b, err := ioutil.ReadFile(file)
	c.Assert(err, IsNil)

	var buf bytes.Buffer
	buf.Write(b)
	buf.WriteString("don't read me bro")

	pduReader := NewPDUReader(&buf)
	pdu, err := pduReader.NextPDU()
	c.Assert(err, IsNil)
	c.Assert(pdu, NotNil)

	c.Assert(pdu.Type, Equals, pduType)
	var rqac AssociateRQAC
	rqac.Read(pdu.Data)

	// ensure that we stopped at the right spot:
	c.Assert(buf.String(), Equals, "don't read me bro")

	return rqac
}
