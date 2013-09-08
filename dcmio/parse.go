package dcmio

import (
    bin "encoding/binary"
    "github.com/kamper/dcm/dcm"
    "errors"
    "io"
)

type Tag struct {
    Group uint16
    Tag   uint16
    VR    dcm.VR
    Value io.Reader

    // TODO: remember stream offsets of header and value?
}

type Parser struct {
    in io.Reader
    // these kind of imitate the transfer syntax:
    order bin.ByteOrder
    explicitVR bool

    Preamble [128]byte

    // TODO: store sequence / item stack...
    // TODO: remember stream position
}

func (p *Parser) readTag(tag *Tag) (err error) {
    err = bin.Read(p.in, p.order, &tag.Group)
    if err != nil {
        return err
    }

    return bin.Read(p.in, p.order, &tag.Tag)
}

// TODO: differentiate between EOF as error and legitimate finish
// probably we want to return a nil Tag
func (p *Parser) NextTag() (tag Tag, err error) {
    // TODO: make sure previous value is drained...
    // TODO: detect when finished file meta info and switch ts

    err = p.readTag(&tag)
    if err != nil {
        return tag, err
    }

    if dcm.TagHasVR(tag.Group, tag.Tag) && p.explicitVR {
        var code uint16
        // always big endian?
        err = bin.Read(p.in, bin.BigEndian, &code)
        if err != nil {
            return tag, err
        }

        tag.VR = dcm.GetVR(code)

        var vallen uint16
        err = bin.Read(p.in, p.order, &vallen)
        if err != nil {
            return tag, err
        }

        // if header is longer, the previous 2 bytes were actually just
        // meaningless filler
        if tag.VR.HeaderLength == 8 {
            // TODO: make sure vallen != -1
            tag.Value = io.LimitReader(p.in, int64(vallen))
            return tag, nil
        }
    }

    // nb: signed, it can be -1
    var vallen int32
    err = bin.Read(p.in, p.order, &vallen)
    if err != nil {
        return tag, err
    }

    // TODO: make sure vallen != -1
    tag.Value = io.LimitReader(p.in, int64(vallen))

    return tag, nil
}

// Construct a new Parser for a part-10 file
func NewFileParser(in io.Reader) (p Parser, err error) {
    // these are the defaults for group 2:
    p = Parser{ in: in, order: bin.LittleEndian, explicitVR: true }

    var preamble [132]byte
    numRead, err := io.ReadFull(in, preamble[:])
    if numRead < len(preamble) {
        return p, io.EOF
    }

    if err != nil {
        return p, err
    }

    //p.Preamble = preamble[0:128]

    dicm := preamble[128:132]
    if !(dicm[0] == 'D' && dicm[1] == 'I' && dicm[2] == 'C' && dicm[3] == 'M') {
        return p, errors.New("Not a Part10 file")
    }

    tag, err := p.NextTag()
    if err != nil {
        return p, err
    }

    if tag.Group != 0x0002 || tag.Tag != 0x0000 {
        return p, errors.New("Missing File Meta-Information Group Length")
    }

    // TODO: ensure VR=UL
    var fmiLength uint32
    err = bin.Read(tag.Value, p.order, &fmiLength)
    if err != nil {
        return p, err
    }

    // this actually limits us to reading the fmi, instead we should be
    // providing a way to switch over to the data in a different transfer
    // syntax
    p.in = io.LimitReader(p.in, int64(fmiLength))

    return p, nil
}

// Construct a new Parser for a stream without a part-10 header
func NewStreamParser(in io.Reader, order bin.ByteOrder, explicitVR bool) Parser {
    return Parser{ in: in, order: order, explicitVR: explicitVR }
}
