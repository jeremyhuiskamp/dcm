package dcm

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestElementOrder(t *testing.T) {
	RegisterTestingT(t)
	o := NewObject()

	o.Put(SimpleElement{
		Tag:  PatientID,
		VR:   LO,
		Data: []byte("pid"),
	})
	o.Put(SimpleElement{
		Tag:  IssuerOfPatientID,
		VR:   LO,
		Data: []byte("issuer"),
	})
	o.Put(SimpleElement{
		Tag:  StudyInstanceUID,
		VR:   UI,
		Data: []byte("1.2.3"),
	})
	o.Put(SimpleElement{
		Tag:  TransferSyntaxUID,
		VR:   UI,
		Data: []byte(ImplicitVRLittleEndian.UID()),
	})

	var lastTag Tag

	o.ForEach(func(tag Tag, e Element) bool {
		Expect(lastTag).To(BeNumerically("<", tag))

		lastTag = tag

		return true
	})
}

func TestScan(t *testing.T) {
	RegisterTestingT(t)

	o := NewObject()
	o.Put(SimpleElement{
		Tag:  CommandField,
		VR:   US,
		Data: []byte{0x01, 0x00},
	})

	var cmd uint16
	Expect(o.Scan(CommandField, &cmd)).To(Succeed())
	Expect(cmd).To(Equal(uint16(1)))

	// not found:
	Expect(NewObject().Scan(CommandField, &cmd)).ToNot(Succeed())

	// not simple element:
	container := NewObject()
	container.Put(SequenceElement{
		Tag:     IssuerOfPatientIDQualifiersSequence,
		Objects: []Object{o},
	})
	Expect(container.Scan(IssuerOfPatientIDQualifiersSequence, &cmd)).
		ToNot(Succeed())
}
