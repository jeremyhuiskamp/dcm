package dcmio

// Things to try to add:
// - switch transfer syntaxes in sequence items with an unknown VR
// - implement fast jumps over known-length things
// -- eg, if we have group, sequence or item lengths
// -- might want to check if underlying stream supports seek, for speed

import (
	"bytes"
	bin "encoding/binary"
	"errors"
	"github.com/kamper/dcm/dcm"
	"io"
	"io/ioutil"
	"strings"
)

// positionReader wraps another io.Reader and counts the
// number of bytes that have been read
type positionReader struct {
	position uint64
	in       io.Reader
}

func (pr *positionReader) Read(p []byte) (n int, err error) {
	n, err = pr.in.Read(p)
	pr.position += uint64(n)
	return
}

// DICOM files consist of a list of Tags
// TODO: This would probably better be called an Element...
type Tag struct {
	// offset of beginning of header from start of stream
	Offset uint64
	Tag    dcm.Tag
	VR     *dcm.VR
	// offset of beginning of value from start of stream
	ValueOffset uint64
	ValueLength int32
	Value       io.Reader
}

type Parser interface {
	// Returns the current number of bytes read from the
	// start of the stream
	GetPosition() uint64

	// Reads the next tag from the stream.
	// Returns (nil, nil) if EOF is reached gracefully
	NextTag() (*Tag, error)
}

// SimpleParser is a Parser implementation that simply reads
// from the stream, always assuming a given transfer syntax
type SimpleParser struct {
	basein *positionReader
	in     io.Reader

	ts dcm.TransferSyntax

	// track the previous tag we returned, so that we can
	// make sure it's stream is drained
	previousTag *Tag
}

func (p *SimpleParser) GetPosition() uint64 {
	return p.basein.position
}

func (p *SimpleParser) readTag() (tag *Tag, err error) {
	tag = new(Tag)
	tag.Offset = p.GetPosition()

	var bytes [4]byte
	_, err = io.ReadFull(p.in, bytes[:])

	if err == io.EOF {
		// 0 bytes read, legitimate EOF
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	order := p.ts.ByteOrder()
	tag.Tag = dcm.NewTag(
		order.Uint16(bytes[:2]),
		order.Uint16(bytes[2:]))

	return tag, nil
}

func (p *SimpleParser) skipPreviousValue() (err error) {
	if p.previousTag != nil {
		_, err = io.Copy(ioutil.Discard, p.previousTag.Value)
	}
	return err
}

func (p *SimpleParser) NextTag() (tag *Tag, err error) {
	err = p.skipPreviousValue()
	if err != nil {
		return nil, err
	}

	tag, err = p.readTag()
	if tag == nil {
		return nil, err
	}

	p.previousTag = tag

	// we leave the stream at the beginning of the value, so:
	defer func() { tag.ValueOffset = p.GetPosition() }()

	if tag.Tag.HasVR() && p.ts.VR() == dcm.Explicit {

		var vr [2]byte
		_, err = io.ReadFull(p.in, vr[:])
		if err != nil {
			return nil, err
		}

		tag.VR = dcm.GetVRByName(string(vr[:]))

		var vallen uint16
		err = bin.Read(p.in, p.ts.ByteOrder(), &vallen)
		if err != nil {
			return nil, err
		}

		if !tag.VR.Long {
			tag.ValueLength = int32(vallen)
			// TODO: make sure vallen != -1
			tag.Value = io.LimitReader(p.in, int64(vallen))
			return tag, nil
		}

		// if header is long, the previous 2 bytes were actually just
		// meaningless filler and we go on to do the 4-byte length

	} else {
		// implicit vr, have to guess:
		// TODO: track private creator uids
		vr := dcm.VRForTag("", tag.Tag)
		tag.VR = &vr
	}

	// nb: signed, it can be -1
	var vallen int32
	err = bin.Read(p.in, p.ts.ByteOrder(), &vallen)
	if err != nil {
		return nil, err
	}

	tag.ValueLength = vallen
	// TODO: make sure vallen != -1
	tag.Value = io.LimitReader(p.in, int64(vallen))

	return tag, nil
}

type part10state int

const (
	beforeGroup2 part10state = iota
	inGroup2
	pastGroup2
)

// Part10Parser is a Parser implementation that manages multiple
// SimpleParser instances to read the parts of the file that have
// different transfer syntaxes
type Part10Parser struct {
	basein *positionReader
	parser SimpleParser
	state  part10state
	ts     dcm.TransferSyntax
}

func (p *Part10Parser) GetPosition() uint64 {
	return p.basein.position
}

func (p *Part10Parser) NextTag() (tag *Tag, err error) {
	switch p.state {
	case beforeGroup2:
		tag, err = p.parser.NextTag()
		if err != nil {
			return nil, err
		}

		if tag == nil {
			return nil, nil
		}

		if tag.Tag == dcm.FileMetaInformationGroupLength {
			p.state = inGroup2

			buf, err := bufferValue(tag)
			if err != nil {
				return nil, err
			}

			// TODO: validate VR and VL?
			var fmiLength uint32
			err = bin.Read(&buf, p.ts.ByteOrder(), &fmiLength)
			if err != nil {
				return nil, err
			}

			p.parser = SimpleParser{
				basein: p.basein,
				in:     io.LimitReader(p.basein, int64(fmiLength)),
				ts:     p.ts,
			}
		}

		return tag, nil

	case inGroup2:
		tag, err = p.parser.NextTag()
		if err != nil {
			return nil, err
		}

		if tag == nil {
			p.state = pastGroup2

			if p.ts == nil {
				// shouldn't happen, but we'll keep
				// going with the group2 default
				p.ts = dcm.ExplicitVRLittleEndian
			}

			p.parser = SimpleParser{
				basein: p.basein,
				in:     p.basein,
				ts:     p.ts,
			}

			// recurse:
			return p.NextTag()
		}

		if tag.Tag == dcm.TransferSyntaxUID {
			buf, err := bufferValue(tag)
			if err != nil {
				return tag, err
			}

			// TODO: learn how buf.String() handles charsets
			// TODO: factor this out into a better place
			tsstr := strings.Trim(buf.String(), " \x00")

			p.ts = dcm.GetTransferSyntax(tsstr)
		}

		return tag, nil

	case pastGroup2:
		fallthrough
	default:
		return p.parser.NextTag()
	}
}

func bufferValue(tag *Tag) (buf bytes.Buffer, err error) {
	_, err = buf.ReadFrom(tag.Value)
	if err != nil {
		return
	}

	// make a copy so the tag can still be read:
	tag.Value = bytes.NewBuffer(buf.Bytes())
	return
}

// Construct a new Parser for a part-10 file
func NewFileParser(in io.Reader) (Parser, error) {
	basein := positionReader{position: 0, in: in}

	var preamble [132]byte
	numRead, err := io.ReadFull(&basein, preamble[:])
	if numRead < len(preamble) {
		return nil, io.EOF
	}

	if err != nil {
		return nil, err
	}

	dicm := preamble[128:132]
	if !(dicm[0] == 'D' && dicm[1] == 'I' && dicm[2] == 'C' && dicm[3] == 'M') {
		return nil, errors.New("Not a Part10 file")
	}

	defTS := dcm.ExplicitVRLittleEndian

	p := &Part10Parser{
		basein: &basein,
		parser: SimpleParser{
			basein: &basein,
			in:     &basein,
			ts:     defTS,
		},
		ts:    defTS,
		state: beforeGroup2,
	}

	return p, nil
}

// Construct a new Parser for a stream without a part-10 header
func NewStreamParser(in io.Reader, ts dcm.TransferSyntax) Parser {
	basein := positionReader{position: 0, in: in}

	return &SimpleParser{
		basein: &basein,
		in:     &basein,
		ts:     ts,
	}
}
