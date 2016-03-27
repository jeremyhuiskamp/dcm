package dcmnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jeremyhuiskamp/dcm/stream"
	"io"
)

type PDUType uint8

//go:generate stringer -type PDUType
const (
	PDUAssociateRQ      PDUType = 0x01
	PDUAssociateAC      PDUType = 0x02
	PDUAssociateRJ      PDUType = 0x03
	PDUPresentationData PDUType = 0x04
	PDUReleaseRQ        PDUType = 0x05
	PDUReleaseRP        PDUType = 0x06
	PDUAbort            PDUType = 0x07
)

// Protocol Data Unit
type PDU struct {
	Type   PDUType
	Length uint32
	Data   stream.Stream
}

func (p PDU) String() string {
	return fmt.Sprintf("[%s, len=%d]", p.Type, p.Length)
}

// PDUDecoder parses a stream for PDUs
type PDUDecoder struct {
	data StreamDecoder
}

func NewPDUDecoder(data io.Reader) PDUDecoder {
	return PDUDecoder{StreamDecoder{data, nil}}
}

// Read the next PDU from the stream
func (d *PDUDecoder) NextPDU() (pdu *PDU, err error) {
	pdu = &PDU{}
	pdu.Data, err = d.data.NextChunk(6, func(header []byte) int64 {
		pdu.Type = PDUType(header[0])
		pdu.Length = binary.BigEndian.Uint32(header[2:6])
		return int64(pdu.Length)
	})

	// either error, or no more pdus
	if err != nil || pdu.Data == nil {
		return nil, err
	}

	return pdu, err
}

type PDUEncoder struct {
	out    io.Writer
	header [6]byte
}

func NewPDUEncoder(out io.Writer) PDUEncoder {
	return PDUEncoder{out: out}
}

func (w *PDUEncoder) NextPDU(pdu PDU) (err error) {
	w.header[0] = byte(pdu.Type)
	w.header[1] = 0
	// here, we could consider ignoring the passed-in length, but instead
	// writing the data to a buffer and calculating the length?
	// less efficient, but simpler for higher layers?
	// length should be capped at something reasonable based on assoc negotiation
	binary.BigEndian.PutUint32(w.header[2:6], pdu.Length)

	_, err = w.out.Write(w.header[:])
	if err != nil {
		return
	}

	// should we validate that the amount of data written matches what we said
	// the length was?
	_, err = pdu.Data.WriteTo(w.out)

	return err
}
