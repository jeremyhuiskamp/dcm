package dcmnet

import (
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
		log.Debug("No current PDU, checking for next one.")
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
		if pdr.pdu.Type != PDUPresentationData {
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

		if pdr.pdu == nil {
			return 0, io.EOF
		}
	}
}

func (pdr *PDataReader) nextPDU() error {
	nextpdu, err := pdr.pdus.NextPDU()
	if err != nil {
		// TODO: preserve error for inspection after PData?
		// don't preserve current pdu, if any:
		pdr.pdu = nil

		return err
	}

	pdr.pdu = nextpdu

	return nil
}