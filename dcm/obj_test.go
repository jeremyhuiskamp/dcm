package dcm

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestElementOrder(t *testing.T) {
	RegisterTestingT(t)
	o := Object{
			Elements: make(map[Tag]Element),
		}

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
