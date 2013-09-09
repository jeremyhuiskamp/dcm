package dcmio

import (
    bin "encoding/binary"
    "github.com/kamper/dcm/dcm"
    "errors"
    "io"
    "io/ioutil"
)

type positionReader struct {
    position uint64
    in io.Reader
}

func (pr *positionReader) Read(p []byte) (n int, err error) {
    n, err = pr.in.Read(p)
    pr.position += uint64(n)
    return
}

type Tag struct {
    // offset of beginning of header from start of stream
    Offset      uint64
    Group       uint16
    Tag         uint16
    VR          dcm.VR
    // offset of beginning of value from start of stream
    ValueOffset uint64
    ValueLength int32
    Value       io.Reader
}

type Parser struct {
    basein *positionReader
    in io.Reader
    // these kind of imitate the transfer syntax:
    order bin.ByteOrder
    explicitVR bool

    Preamble [128]byte

    previousTag *Tag

    // TODO: store sequence / item stack...
    // TODO: remember stream position
}

func (p *Parser) GetPosition() uint64 {
    return p.basein.position
}

func (p *Parser) readTag() (tag *Tag, err error) {
    offset := p.GetPosition()

    var bytes [4]byte
    _, err = io.ReadFull(p.in, bytes[:])

    if err == io.EOF {
        // 0 bytes read, legitimate EOF
        return nil, nil
    } else if err != nil {
        return nil, err
    }

    tag = new(Tag)

    tag.Offset = offset
    tag.Group  = p.order.Uint16(bytes[:2])
    tag.Tag    = p.order.Uint16(bytes[2:])

    return tag, nil
}

func (p *Parser) skipPreviousValue() (err error) {
    if p.previousTag != nil {
        _, err = io.Copy(ioutil.Discard, p.previousTag.Value)
    }
    return err
}

// Returns (nil, nil) if there are no more tags.
func (p *Parser) NextTag() (tag *Tag, err error) {
    err = p.skipPreviousValue()
    if err != nil {
        return nil, err
    }

    // TODO: detect when finished file meta info and switch ts

    tag, err = p.readTag()
    if tag == nil {
        return nil, err
    }

    p.previousTag = tag

    tag.ValueOffset = p.GetPosition()

    if dcm.TagHasVR(tag.Group, tag.Tag) && p.explicitVR {
        var code uint16
        // always big endian?
        err = bin.Read(p.in, bin.BigEndian, &code)
        if err != nil {
            return nil, err
        }

        tag.VR = dcm.GetVR(code)

        var vallen uint16
        err = bin.Read(p.in, p.order, &vallen)
        if err != nil {
            return nil, err
        }

        // if header is longer, the previous 2 bytes were actually just
        // meaningless filler
        if tag.VR.HeaderLength == 8 {
            tag.ValueLength = int32(vallen)
            // TODO: make sure vallen != -1
            tag.Value = io.LimitReader(p.in, int64(vallen))
            return tag, nil
        }
    }

    // nb: signed, it can be -1
    var vallen int32
    err = bin.Read(p.in, p.order, &vallen)
    if err != nil {
        return nil, err
    }

    tag.ValueLength = vallen
    // TODO: make sure vallen != -1
    tag.Value = io.LimitReader(p.in, int64(vallen))

    return tag, nil
}

// Construct a new Parser for a part-10 file
func NewFileParser(in io.Reader) (p Parser, err error) {
    basein := positionReader{position: 0, in: in}
    p = Parser{
        basein: &basein,
        in: &basein,
        // these are the defaults for group 2:
        order: bin.LittleEndian,
        explicitVR: true,
    }

    var preamble [132]byte
    numRead, err := io.ReadFull(p.in, preamble[:])
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

    // TODO: we should actually deliver group length to the caller...
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
