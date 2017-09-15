package dcmnet

// An integrated test that parses a dicom stream structured with all the layers
// from pdu to message

import (
	"bytes"
	"testing"
)

func TestCommandAndDataSamePDU(t *testing.T) {
	data := bufcat(
		bufpdu(PDUPresentationData,
			bufpdv(1, Command, true, "command"),
			bufpdv(1, Data, true, "data")),
		bufpdu(PDUReleaseRQ))

	pdata, msgs := setupParse(data)

	expectMessageElement(t, msgs, 1, Command, "command")
	expectMessageElement(t, msgs, 1, Data, "data")
	expectNoMoreMessageElements(t, msgs)

	expectRelease(t, pdata)
}

func TestCommandAndDataDifferentPDU(t *testing.T) {
	data := bufcat(
		bufpdu(PDUPresentationData,
			bufpdv(1, Command, true, "command")),
		bufpdu(PDUPresentationData,
			bufpdv(1, Data, true, "data")),
		bufpdu(PDUReleaseRQ))

	pdata, msgs := setupParse(data)

	expectMessageElement(t, msgs, 1, Command, "command")
	expectMessageElement(t, msgs, 1, Data, "data")
	expectNoMoreMessageElements(t, msgs)

	expectRelease(t, pdata)
}

func TestCommandAndDataOverTwoPDUs(t *testing.T) {
	data := bufcat(
		bufpdu(PDUPresentationData,
			bufpdv(1, Command, true, "command")),
		bufpdu(PDUPresentationData,
			bufpdv(1, Data, false, "data1")),
		bufpdu(PDUPresentationData,
			bufpdv(1, Data, true, "data2")),
		bufpdu(PDUReleaseRQ))

	pdata, msgs := setupParse(data)

	expectMessageElement(t, msgs, 1, Command, "command")
	expectMessageElement(t, msgs, 1, Data, "data1data2")
	expectNoMoreMessageElements(t, msgs)

	expectRelease(t, pdata)
}

func expectRelease(t *testing.T, pdata *PDataReader) {
	releaserq := pdata.GetFinalPDU()
	if releaserq == nil {
		t.Fatal("didn't get expected release request pdu")
	}
	if releaserq.Type != PDUReleaseRQ {
		t.Fatalf("unexpected pdu type: %s", releaserq.Type)
	}
	if releaserq.Length != uint32(0) {
		t.Fatalf("unexpected pdu length: %d", releaserq.Length)
	}
}

func setupParse(buf bytes.Buffer) (*PDataReader, *MessageElementDecoder) {
	pdus := NewPDUDecoder(&buf)
	pdata := NewPDataReader(pdus)
	pdvs := NewPDVDecoder(&pdata)
	msgs := NewMessageElementDecoder(pdvs)
	return &pdata, msgs
}
