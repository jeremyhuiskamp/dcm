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

type Message struct {
	Context PCID
	Type    PDVType
	Data    stream.Stream
}

// MessageDecoder decodes successive messages from an underlying stream of
// PDVs
type MessageDecoder struct {
	pdvs PDVDecoder
	msg  *MessageReader
}

func NewMessageDecoder(pdvs PDVDecoder) MessageDecoder {
	return MessageDecoder{pdvs, nil}
}

func (md *MessageDecoder) NextMessage() (*Message, error) {
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

	md.msg = &MessageReader{md.pdvs, pdv}

	return &Message{
		pdv.Context,
		pdv.GetType(),
		stream.NewReaderStream(md.msg),
	}, nil
}

// MessageReader implements io.Reader by combining the data of several PDVs
type MessageReader struct {
	pdvs PDVDecoder
	pdv  *PDV
}

func (mr *MessageReader) Read(buf []byte) (int, error) {
	// TODO: only set this up at beginning and put it in struct
	log := msgLog.WithFields(logrus.Fields{
		"contextid": mr.pdv.Context,
		"pdvtype":   mr.pdv.GetType()})

	for {
		n, err := mr.pdv.Data.Read(buf)
		log.Debugf("Read %d bytes", n)
		if n > 0 {
			return n, nil
		}

		if err != nil && err != io.EOF {
			log.WithError(err).Warn("Unexpected error while reading from existing PDV")
			return 0, err
		}

		if err == io.EOF && mr.pdv.IsLast() {
			log.Debug("No more data in this message")
			return 0, io.EOF
		}

		log.Debug("This PDV has been read. Checking for the next one.")
		err = mr.nextPDV()
		if err != nil {
			log.WithError(err).Warn("Unexpected error while getting next PDV")
			return 0, err
		}
	}
}

func (mr *MessageReader) nextPDV() error {
	nextpdv, err := mr.pdvs.NextPDV()
	if err != nil {
		// hmm, probably want to mark some struct state, since we don't really
		// want to keep trying this in subsequent calls
		msgLog.WithError(err).Warn("Unexpected error while retrieving next PDV")
		return err
	}

	if nextpdv == nil {
		return io.ErrUnexpectedEOF
	}

	if mr.pdv.Context != nextpdv.Context {
		return errors.New(fmt.Sprintf(
			"Received PDV with unexpected presentation context %d "+
				"(expected %d)", nextpdv.Context, mr.pdv.Context))
	}

	if mr.pdv.GetType() != nextpdv.GetType() {
		return errors.New(fmt.Sprintf(
			"Received PDV with unexpected type %s "+
				"(expected %s)", nextpdv.GetType(), mr.pdv.GetType()))
	}

	mr.pdv = nextpdv

	return nil
}

const (
	pdvHeaderLen = 6
	minPDULen    = pdvHeaderLen + 1
)

// MessageEncoder encodes successive messages to PDUs.  Exactly one PDV is
// written per PDU (the standard does not require this in all cases, but it is
// simpler to implement and the overhead of an unnecessary PDU header is not
// very big).
//
// Unlike MessageDecoder, this ties down into the PDU layer because we have to
// be (and can be) more strict in the sense that MessageDecoder & friends can
// handle PDVs split across multiple PDUs, which is not legal to send.  As such,
// we have to know where the PDU buf ends instead of hoping that we won't make
// our PDVs too long.  It turns out to be simpler to bake everything into one
// layer.
type MessageEncoder struct {
	pdus      PDUEncoder
	maxPDUlen uint32
}

func NewMessageEncoder(pdus PDUEncoder, maxPDULen uint32) MessageEncoder {
	return MessageEncoder{pdus, maxPDULen}
}

func (me *MessageEncoder) NextMessage(msg Message) error {
	mw := NewMessageWriter(me.pdus, me.maxPDUlen, msg.Context, msg.Type)
	_, err := msg.Data.WriteTo(&mw)
	if err != nil {
		return err
	}

	return mw.flush(true)
}

// MessageWriter implements io.Reader to write a single Message in a series of PDUs.
type MessageWriter struct {
	pdus PDUEncoder

	// cap() == maxPDULen
	pduBuf []byte

	// is pduBuf[:6]
	pdvHeader []byte

	pdvFlags PDVFlags

	// is pduBuf[6:x] where x is the current amount of data buffered
	pdvBody []byte
}

func NewMessageWriter(pdus PDUEncoder, maxPDULen uint32, context PCID,
	pdvType PDVType) MessageWriter {

	if maxPDULen < minPDULen {
		panic(fmt.Sprintf("PDU length %d must be at least %d to allow room "+
			"for the PDV header plus at least one byte of data.",
			maxPDULen, minPDULen))
	}

	mw := MessageWriter{
		pdus:   pdus,
		pduBuf: make([]byte, maxPDULen),
	}

	mw.pdvHeader = mw.pduBuf[:pdvHeaderLen]
	// these are the same for the whole message:
	mw.pdvHeader[4] = byte(context)
	mw.pdvFlags.SetType(pdvType)
	// the other header fields are dependent on each individual pdv

	mw.pdvBody = mw.pduBuf[pdvHeaderLen:pdvHeaderLen]

	return mw
}

func (mw *MessageWriter) Write(buf []byte) (int, error) {
	written := 0

	for written < len(buf) {
		if len(mw.pdvBody) == cap(mw.pdvBody) {
			// we have more data, so the current buffer can't be the last one
			err := mw.flush(false)
			if err != nil {
				return written, err
			}
		}

		tocopy := cap(mw.pdvBody) - len(mw.pdvBody)
		if len(buf)-written < tocopy {
			tocopy = len(buf) - written
		}

		// this had better not cause a re-allocation!
		// TODO: if we're going to hit capacity, try to flush from buf directly
		// instead of copying
		mw.pdvBody = append(mw.pdvBody, buf[written:(tocopy+written)]...)
		written += tocopy
	}

	return written, nil
}

func (mw *MessageWriter) flush(last bool) error {
	// pdv length is body length + space for flags & context:
	pdvlen := uint32(len(mw.pdvBody) + 2)

	// to safely allow flushing after writing all data, if we're not sure if
	// there's unflushed data or not:
	if pdvlen <= 2 {
		return nil
	}

	// lay out header:
	binary.BigEndian.PutUint32(mw.pdvHeader[:4], pdvlen)
	mw.pdvFlags.SetLast(last)
	mw.pdvHeader[5] = byte(mw.pdvFlags)

	// -2 because we already counted flags & context in the pdvlen
	pdulen := pdvlen + (pdvHeaderLen - 2)
	pdu := PDU{
		Type:   PDUPresentationData,
		Length: pdulen,
		Data:   bytes.NewReader(mw.pduBuf[:pdulen]),
	}

	err := mw.pdus.NextPDU(pdu)

	mw.pdvBody = mw.pdvBody[:0]

	return err
}
