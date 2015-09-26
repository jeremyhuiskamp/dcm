package dcmnet

import (
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
	Context uint8
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
