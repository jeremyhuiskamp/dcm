package dcmnet

import (
	"encoding/binary"
	"github.com/jeremyhuiskamp/dcm/stream"
	"io"
)

type ItemType uint8

//go:generate stringer -type ItemType
const (
	// consider breaking these out into the files that use them?
	// can stringer support that?
	ApplicationContext         ItemType = 0x10
	RequestPresentationContext ItemType = 0x20
	AcceptPresentationContext  ItemType = 0x21
	AbstractSyntax             ItemType = 0x30
	TransferSyntax             ItemType = 0x40
	UserInfo                   ItemType = 0x50
	MaxPDULength               ItemType = 0x51
	ImplementationClassUID     ItemType = 0x52
	AsyncOperations            ItemType = 0x53
	RoleSelection              ItemType = 0x54
	ImplementationVersion      ItemType = 0x55
)

type Item struct {
	Type   ItemType
	Length uint16
	Data   stream.Stream
}

// TODO: rename to ItemDecoder

type ItemReader struct {
	data StreamDecoder
}

func NewItemReader(data io.Reader) ItemReader {
	return ItemReader{StreamDecoder{data, nil}}
}

func (reader *ItemReader) NextItem() (item *Item, err error) {
	item = &Item{}
	item.Data, err = reader.data.NextChunk(4, func(header []byte) int64 {
		item.Type = ItemType(header[0])
		item.Length = binary.BigEndian.Uint16(header[2:4])
		return int64(item.Length)
	})

	if err != nil || item.Data == nil {
		return nil, err
	}

	return item, nil
}

// Poor man's iterator over ItemReader.NextItem()
func EachItem(src io.Reader, f func(*Item) error) error {
	items := NewItemReader(src)

	// this seems uglier than it should have to be...
	for item, err := items.NextItem(); item != nil; item, err = items.NextItem() {
		if err != nil {
			return err
		}

		err = f(item)

		if err != nil {
			return err
		}
	}

	return nil
}
