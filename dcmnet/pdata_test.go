package dcmnet

import (
	"bytes"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestPDataReaderSinglePData(t *testing.T) {
	RegisterTestingT(t)

	data := bufpdu(PDUPresentationData, "data")
	data, finalPDU, err := readPDUs(data)

	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data"))
	Expect(finalPDU).To(BeNil())
}

func TestPDataReaderTwoPData(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData, "data1"),
		bufpdu(PDUPresentationData, "data2"))

	data, finalPDU, err := readPDUs(data)
	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data1data2"))
	Expect(finalPDU).To(BeNil())
}

func TestPDataReaderOnePDataThenAbort(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData, "data"),
		bufpdu(PDUAbort, "notdata"))

	data, abort, err := readPDUs(data)
	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data"))

	Expect(abort).ToNot(BeNil())
	Expect(abort.Type).To(Equal(PDUAbort))
}

func readPDUs(data bytes.Buffer) (buf bytes.Buffer, finalPDU *PDU, err error) {
	pduDecoder := NewPDUDecoder(&data)
	pdataReader := NewPDataReader(pduDecoder)

	output, err := ioutil.ReadAll(&pdataReader)
	if err != nil {
		return
	}

	return *bytes.NewBuffer(output), pdataReader.GetFinalPDU(), nil
}
