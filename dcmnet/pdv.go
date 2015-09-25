// Copyright 2015 Jeremy Huiskamp <jeremy.huiskamp@gmail.com>
// License: 3-clause BSD

package dcmnet

import (
	"encoding/binary"
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

// PDVDecoder parses successive PDVs from an underlying stream (normally
// a PDataReader)
type PDVDecoder struct {
	data StreamDecoder
}

func NewPDVDecoder(data io.Reader) PDVDecoder {
	return PDVDecoder{StreamDecoder{data, nil}}
}

func (d *PDVDecoder) NextPDV() (pdv *PDV, err error) {
	pdv = &PDV{}
	pdv.Data, err = d.data.NextChunk(6, func(header []byte) int64 {
		pdv.Length = binary.BigEndian.Uint32(header[:4])
		pdv.Context = header[4]
		pdv.Flags = header[5]
		// length includes context and flags:
		return int64(pdv.Length) - 2
	})

	// either error or no more pdvs
	if err != nil || pdv.Data == nil {
		return nil, err
	}

	return pdv, err
}
