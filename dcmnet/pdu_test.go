package dcmnet

import (
	"bytes"
	"encoding/binary"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"testing"
)

func TestPDU(t *testing.T) {
	Suite(&PDUSuite{})
	Suite(&ItemSuite{})
	TestingT(t)
}

type PDUSuite struct{}

func (s *PDUSuite) TestReadOnePDU(c *C) {
	reader := pduDecoder(pdu(0x01, "hai!"))
	assertNextPDU(c, reader, 0x01, "hai!")
}

func (s *PDUSuite) TestReadTwoPDUs(c *C) {
	reader := pduDecoder(pdu(0x01, "one"), pdu(0x02, "two"))
	assertNextPDU(c, reader, 0x01, "one")
	assertNextPDU(c, reader, 0x02, "two")
}

func (s *PDUSuite) TestNilForEOF(c *C) {
	reader := pduDecoder()
	pdu, err := reader.NextPDU()
	c.Assert(pdu, IsNil)
	c.Assert(err, IsNil)
}

func (s *PDUSuite) TestDrainFirstPDUWhenAskedForSecond(c *C) {
	reader := pduDecoder(pdu(0x01, "one"), pdu(0x02, "two"))
	// not reading value...
	reader.NextPDU()
	assertNextPDU(c, reader, 0x02, "two")
}

type ItemSuite struct{}

func (s *ItemSuite) TestReadOneItem(c *C) {
	reader := itemReader(item(0x01, "hai!"))
	assertNextItem(c, reader, 0x01, "hai!")
}

func (s *ItemSuite) TestReadTwoItems(c *C) {
	reader := itemReader(item(0x01, "one"), item(0x02, "two"))
	assertNextItem(c, reader, 0x01, "one")
	assertNextItem(c, reader, 0x02, "two")
}

func (s *ItemSuite) TestNilForEOF(c *C) {
	reader := itemReader()
	item, err := reader.NextItem()
	c.Assert(item, IsNil)
	c.Assert(err, IsNil)
}

func (s *ItemSuite) TestDrainFirstItemWhenAskedForSecond(c *C) {
	reader := itemReader(item(0x01, "one"), item(0x02, "two"))
	// not reading value...
	reader.NextItem()
	assertNextItem(c, reader, 0x02, "two")
}

func pdu(pdutype uint8, data string) []byte {
	pdu := make([]byte, 6)
	pdu[0] = pdutype
	binary.BigEndian.PutUint32(pdu[2:6], uint32(len(data)))
	return append(pdu, []byte(data)...)
}

func concat(things ...[]byte) []byte {
	alldata := make([]byte, 0)
	for i := range things {
		alldata = append(alldata, things[i]...)
	}
	return alldata
}

func pduDecoder(pdus ...[]byte) PDUDecoder {
	return NewPDUDecoder(bytes.NewBuffer(concat(pdus...)))
}

func assertNextPDU(c *C, reader PDUDecoder, pduType PDUType, value string) {
	pdu, err := reader.NextPDU()

	c.Assert(err, IsNil)
	c.Assert(pdu, NotNil)

	c.Assert(pdu.Type, Equals, pduType)
	c.Assert(pdu.Length, Equals, uint32(len(value)))

	actualvalue, err := ioutil.ReadAll(pdu.Data)
	c.Assert(err, IsNil)
	c.Assert(string(actualvalue), Equals, value)
}

func item(itemtype uint8, data string) []byte {
	item := make([]byte, 4)
	item[0] = itemtype
	binary.BigEndian.PutUint16(item[2:4], uint16(len(data)))
	return append(item, []byte(data)...)
}

func itemReader(items ...[]byte) ItemReader {
	return NewItemReader(bytes.NewBuffer(concat(items...)))
}

func assertNextItem(c *C, reader ItemReader, itemType uint8, value string) {
	item, err := reader.NextItem()

	c.Assert(err, IsNil)
	c.Assert(item, NotNil)

	c.Assert(item.Type, Equals, ItemType(itemType))
	c.Assert(item.Length, Equals, uint16(len(value)))

	actualValue, err := ioutil.ReadAll(item.Data)
	c.Assert(err, IsNil)
	c.Assert(string(actualValue), Equals, value)
}
