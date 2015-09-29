package dcmnet

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestCommandVsData(t *testing.T) {
	RegisterTestingT(t)

	var flags PDVFlags

	Expect(flags.GetType()).To(Equal(Data)) // Data is default

	flags.SetType(Command)
	Expect(flags.GetType()).To(Equal(Command))

	flags.SetType(Data)
	Expect(flags.GetType()).To(Equal(Data))
}

func TestLast(t *testing.T) {
	RegisterTestingT(t)

	var flags PDVFlags

	Expect(flags.IsLast()).To(BeFalse()) // Not Last is default

	flags.SetLast(true)
	Expect(flags.IsLast()).To(BeTrue())

	flags.SetLast(false)
}

func TestLastDoesntAffectCommand(t *testing.T) {
	RegisterTestingT(t)

	var flags PDVFlags

	Expect(flags.GetType()).To(Equal(Data))
	flags.SetLast(!flags.IsLast())
	Expect(flags.GetType()).To(Equal(Data))
	flags.SetLast(!flags.IsLast())
	Expect(flags.GetType()).To(Equal(Data))

	flags.SetType(Command)
	flags.SetLast(!flags.IsLast())
	Expect(flags.GetType()).To(Equal(Command))
	flags.SetLast(!flags.IsLast())
	Expect(flags.GetType()).To(Equal(Command))

	flags.SetType(Data)
	flags.SetLast(!flags.IsLast())
	Expect(flags.GetType()).To(Equal(Data))
	flags.SetLast(!flags.IsLast())
	Expect(flags.GetType()).To(Equal(Data))
}

func TestCommandDoesntAffectLast(t *testing.T) {
	RegisterTestingT(t)

	var flags PDVFlags

	Expect(flags.IsLast()).To(BeFalse())
	flags.SetType(Command)
	Expect(flags.IsLast()).To(BeFalse())
	flags.SetType(Data)
	Expect(flags.IsLast()).To(BeFalse())

	flags.SetLast(true)
	flags.SetType(Command)
	Expect(flags.IsLast()).To(BeTrue())
	flags.SetType(Data)
	Expect(flags.IsLast()).To(BeTrue())

	flags.SetLast(false)
	flags.SetType(Command)
	Expect(flags.IsLast()).To(BeFalse())
	flags.SetType(Data)
	Expect(flags.IsLast()).To(BeFalse())
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

func expectNextPDV(decoder PDVDecoder, context PCID, tipe PDVType, last bool,
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
