package dcmnet

import (
	"encoding/binary"
	"io"
	"io/ioutil"
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
	Data   io.Reader
}

// PDUDecoder parses a stream for PDUs
type PDUDecoder struct {
	data    io.Reader
	lastPDU *PDU
}

func NewPDUDecoder(data io.Reader) PDUDecoder {
	return PDUDecoder{data, nil}
}

// Read the next PDU from the stream
func (reader *PDUDecoder) NextPDU() (*PDU, error) {
	// discard previous pdu, if caller didn't already do so
	if reader.lastPDU != nil {
		_, err := io.Copy(ioutil.Discard, reader.lastPDU.Data)
		if err != nil {
			return nil, err
		}
	}

	header := make([]byte, 6)
	len, err := io.ReadFull(reader.data, header)
	if len == 0 {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	reader.lastPDU = &PDU{
		Type:   PDUType(header[0]),
		Length: binary.BigEndian.Uint32(header[2:6]),
	}

	// TODO: sanity check the length?
	reader.lastPDU.Data = io.LimitReader(reader.data, int64(reader.lastPDU.Length))

	return reader.lastPDU, nil
}

type PDUWriter struct {
	out    io.Writer
	header [6]byte
}

func NewPDUWriter(out io.Writer) PDUWriter {
	return PDUWriter{
		out: out,
	}
}

func (w *PDUWriter) Write(pdu PDU) (err error) {
	w.header[0] = uint8(pdu.Type)
	w.header[1] = 0
	binary.BigEndian.PutUint32(w.header[2:6], pdu.Length)

	_, err = w.out.Write(w.header[:])
	if err != nil {
		return err
	}

	_, err = io.Copy(w.out, pdu.Data)

	return err
}
