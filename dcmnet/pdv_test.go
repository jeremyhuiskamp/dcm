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
