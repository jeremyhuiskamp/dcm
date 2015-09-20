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

	data := pduX(PDUPresentationData, *bytes.NewBuffer([]byte("data")))
	data, err := readPDUs(data)

	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data"))
}

func TestPDUReaderTwoPData(t *testing.T) {
	RegisterTestingT(t)

	data := combine(
		pduX(PDUPresentationData, *bytes.NewBuffer([]byte("data1"))),
		pduX(PDUPresentationData, *bytes.NewBuffer([]byte("data2"))))

	data, err := readPDUs(data)
	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data1data2"))
}

func TestPDUReaderOnePDataThenAbort(t *testing.T) {
	RegisterTestingT(t)

	data := combine(
		pduX(PDUPresentationData, *bytes.NewBuffer([]byte("data"))),
		pduX(PDUAbort, *bytes.NewBuffer([]byte("notdata"))))

	data, err := readPDUs(data)
	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data"))
}

func readPDUs(data bytes.Buffer) (buf bytes.Buffer, err error) {
	pduDecoder := NewPDUDecoder(&data)
	pdataReader := NewPDataReader(pduDecoder)

	output, err := ioutil.ReadAll(&pdataReader)
	if err != nil {
		return
	}

	return *bytes.NewBuffer(output), nil
}
