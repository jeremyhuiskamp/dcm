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

func TestPDataReaderSinglePData(t *testing.T) {
	RegisterTestingT(t)

	data := bufpdu(PDUPresentationData, "data")
	data, err, finalPDU := readPDUs(data)

	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data"))
	Expect(finalPDU).To(BeNil())
}

func TestPDataReaderTwoPData(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData, "data1"),
		bufpdu(PDUPresentationData, "data2"))

	data, err, finalPDU := readPDUs(data)
	Expect(err).To(BeNil())
	Expect(string(data.Bytes())).To(Equal("data1data2"))
	Expect(finalPDU).To(BeNil())
}

func TestPDataReaderOnePDataThenAbort(t *testing.T) {
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

func TestPDataWriter(t *testing.T) {
	RegisterTestingT(t)

	input := "abcdefghijklmnopqrstuvwxyz"

	for pdulen := uint32(1); pdulen < 10; pdulen++ {
		for inputlen := 0; inputlen <= len(input); inputlen++ {
			curinput := input[:inputlen]

			data := new(bytes.Buffer)
			pdus := NewPDUEncoder(data)
			pdata := NewPDataWriter(pdus, pdulen)

			pdata.Write([]byte(curinput))
			pdata.Close()

			output := new(bytes.Buffer)
			for data.Len() > 0 {
				typ, content, err := getpdu(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(typ).To(Equal(PDUPresentationData))
				output.Write(content.Bytes())
			}

			Expect(data.Len()).To(Equal(0))
			Expect(output.String()).To(Equal(curinput),
				"pdulen=%d, inputlen=%d", pdulen, inputlen)
		}
	}
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
