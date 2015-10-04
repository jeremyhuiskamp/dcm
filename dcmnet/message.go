package dcmnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/kamper/dcm/log"
	"github.com/kamper/dcm/stream"
	"io"
	"io/ioutil"
)

var msgLog = log.Category("dcm.msg")

// A MessageElement is either a Command Set or a Data Set
type MessageElement struct {
	Context PCID
	Type    PDVType
	Data    stream.Stream
}

func (msg MessageElement) String() string {
	return fmt.Sprintf("[%s MessageElement context=%d]", msg.Type, msg.Context)
}

// MessageElementDecoder decodes successive message elements from an underlying
// stream of PDVs
type MessageElementDecoder struct {
	pdvs PDVDecoder
	msg  *MessageElementReader
}

func NewMessageElementDecoder(pdvs PDVDecoder) MessageElementDecoder {
	return MessageElementDecoder{pdvs, nil}
}

func (md *MessageElementDecoder) NextMessageElement() (*MessageElement, error) {
	if md.msg != nil {
		msgLog.Debug("Draining previous message")
		io.Copy(ioutil.Discard, md.msg)
	}

	pdv, err := md.pdvs.NextPDV()
	if pdv == nil && (err == nil || err == io.EOF) {
		msgLog.Debug("No more PDVs")
		return nil, nil
	}

	if err != nil {
		msgLog.WithError(err).Debug("Unexpected error while retrieving next PDV")
		return nil, err
	}

	md.msg = &MessageElementReader{md.pdvs, pdv}

	return &MessageElement{
		pdv.Context,
		pdv.GetType(),
		stream.NewReaderStream(md.msg),
	}, nil
}

// MessageElementReader implements io.Reader by combining the data of several PDVs
type MessageElementReader struct {
	pdvs PDVDecoder
	pdv  *PDV
}

func (mer *MessageElementReader) Read(buf []byte) (int, error) {
	// TODO: only set this up at beginning and put it in struct
	log := msgLog.WithFields(logrus.Fields{
		"contextid": mer.pdv.Context,
		"pdvtype":   mer.pdv.GetType()})

	for {
		n, err := mer.pdv.Data.Read(buf)
		log.Debugf("Read %d bytes", n)
		if n > 0 {
			return n, nil
		}

		if err != nil && err != io.EOF {
			log.WithError(err).Warn("Unexpected error while reading from existing PDV")
			return 0, err
		}

		if err == io.EOF && mer.pdv.IsLast() {
			log.Debug("No more data in this message element")
			return 0, io.EOF
		}

		log.Debug("This PDV has been read. Checking for the next one.")
		err = mer.nextPDV()
		if err != nil {
			log.WithError(err).Warn("Unexpected error while getting next PDV")
			return 0, err
		}
	}
}

func (mer *MessageElementReader) nextPDV() error {
	nextpdv, err := mer.pdvs.NextPDV()
	if err != nil {
		// hmm, probably want to mark some struct state, since we don't really
		// want to keep trying this in subsequent calls
		msgLog.WithError(err).Warn("Unexpected error while retrieving next PDV")
		return err
	}

	if nextpdv == nil {
		return io.ErrUnexpectedEOF
	}

	if mer.pdv.Context != nextpdv.Context {
		return errors.New(fmt.Sprintf(
			"Received PDV with unexpected presentation context %d "+
				"(expected %d)", nextpdv.Context, mer.pdv.Context))
	}

	if mer.pdv.GetType() != nextpdv.GetType() {
		return errors.New(fmt.Sprintf(
			"Received PDV with unexpected type %s "+
				"(expected %s)", nextpdv.GetType(), mer.pdv.GetType()))
	}

	mer.pdv = nextpdv

	return nil
}

const (
	pdvHeaderLen = 6
	minPDULen    = pdvHeaderLen + 1
)

// MessageElementEncoder encodes successive message elements to PDUs.
// Exactly one PDV is written per PDU (the standard does not require this in all
// cases, but it is simpler to implement and the overhead of an unnecessary PDU
// header is not very big).
//
// Unlike MessageElementDecoder, this ties down into the PDU layer because we
// have to be (and can be) more strict in the sense that MessageElementDecoder &
// friends can handle PDVs split across multiple PDUs, which is not legal to
// send.  As such, we have to know where the PDU buf ends instead of hoping that
// we won't make our PDVs too long.  It turns out to be simpler to bake
// everything into one layer.
type MessageElementEncoder struct {
	pdus      PDUEncoder
	maxPDUlen uint32
}

func NewMessageElementEncoder(pdus PDUEncoder, maxPDULen uint32) MessageElementEncoder {
	return MessageElementEncoder{pdus, maxPDULen}
}

func (mee *MessageElementEncoder) NextMessageElement(msg MessageElement) error {
	mew := NewMessageElementWriter(mee.pdus, mee.maxPDUlen, msg.Context, msg.Type)
	_, err := msg.Data.WriteTo(&mew)
	if err != nil {
		return err
	}

	return mew.flush(true)
}

// MessageElementWriter implements io.Reader to write a single MessageElement
// in a series of PDUs.
type MessageElementWriter struct {
	pdus PDUEncoder

	// cap() == maxPDULen
	pduBuf []byte

	// is pduBuf[:6]
	pdvHeader []byte

	pdvFlags PDVFlags

	// is pduBuf[6:x] where x is the current amount of data buffered
	pdvBody []byte
}

func NewMessageElementWriter(pdus PDUEncoder, maxPDULen uint32, context PCID,
	pdvType PDVType) MessageElementWriter {

	if maxPDULen < minPDULen {
		panic(fmt.Sprintf("PDU length %d must be at least %d to allow room "+
			"for the PDV header plus at least one byte of data.",
			maxPDULen, minPDULen))
	}

	mew := MessageElementWriter{
		pdus:   pdus,
		pduBuf: make([]byte, maxPDULen),
	}

	mew.pdvHeader = mew.pduBuf[:pdvHeaderLen]
	// these are the same for the whole message:
	mew.pdvHeader[4] = byte(context)
	mew.pdvFlags.SetType(pdvType)
	// the other header fields are dependent on each individual pdv

	mew.pdvBody = mew.pduBuf[pdvHeaderLen:pdvHeaderLen]

	return mew
}

func (mew *MessageElementWriter) Write(buf []byte) (int, error) {
	written := 0

	for written < len(buf) {
		if len(mew.pdvBody) == cap(mew.pdvBody) {
			// we have more data, so the current buffer can't be the last one
			err := mew.flush(false)
			if err != nil {
				return written, err
			}
		}

		tocopy := cap(mew.pdvBody) - len(mew.pdvBody)
		if len(buf)-written < tocopy {
			tocopy = len(buf) - written
		}

		// this had better not cause a re-allocation!
		// TODO: if we're going to hit capacity, try to flush from buf directly
		// instead of copying
		mew.pdvBody = append(mew.pdvBody, buf[written:(tocopy+written)]...)
		written += tocopy
	}

	return written, nil
}

func (mew *MessageElementWriter) flush(last bool) error {
	// pdv length is body length + space for flags & context:
	pdvlen := uint32(len(mew.pdvBody) + 2)

	// to safely allow flushing after writing all data, if we're not sure if
	// there's unflushed data or not:
	if pdvlen <= 2 {
		return nil
	}

	// lay out header:
	binary.BigEndian.PutUint32(mew.pdvHeader[:4], pdvlen)
	mew.pdvFlags.SetLast(last)
	mew.pdvHeader[5] = byte(mew.pdvFlags)

	// -2 because we already counted flags & context in the pdvlen
	pdulen := pdvlen + (pdvHeaderLen - 2)
	pdu := PDU{
		Type:   PDUPresentationData,
		Length: pdulen,
		Data:   bytes.NewReader(mew.pduBuf[:pdulen]),
	}

	err := mew.pdus.NextPDU(pdu)

	mew.pdvBody = mew.pdvBody[:0]

	return err
}
