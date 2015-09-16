// Copyright 2015 Jeremy Huiskamp <jeremy.huiskamp@gmail.com>
// License: 3-clause BSD

package dcmnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type PDVType uint8

//go:generate stringer -type PDVType
const (
	Data    PDVType = 0x00
	Command PDVType = 0x01
)

const (
	commandDataMask uint8 = 0x01
	lastMask        uint8 = 0x02
)

type PDV struct {
	// TODO make presentation context id a specific type
	Context uint8

	// Flags is a bitmap that contains boolean values for Command/Data and
	// Last/Not Last.  It should be preferred to access these through their
	// respective accessor methods.
	Flags uint8

	Length uint32

	Data io.Reader
}

func (pdv PDV) GetType() PDVType {
	return PDVType(pdv.Flags & commandDataMask)
}

func (pdv *PDV) SetType(pdvType PDVType) {
	pdv.Flags = (pdv.Flags &^ commandDataMask) | uint8(pdvType)
}

func (pdv PDV) IsLast() bool {
	return pdv.Flags&lastMask != 0
}

func (pdv *PDV) SetLast(last bool) {
	if last {
		pdv.Flags = (pdv.Flags &^ lastMask) | 0x02
	} else {
		pdv.Flags = pdv.Flags &^ lastMask
	}
}

func NextPDV(pdata io.Reader) (*PDV, error) {
	var header [6]byte
	_, err := pdata.Read(header[:])
	if err != nil {
		return nil, err
	}

	pdv := PDV{
		Length:  binary.BigEndian.Uint32(header[:4]),
		Context: header[4],
		Flags:   header[5],
	}

	// length includes the Context and Flags
	pdv.Data = io.LimitReader(pdata, int64(pdv.Length)-2)

	return &pdv, nil
}

// PDVReader wraps a PDUReader to act as an io.Reader that reads data spread
// across multiple PDUs
type PDVReader struct {
	pdus PDUReader
	pdu  PDU
	pdv  PDV
}

func ReadPDVs(pdv PDV, pdu PDU, pdus PDUReader) PDVReader {
	return PDVReader{pdus, pdu, pdv}
}

func (p *PDVReader) Read(buf []byte) (n int, err error) {
	for {
		n, err = p.pdv.Data.Read(buf)
		if n > 0 {
			// Got some data from this pdv, return it.
			// We want to ignore underlying EOF here, in case we need to move
			// to the next PDV.  We should be able to ignore other errors, as
			// the underlying stream should return them along with n==0 in the
			// next call.
			return n, nil
		}

		if err != nil && err != io.EOF {
			// real underlying error
			return 0, err
		}

		if err == io.EOF && p.pdv.IsLast() {
			// end of this pdv and its the last one: normal EOF
			return 0, io.EOF
		}

		// move on to the next PDV and then try again
		err = p.nextPDV()
		if err != nil {
			// hmm, an EOF here would be unexpected, should make more serious
			return 0, err
		}
	}
}

func (p *PDVReader) nextPDV() error {
	for {
		nextpdv, err := NextPDV(p.pdu.Data)

		if err == io.EOF {
			err = p.nextPDU()
			if err != nil {
				return err
			}

			// go around again

		} else if err != nil {
			return err
		} else {
			if p.pdv.Context != nextpdv.Context {
				return errors.New(
					fmt.Sprintf("Unexpected PDV context id: %d (expected: %d)",
						nextpdv.Context, p.pdv.Context))

			} else if p.pdv.GetType() != nextpdv.GetType() {
				return errors.New(
					fmt.Sprintf("Unexpected PDV type: %s (expected: %s)",
						nextpdv.GetType(), p.pdv.GetType()))
			}

			p.pdv = *nextpdv
			return nil
		}
	}
}

func (p *PDVReader) nextPDU() error {
	nextpdu, err := p.pdus.NextPDU()
	if err != nil {
		return err
	}

	if nextpdu.Type != PDUPresentationData {
		// TODO: preserve the pdu, eg, in case it's an Abort
		return errors.New(
			fmt.Sprintf("Unepxected PDU type: %s (expected: %s)",
				nextpdu.Type, PDUPresentationData))
	}

	p.pdu = *nextpdu

	return nil
}
