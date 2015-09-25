package dcmnet

import (
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

func TestPDUNilForEOF(t *testing.T) {
	RegisterTestingT(t)

	decoder := pduDecoder()
	pdu, err := decoder.NextPDU()
	Expect(pdu).To(BeNil())
	Expect(err).ToNot(HaveOccurred())
}

func TestDrainFirstPDUWhenAskedForSecond(t *testing.T) {
	RegisterTestingT(t)

	decoder := pduDecoder(bufpdu(0x01, "one"), bufpdu(0x02, "two"))
	// not reading value...
	decoder.NextPDU()
	expectNextPDU(decoder, 0x02, "two")
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
