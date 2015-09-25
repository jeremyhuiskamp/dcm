package dcmnet

import (
	. "github.com/onsi/gomega"
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

func TestReadOnePDV(t *testing.T) {
	RegisterTestingT(t)

	decoder := pdvDecoder(bufpdv(1, Command, true, "command"))
	expectNextPDV(decoder, 1, Command, true, "command")
	expectNoMorePDVs(decoder)
}

func TestReadMultiplePDVs(t *testing.T) {
	RegisterTestingT(t)

	decoder := pdvDecoder(
		bufpdv(1, Command, true, "command"),
		bufpdv(2, Data, false, "data1"),
		bufpdv(3, Data, true, "data2"))

	expectNextPDV(decoder, 1, Command, true, "command")
	expectNextPDV(decoder, 2, Data, false, "data1")
	expectNextPDV(decoder, 3, Data, true, "data2")
	expectNoMorePDVs(decoder)
}

func TestNoPDVs(t *testing.T) {
	RegisterTestingT(t)

	decoder := pdvDecoder()
	expectNoMorePDVs(decoder)
}

func TestDrainFirstPDVWhenAskedForSecond(t *testing.T) {
	RegisterTestingT(t)

	decoder := pdvDecoder(
		bufpdv(1, Command, true, "command"),
		bufpdv(2, Data, false, "data"))
	// not reading value...
	decoder.NextPDV()
	expectNextPDV(decoder, 2, Data, false, "data")
	expectNoMorePDVs(decoder)
}

func expectNextPDV(decoder PDVDecoder, context uint8, tipe PDVType, last bool,
	content string) {

	pdv, err := decoder.NextPDV()
	Expect(err).To(BeNil())
	Expect(pdv).ToNot(BeNil())

	Expect(pdv.Context).To(Equal(context))
	Expect(pdv.GetType()).To(Equal(tipe))
	Expect(pdv.IsLast()).To(Equal(last))
	Expect(toString(pdv.Data)).To(Equal(content))
}

func expectNoMorePDVs(decoder PDVDecoder) {
	pdv, err := decoder.NextPDV()
	Expect(err).To(BeNil())
	Expect(pdv).To(BeNil())
}

func pdvDecoder(pdvs ...interface{}) PDVDecoder {
	b := bufcat(pdvs...)
	return NewPDVDecoder(&b)
}
