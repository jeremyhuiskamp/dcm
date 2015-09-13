package dcmio

import (
    "github.com/kamper/dcm/dcm"
    "testing"
    "bytes"
)

var (
    part10header = []byte {
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
        0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
         'D', 'I', 'C', 'M',
    }
)

func combine(bs ... []byte) []byte {
    var combined []byte
    for i := 0; i < len(bs); i++ {
        combined = append(combined, bs[i]...)
    }
    return combined
}

func assertElement(t *testing.T, el *Tag, offset uint64, tag dcm.Tag,
                vr *dcm.VR, valueOffset uint64, valueLength int32) {
    if el == nil {
        t.Fatal("element == nil")
    }

    if el.Offset != offset {
        t.Fatal("Wrong offset:", el.Offset, "(expected:", offset, ")")
    }

	if el.Tag != tag {
		t.Fatalf("Wrong tag: %s (expected %s", el.Tag, tag)
	}

    // TODO: this blows up for nil
    if *el.VR != *vr {
        t.Fatalf("Wrong vr for %s: %v (expected %v)", el.Tag, el.VR, vr)
    }

    if el.ValueOffset != valueOffset {
        t.Fatalf("Wrong value offset: %d (expected %d)",
            el.ValueOffset, valueOffset)
    }

    if el.ValueLength != valueLength {
        t.Fatalf("Wrong value length: %d (expected: %d)",
            el.ValueLength, valueLength)
    }

    t.Log("Valid element:", el)
}

func assertNextElement(t *testing.T, p Parser, offset uint64,
                    tag dcm.Tag, vr *dcm.VR, valueOffset uint64,
                    valueLength int32) {
    el, err := p.NextTag()

    if err != nil {
        t.Fatal("unexpected error:", err)
    }

    assertElement(t, el, offset, tag, vr, valueOffset, valueLength)
}

func assertNoMoreElements(t *testing.T, p Parser) {
    el, err := p.NextTag()

    if err != nil {
        t.Fatal("unexpected error:", err)
    }

    if el != nil {
        t.Fatal("unexpected element:", el)
    }
}

// part 10 file, irvle,
// PatientID = pid
func TestPart10Ivrle(t *testing.T) {
    p, err := NewFileParser(bytes.NewBuffer(combine(part10header,
        []byte {
            0x02,0x00,0x00,0x00,0x55,0x4C,0x04,0x00,
            0x1A,0x00,0x00,0x00,0x02,0x00,0x10,0x00,
            0x55,0x49,0x12,0x00,0x31,0x2E,0x32,0x2E,
            0x38,0x34,0x30,0x2E,0x31,0x30,0x30,0x30,
            0x38,0x2E,0x31,0x2E,0x32,0x00,0x10,0x00,
            0x20,0x00,0x04,0x00,0x00,0x00,0x70,0x69,
            0x64,0x20,
        })))

    if err != nil {
        t.Fatal("unexpected error:", err)
    }

    // TODO: assert values
    assertNextElement(t, p, 132, dcm.FileMetaInformationGroupLength, &dcm.UL, 140,  4)
    assertNextElement(t, p, 144, dcm.TransferSyntaxUID,              &dcm.UI, 152, 18)
    assertNextElement(t, p, 170, dcm.PatientID,                      &dcm.LO, 178,  4)
    assertNoMoreElements(t, p)
}
