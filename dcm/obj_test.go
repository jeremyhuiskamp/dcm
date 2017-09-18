package dcm

import (
	"testing"
)

func TestElementOrder(t *testing.T) {
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
		if lastTag >= tag {
			t.Fatalf("tags not sorted (%s>=%s)", lastTag, tag)
		}
		lastTag = tag
		return true
	})
}

func TestScan(t *testing.T) {
	o := NewObject()
	o.Put(SimpleElement{
		Tag:  CommandField,
		VR:   US,
		Data: []byte{0x01, 0x00},
	})

	var cmd uint16
	if err := o.Scan(CommandField, &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd != uint16(1) {
		t.Fatal("unexpected command field: %d", cmd)
	}

	if err := NewObject().Scan(CommandField, &cmd); err == nil {
		t.Fatal("expected error for non-existent field")
	}

	container := NewObject()
	container.Put(SequenceElement{
		Tag:     IssuerOfPatientIDQualifiersSequence,
		Objects: []Object{o},
	})
	if err := container.Scan(IssuerOfPatientIDQualifiersSequence, &cmd); err == nil {
		t.Fatal("expected error trying to scan a sequence")
	}
}
