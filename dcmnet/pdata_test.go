package dcmnet

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestPDUReaderSinglePData(t *testing.T) {
	RegisterTestingT(t)

	data := bufpdu(PDUPresentationData, "data")
	data, err, finalPDU := readPDUs(data)

	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data"))
	Expect(finalPDU).To(BeNil())
}

func TestPDUReaderTwoPData(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData, "data1"),
		bufpdu(PDUPresentationData, "data2"))

	data, err, finalPDU := readPDUs(data)
	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data1data2"))
	Expect(finalPDU).To(BeNil())
}

func TestPDUReaderOnePDataThenAbort(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData, "data"),
		bufpdu(PDUAbort, "notdata"))

	data, err, abort := readPDUs(data)
	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data"))

	Expect(abort).ToNot(BeNil())
	Expect(abort.Type).To(Equal(PDUAbort))
}

func readPDUs(data bytes.Buffer) (buf bytes.Buffer, err error, finalPDU *PDU) {
	pduDecoder := NewPDUDecoder(&data)
	pdataReader := NewPDataReader(pduDecoder)

	output, err := ioutil.ReadAll(&pdataReader)
	if err != nil {
		return
	}

	return *bytes.NewBuffer(output), nil, pdataReader.GetFinalPDU()
}
