package dcmnet

import (
    "encoding/binary"
    "io"
    "io/ioutil"
)

// PDU Types
const (
    AssociateRQ uint8 = 0x01
    AssociateAC       = 0x02
    AssociateRJ       = 0x03
    PresentationData  = 0x04
    ReleaseRQ         = 0x05
    ReleaseRP         = 0x06
    Abort             = 0x07
)

// Protocol Data Unit
type PDU struct {
    Type uint8
    Length uint32
    Data io.Reader
}

// PDUReader parses a stream for PDUs
type PDUReader struct {
    data io.Reader
    lastPDU *PDU
}

func NewPDUReader(data io.Reader) (PDUReader) {
    return PDUReader{data, nil}
}

// Read the next PDU from the stream
func (reader *PDUReader) NextPDU() (*PDU, error) {
    // discard previous pdu, if caller didn't already do so
    if reader.lastPDU != nil {
        _, err := io.Copy(ioutil.Discard, reader.lastPDU.Data)
        if err != nil {
            return nil, err
        }
    }

    // TODO: detect eof and return nil?
    header := make([]byte, 6)
    headerlen, err := io.ReadFull(reader.data, header)
    if headerlen == 0 {
        return nil, nil
    }

    if err != nil {
        return nil, err
    }

    reader.lastPDU = &PDU{
        Type: header[0],
        Length: binary.BigEndian.Uint32(header[2:6]),
    }

    // TODO: sanity check the length?
    reader.lastPDU.Data = io.LimitReader(reader.data, int64(reader.lastPDU.Length))

    return reader.lastPDU, nil
}

// TODO: describe Item
type Item struct {
    Type uint8
    Length uint16
    Data io.Reader
}

type ItemReader struct {
    data io.Reader
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
        Type: header[0],
        Length: binary.BigEndian.Uint16(header[2:4]),
    }

    // TODO: sanity check the length?
    reader.lastItem.Data = io.LimitReader(reader.data, int64(reader.lastItem.Length))

    return reader.lastItem, nil
}

