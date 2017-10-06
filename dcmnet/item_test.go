package dcmnet

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"testing"
)

func TestReadOneItem(t *testing.T) {
	reader := itemReader(item(0x01, "hai!"))
	expectNextItem(t, reader, 0x01, "hai!")
}

func TestReadTwoItems(t *testing.T) {
	reader := itemReader(item(0x01, "one"), item(0x02, "two"))
	expectNextItem(t, reader, 0x01, "one")
	expectNextItem(t, reader, 0x02, "two")
}

func TestItemNilForEOF(t *testing.T) {
	reader := itemReader()
	item, err := reader.NextItem()
	if item != nil {
		t.Error("expected no item")
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestDrainFirstItemWhenAskedForSecond(t *testing.T) {
	reader := itemReader(item(0x01, "one"), item(0x02, "two"))
	// not reading value...
	reader.NextItem()
	expectNextItem(t, reader, 0x02, "two")
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

func expectNextItem(t *testing.T, reader ItemReader, itemType ItemType, value string) {
	item, err := reader.NextItem()
	if err != nil {
		t.Fatal(err)
	}
	if item == nil {
		t.Fatal("no item found")
	}
	if itemType != item.Type {
		t.Error("expected type %s, got %s", itemType, item.Type)
	}
	if len(value) != int(item.Length) {
		t.Error("expected length %d, got %d", len(value), item.Length)
	}
	actualValue, err := ioutil.ReadAll(item.Data)
	if err != nil {
		t.Fatal(err)
	}
	if value != string(actualValue) {
		t.Error("expected value %q, got %q", value, string(actualValue))
	}
}
