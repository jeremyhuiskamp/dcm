package dcmnet

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"io"
)

var pdataLog = log.WithField("unit", "pdata")

// PDataReader combines successive PDUs of type
// PDUPresentationData into a single logical stream.
type PDataReader struct {
	pdus PDUDecoder
	pdu  *PDU
}

func NewPDataReader(pdus PDUDecoder) PDataReader {
	return PDataReader{pdus: pdus, pdu: nil}
}

func (pdr *PDataReader) Read(buf []byte) (int, error) {
	// first time, or after encountering an error (in which case we should
	// probably just hit the same error again):
	if pdr.pdu == nil {
		pdataLog.Debug("No current PDU, checking for next one.")
		err := pdr.nextPDU()
		if err != nil {
			pdataLog.WithError(err).Debug("No next PDU")
			return 0, err
		}

		if pdr.pdu == nil {
			return 0, io.EOF
		}
	}

	for {
		if pdr.pdu == nil || pdr.pdu.Type != PDUPresentationData {
			return 0, io.EOF
		}

		n, err := pdr.pdu.Data.Read(buf)
		if n > 0 {
			// Got some data from this pdu, return it.
			// We want to ignore underlying EOF here, in case we need to move
			// to the next PDU.  We should be able to ignore other errors, as
			// the underlying stream should return them along with n==0 in the
			// next call.
			return n, nil
		}

		if err != nil && err != io.EOF {
			return 0, err
		}

		// end of current PDU, look for next one:
		err = pdr.nextPDU()
		if err != nil {
			return 0, err
		}
	}
}

func (pdr *PDataReader) nextPDU() error {
	nextpdu, err := pdr.pdus.NextPDU()
	if err != nil {
		// TODO: preserve error for inspection after PData?
		// don't preserve current pdu, if any:
		pdr.pdu = nil

		pdataLog.WithError(err).Warn("Unable to read next PDU")

		return err
	}

	pdr.pdu = nextpdu
	if nextpdu != nil {
		pdataLog.WithField("pdutype", nextpdu.Type).Debug("Read next PDU")
	} else {
		pdataLog.Debug("No more PDUs")
	}

	return nil
}

// GetFinalPDU returns the first PDU that was not presentation data.
// Not valid until Read() has returned EOF.  May be nil if the underlying
// stream did not contain another PDU.
func (pdr PDataReader) GetFinalPDU() *PDU {
	return pdr.pdu
}

type PDataWriter struct {
	pdus PDUEncoder
	buf  []byte
}

func NewPDataWriter(pdus PDUEncoder, pdulen uint32) PDataWriter {
	return PDataWriter{pdus, make([]byte, 0, pdulen)}
}

func (pdw *PDataWriter) Write(buf []byte) (int, error) {
	// attempt to copy to internal buffer
	// when full, flush
	written := 0

	// TODO: don't copy to pdw.buf if we don't need to
	// just flush directly from input!
	for written < len(buf) {
		tocopy := cap(pdw.buf) - len(pdw.buf)
		if (len(buf) - written) < tocopy {
			tocopy = len(buf) - written
		}

		pdw.buf = append(pdw.buf, buf[written:(tocopy+written)]...)
		written += tocopy

		if len(pdw.buf) == cap(pdw.buf) {
			err := pdw.flush()
			if err != nil {
				return written, err
			}
		}
	}

	return written, nil
}

func (pdw *PDataWriter) flush() (err error) {
	buf := bytes.NewBuffer(pdw.buf)
	pdu := PDU{
		PDUPresentationData,
		uint32(len(pdw.buf)),
		buf,
	}

	err = pdw.pdus.NextPDU(pdu)
	pdw.buf = pdw.buf[:0]
	return
}

func (pdw *PDataWriter) Close() error {
	if len(pdw.buf) > 0 {
		return pdw.flush()
	} else {
		return nil
	}
}
