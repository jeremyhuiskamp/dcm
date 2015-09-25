package dcmnet

// An integrated test that parses a dicom stream structured with all the layers
// from pdu to message

import (
	"bytes"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCommandAndDataSamePDU(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData,
			bufpdv(1, Command, true, "command"),
			bufpdv(1, Data, true, "data")),
		bufpdu(PDUReleaseRQ))

	pdata, msgs := setupParse(data)

	expectMessage(msgs, 1, Command, "command")
	expectMessage(msgs, 1, Data, "data")
	expectNoMoreMessages(msgs)

	expectRelease(pdata)
}

func TestCommandAndDataDifferentPDU(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData,
			bufpdv(1, Command, true, "command")),
		bufpdu(PDUPresentationData,
			bufpdv(1, Data, true, "data")),
		bufpdu(PDUReleaseRQ))

	pdata, msgs := setupParse(data)

	expectMessage(msgs, 1, Command, "command")
	expectMessage(msgs, 1, Data, "data")
	expectNoMoreMessages(msgs)

	expectRelease(pdata)
}

func TestCommandAndDataOverTwoPDUs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData,
			bufpdv(1, Command, true, "command")),
		bufpdu(PDUPresentationData,
			bufpdv(1, Data, false, "data1")),
		bufpdu(PDUPresentationData,
			bufpdv(1, Data, true, "data2")),
		bufpdu(PDUReleaseRQ))

	pdata, msgs := setupParse(data)

	expectMessage(msgs, 1, Command, "command")
	expectMessage(msgs, 1, Data, "data1data2")
	expectNoMoreMessages(msgs)

	expectRelease(pdata)
}

func expectRelease(pdata *PDataReader) {
	releaserq := pdata.GetFinalPDU()
	Expect(releaserq).ToNot(BeNil())
	Expect(releaserq.Type).To(Equal(PDUReleaseRQ))
	Expect(releaserq.Length).To(Equal(uint32(0)))
}

func setupParse(buf bytes.Buffer) (*PDataReader, MessageDecoder) {
	pdus := NewPDUDecoder(&buf)
	pdata := NewPDataReader(pdus)
	pdvs := NewPDVDecoder(&pdata)
	msgs := NewMessageDecoder(pdvs)
	return &pdata, msgs
}
