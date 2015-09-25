package dcmnet

import (
	. "gopkg.in/check.v1"
	"io/ioutil"
	"testing"
)

func TestPDU(t *testing.T) {
	Suite(&PDUSuite{})
	TestingT(t)
}

type PDUSuite struct{}

func (s *PDUSuite) TestReadOnePDU(c *C) {
	decoder := pduDecoder(bufpdu(0x01, "hai!"))
	assertNextPDU(c, decoder, 0x01, "hai!")
}

func (s *PDUSuite) TestReadTwoPDUs(c *C) {
	decoder := pduDecoder(bufpdu(0x01, "one"), bufpdu(0x02, "two"))
	assertNextPDU(c, decoder, 0x01, "one")
	assertNextPDU(c, decoder, 0x02, "two")
}

func (s *PDUSuite) TestNilForEOF(c *C) {
	decoder := pduDecoder()
	pdu, err := decoder.NextPDU()
	c.Assert(pdu, IsNil)
	c.Assert(err, IsNil)
}

func (s *PDUSuite) TestDrainFirstPDUWhenAskedForSecond(c *C) {
	decoder := pduDecoder(bufpdu(0x01, "one"), bufpdu(0x02, "two"))
	// not reading value...
	decoder.NextPDU()
	assertNextPDU(c, decoder, 0x02, "two")
}

func pduDecoder(pdus ...interface{}) PDUDecoder {
	b := bufcat(pdus...)
	return NewPDUDecoder(&b)
}

func assertNextPDU(c *C, decoder PDUDecoder, pduType PDUType, value string) {
	pdu, err := decoder.NextPDU()

	c.Assert(err, IsNil)
	c.Assert(pdu, NotNil)

	c.Assert(pdu.Type, Equals, pduType)
	c.Assert(pdu.Length, Equals, uint32(len(value)))

	actualvalue, err := ioutil.ReadAll(pdu.Data)
	c.Assert(err, IsNil)
	c.Assert(string(actualvalue), Equals, value)
}
