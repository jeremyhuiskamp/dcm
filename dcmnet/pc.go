package dcmnet

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
)

type PresentationContext struct {
	Id               uint8
	Result           uint32
	AbstractSyntax   string
	TransferSyntaxes []string
}

func (pc PresentationContext) Write(dst io.Writer) {
	// item type
	// TODO wrong! differentiate between request and response
	// TODO this should be a uint8 and a reserved 0x0
	binary.Write(dst, binary.LittleEndian, uint16(0x20))

	buf := new(bytes.Buffer)

	buf.WriteByte(pc.Id)
	buf.WriteByte(0)
	buf.WriteByte(0)
	buf.WriteByte(0)

	// TODO use constants here and below
	binary.Write(buf, binary.LittleEndian, uint16(0x30))
	binary.Write(buf, binary.BigEndian, uint16(len(pc.AbstractSyntax)))
	buf.WriteString(pc.AbstractSyntax)

	for _, transferSyntax := range pc.TransferSyntaxes {
		binary.Write(buf, binary.LittleEndian, uint16(0x40))
		binary.Write(buf, binary.BigEndian, uint16(len(transferSyntax)))
		buf.WriteString(transferSyntax)
	}

	binary.Write(dst, binary.BigEndian, uint16(buf.Len()))
	dst.Write(buf.Bytes())
}

func readString(src io.Reader) (string, error) {
	bytes, err := ioutil.ReadAll(src)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (pc *PresentationContext) Read(src io.Reader) error {
	var buf [4]byte
	_, err := src.Read(buf[:])
	if err != nil {
		return err
	}
	pc.Id = buf[0]
	log.Printf("Read presentation context id %d\n", pc.Id)

	return EachItem(src, func(item *Item) (err error) {
		switch item.Type {
		case AbstractSyntax:
			pc.AbstractSyntax, err = readString(item.Data)
			if err != nil {
				return err
			}
			log.Printf("Read pc abstract syntax: %s\n", pc.AbstractSyntax)
		case TransferSyntax:
			ts, err := readString(item.Data)
			if err != nil {
				return err
			}
			log.Printf("Read pc transfer syntax: %s\n", ts)
			pc.TransferSyntaxes = append(pc.TransferSyntaxes, ts)
		default:
			log.Printf("Unknown item type in presentation context: 0x%X\n",
				uint8(item.Type))
		}

		return nil
	})
}
