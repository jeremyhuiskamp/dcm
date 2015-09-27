package dcmnet

import (
	"bytes"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func TestReadOnePDU(t *testing.T) {
	RegisterTestingT(t)

	decoder := pduDecoder(bufpdu(0x01, "hai!"))
	expectNextPDU(decoder, 0x01, "hai!")
}

func TestReadTwoPDUs(t *testing.T) {
	RegisterTestingT(t)

	decoder := pduDecoder(bufpdu(0x01, "one"), bufpdu(0x02, "two"))
	expectNextPDU(decoder, 0x01, "one")
	expectNextPDU(decoder, 0x02, "two")
}

func TestReadPDUNilForEOF(t *testing.T) {
	RegisterTestingT(t)

	decoder := pduDecoder()
	pdu, err := decoder.NextPDU()
	Expect(pdu).To(BeNil())
	Expect(err).ToNot(HaveOccurred())
}

func TestReadDrainFirstPDUWhenAskedForSecond(t *testing.T) {
	RegisterTestingT(t)

	decoder := pduDecoder(bufpdu(0x01, "one"), bufpdu(0x02, "two"))
	// not reading value...
	decoder.NextPDU()
	expectNextPDU(decoder, 0x02, "two")
}

func TestWriteOnePDU(t *testing.T) {
	RegisterTestingT(t)

	data := new(bytes.Buffer)
	encoder := NewPDUEncoder(data)
	encoder.NextPDU(toPDU(PDUPresentationData, "data"))

	typ, content, err := getpdu(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(typ).To(Equal(PDUPresentationData))
	Expect(toString(content)).To(Equal("data"))

	Expect(data.Len()).To(Equal(0))
}

func TestWriteTwoPDUs(t *testing.T) {
	RegisterTestingT(t)

	data := new(bytes.Buffer)
	encoder := NewPDUEncoder(data)
	encoder.NextPDU(toPDU(PDUPresentationData, "data1"))
	encoder.NextPDU(toPDU(PDUType(25), "data2"))

	typ, content, err := getpdu(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(typ).To(Equal(PDUPresentationData))
	Expect(toString(content)).To(Equal("data1"))

	typ, content, err = getpdu(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(typ).To(Equal(PDUType(25)))
	Expect(toString(content)).To(Equal("data2"))

	Expect(data.Len()).To(Equal(0))
}

func TestWriteEmptyPDU(t *testing.T) {
	RegisterTestingT(t)

	data := new(bytes.Buffer)
	encoder := NewPDUEncoder(data)
	encoder.NextPDU(toPDU(PDUPresentationData, ""))
	encoder.NextPDU(toPDU(PDUType(25), ""))

	typ, content, err := getpdu(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(typ).To(Equal(PDUPresentationData))
	Expect(toString(content)).To(Equal(""))

	typ, content, err = getpdu(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(typ).To(Equal(PDUType(25)))
	Expect(toString(content)).To(Equal(""))

	Expect(data.Len()).To(Equal(0))
}

func pduDecoder(pdus ...interface{}) PDUDecoder {
	b := bufcat(pdus...)
	return NewPDUDecoder(&b)
}

func expectNextPDU(decoder PDUDecoder, pduType PDUType, value string) {
	pdu, err := decoder.NextPDU()

	Expect(err).ToNot(HaveOccurred())
	Expect(pdu).ToNot(BeNil())

	Expect(pdu.Type).To(Equal(pduType))
	Expect(pdu.Length).To(Equal(uint32(len(value))))

	actualvalue, err := ioutil.ReadAll(pdu.Data)
	Expect(err).ToNot(HaveOccurred())
	Expect(string(actualvalue)).To(Equal(value))
}
