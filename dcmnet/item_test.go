package dcmnet

import (
	"bytes"
	"encoding/binary"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"testing"
)

func TestItem(t *testing.T) {
	Suite(&ItemSuite{})
	TestingT(t)
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

func item(itemtype uint8, data string) (buf bytes.Buffer) {
	item := make([]byte, 4)
	item[0] = itemtype
	binary.BigEndian.PutUint16(item[2:4], uint16(len(data)))
	buf.Write(item)
	buf.Write([]byte(data))
	return buf
}

func itemReader(items ...interface{}) ItemReader {
	b := bufcat(items...)
	return NewItemReader(&b)
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
