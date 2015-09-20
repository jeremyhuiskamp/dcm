package dcmnet

import (
	"bytes"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func TestCommandVsData(t *testing.T) {
	RegisterTestingT(t)

	pdv := PDV{}

	Expect(pdv.GetType()).To(Equal(Data)) // Data is default

	pdv.SetType(Command)
	Expect(pdv.GetType()).To(Equal(Command))

	pdv.SetType(Data)
	Expect(pdv.GetType()).To(Equal(Data))
}

func TestLast(t *testing.T) {
	RegisterTestingT(t)

	pdv := PDV{}

	Expect(pdv.IsLast()).To(BeFalse()) // Not Last is default

	pdv.SetLast(true)
	Expect(pdv.IsLast()).To(BeTrue())

	pdv.SetLast(false)
}

func TestLastDoesntAffectCommand(t *testing.T) {
	RegisterTestingT(t)

	pdv := PDV{}

	Expect(pdv.GetType()).To(Equal(Data))
	pdv.SetLast(!pdv.IsLast())
	Expect(pdv.GetType()).To(Equal(Data))
	pdv.SetLast(!pdv.IsLast())
	Expect(pdv.GetType()).To(Equal(Data))

	pdv.SetType(Command)
	pdv.SetLast(!pdv.IsLast())
	Expect(pdv.GetType()).To(Equal(Command))
	pdv.SetLast(!pdv.IsLast())
	Expect(pdv.GetType()).To(Equal(Command))

	pdv.SetType(Data)
	pdv.SetLast(!pdv.IsLast())
	Expect(pdv.GetType()).To(Equal(Data))
	pdv.SetLast(!pdv.IsLast())
	Expect(pdv.GetType()).To(Equal(Data))
}

func TestCommandDoesntAffectLast(t *testing.T) {
	RegisterTestingT(t)

	pdv := PDV{}

	Expect(pdv.IsLast()).To(BeFalse())
	pdv.SetType(Command)
	Expect(pdv.IsLast()).To(BeFalse())
	pdv.SetType(Data)
	Expect(pdv.IsLast()).To(BeFalse())

	pdv.SetLast(true)
	pdv.SetType(Command)
	Expect(pdv.IsLast()).To(BeTrue())
	pdv.SetType(Data)
	Expect(pdv.IsLast()).To(BeTrue())

	pdv.SetLast(false)
	pdv.SetType(Command)
	Expect(pdv.IsLast()).To(BeFalse())
	pdv.SetType(Data)
	Expect(pdv.IsLast()).To(BeFalse())
}

func TestPDVReaderSinglePDUCommand(t *testing.T) {
	RegisterTestingT(t)

	f, err := ioutil.ReadFile("testdata/cecho_req_pdu.bin")
	Expect(err).To(BeNil())

	var buf bytes.Buffer
	buf.Write(f)
	Expect(buf.Len()).To(Equal(80))

	pdur := NewPDUDecoder(&buf)

	pdu, err := pdur.NextPDU()
	Expect(err).To(BeNil())
	Expect(pdu.Type).To(Equal(PDUPresentationData))
	Expect(pdu.Length).To(Equal(uint32(74)))

	pdv, err := NextPDV(pdu.Data)
	Expect(err).To(BeNil())
	Expect(pdv.Context).To(Equal(uint8(1)))
	Expect(pdv.GetType()).To(Equal(Command))
	Expect(pdv.IsLast()).To(BeTrue())
	Expect(pdv.Length).To(Equal(uint32(70)))

	pdvr := ReadPDVs(*pdv, *pdu, pdur)

	pdvbytes, err := ioutil.ReadAll(&pdvr)
	Expect(err).To(BeNil())
	Expect(pdvbytes).To(HaveLen(68))
}

func TestPDVReaderCommandAndTwoPDVs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdu(PDUPresentationData,
			bufpdv(1, Command, true, "command"),
			bufpdv(1, Data, false, "data1\n")),
		bufpdu(PDUPresentationData,
			bufpdv(1, Data, true, "data2")))

	pdur := NewPDUDecoder(&data)

	pdu, err := pdur.NextPDU()
	Expect(err).To(BeNil())
	Expect(pdu.Type).To(Equal(PDUPresentationData))

	pdv, err := NextPDV(pdu.Data)
	Expect(err).To(BeNil())
	Expect(pdv.Context).To(Equal(uint8(1)))
	Expect(pdv.GetType()).To(Equal(Command))
	Expect(pdv.IsLast()).To(BeTrue())
	Expect(toString(pdv.Data)).To(Equal("command"))

	pdv, err = NextPDV(pdu.Data)
	Expect(err).To(BeNil())
	Expect(pdu.Type).To(Equal(PDUPresentationData))
	Expect(pdv.Context).To(Equal(uint8(1)))
	Expect(pdv.GetType()).To(Equal(Data))
	Expect(pdv.IsLast()).To(BeFalse())

	pdvr := ReadPDVs(*pdv, *pdu, pdur)
	Expect(toString(&pdvr)).To(Equal("data1\ndata2"))
}
