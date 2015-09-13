package dcmnet

import (
	"encoding/binary"
	"io"
	"io/ioutil"
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
	ImplementationVersion      ItemType = 0x55
)

type Item struct {
	Type   ItemType
	Length uint16
	Data   io.Reader
}

type ItemReader struct {
	data     io.Reader
	lastItem *Item
}

func NewItemReader(data io.Reader) ItemReader {
	return ItemReader{data: data}
}

func (reader *ItemReader) NextItem() (*Item, error) {
	if reader.lastItem != nil {
		_, err := io.Copy(ioutil.Discard, reader.lastItem.Data)
		if err != nil {
			return nil, err
		}
	}

	header := make([]byte, 4)
	headerlen, err := io.ReadFull(reader.data, header)
	if headerlen == 0 {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	reader.lastItem = &Item{
		Type:   ItemType(header[0]),
		Length: binary.BigEndian.Uint16(header[2:4]),
	}

	// TODO: sanity check the length?
	reader.lastItem.Data = io.LimitReader(reader.data, int64(reader.lastItem.Length))

	return reader.lastItem, nil
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

		io.Copy(ioutil.Discard, item.Data)
	}

	return nil
}
