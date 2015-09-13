// Copyright 2015 Jeremy Huiskamp <jeremy.huiskamp@gmail.com>
// License: 3-clause BSD

package dcmnet

import (
	"encoding/binary"
	"io"
)

type PDVType uint8

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
	var pdv PDV

	var header [6]byte
	_, err := pdata.Read(header[:])
	if err != nil {
		return nil, err
	}

	var length uint32
	binary.BigEndian.Uint32(header[:4])
	pdv.Context = header[4]
	pdv.Flags = header[5]

	// length includes the Context and Flags
	pdv.Data = io.LimitReader(pdata, int64(length-2))

	return &pdv, nil
}
