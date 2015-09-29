package dcmnet

import (
	"bytes"
	"encoding/binary"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func TestReadOneItem(t *testing.T) {
	RegisterTestingT(t)

	reader := itemReader(item(0x01, "hai!"))
	expectNextItem(reader, 0x01, "hai!")
}

func TestReadTwoItems(t *testing.T) {
	RegisterTestingT(t)

	reader := itemReader(item(0x01, "one"), item(0x02, "two"))
	expectNextItem(reader, 0x01, "one")
	expectNextItem(reader, 0x02, "two")
}

func TestItemNilForEOF(t *testing.T) {
	RegisterTestingT(t)

	reader := itemReader()
	item, err := reader.NextItem()
	Expect(item).To(BeNil())
	Expect(err).ToNot(HaveOccurred())
}

func TestDrainFirstItemWhenAskedForSecond(t *testing.T) {
	RegisterTestingT(t)

	reader := itemReader(item(0x01, "one"), item(0x02, "two"))
	// not reading value...
	reader.NextItem()
	expectNextItem(reader, 0x02, "two")
}

func item(itemtype ItemType, data string) (buf bytes.Buffer) {
	item := make([]byte, 4)
	item[0] = byte(itemtype)
	binary.BigEndian.PutUint16(item[2:4], uint16(len(data)))
	buf.Write(item)
	buf.Write([]byte(data))
	return buf
}

func itemReader(items ...interface{}) ItemReader {
	b := bufcat(items...)
	return NewItemReader(&b)
}

func expectNextItem(reader ItemReader, itemType ItemType, value string) {
	item, err := reader.NextItem()

	Expect(err).ToNot(HaveOccurred())
	Expect(item).ToNot(BeNil())

	Expect(item.Type).To(Equal(itemType))
	Expect(item.Length).To(Equal(uint16(len(value))))

	actualvalue, err := ioutil.ReadAll(item.Data)
	Expect(err).ToNot(HaveOccurred())
	Expect(string(actualvalue)).To(Equal(value))
}
