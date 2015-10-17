// Copyright 2015 Jeremy Huiskamp <jeremy.huiskamp@gmail.com>
// License: 3-clause BSD

package dcmnet

import (
	"encoding/binary"
	"fmt"
	"github.com/kamper/dcm/stream"
	"io"
)

type PDVType uint8

//go:generate stringer -type PDVType
const (
	Data    PDVType = 0x00
	Command PDVType = 0x01
)

const (
	commandDataMask byte = 0x01
	lastMask        byte = 0x02
)

// PDVFlags is a bitmap that contains boolean values for Command/Data and
// Last/Not Last.
type PDVFlags uint8

func (flags PDVFlags) GetType() PDVType {
	return PDVType(uint8(flags) & commandDataMask)
}

func (flags *PDVFlags) SetType(pdvType PDVType) {
	*flags = PDVFlags((uint8(*flags) &^ commandDataMask) | uint8(pdvType))
}

func (flags PDVFlags) IsLast() bool {
	return uint8(flags)&lastMask != 0
}

func (flags *PDVFlags) SetLast(last bool) {
	if last {
		*flags = PDVFlags((uint8(*flags) &^ lastMask) | 0x02)
	} else {
		*flags = PDVFlags(uint8(*flags) &^ lastMask)
	}
}

func (flags PDVFlags) String() string {
	return fmt.Sprintf("[Flags type=%s, last=%t]", flags.GetType(), flags.IsLast())
}

type PDV struct {
	Context PCID

	Flags PDVFlags

	Length uint32

	Data stream.Stream
}

func (pdv PDV) GetType() PDVType {
	return pdv.Flags.GetType()
}

func (pdv PDV) IsLast() bool {
	return pdv.Flags.IsLast()
}

func (pdv PDV) String() string {
	return fmt.Sprintf("[PDV %s, context=%d, len=%d]",
		pdv.Flags, pdv.Context, pdv.Length)
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
		pdv.Context = PCID(header[4])
		pdv.Flags = PDVFlags(header[5])
		// length includes context and flags:
		return int64(pdv.Length) - 2
	})

	// either error or no more pdvs
	if err != nil || pdv.Data == nil {
		return nil, err
	}

	return pdv, err
}
