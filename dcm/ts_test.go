package dcm

import (
	"testing"
)

func TestIVRLE(t *testing.T) {
	if ImplicitVRLittleEndian.VR() == Explicit {
		t.Fatal("Should be implicit!")
	}

	if GetTransferSyntax("1.2.840.10008.1.2").VR() == Explicit {
		t.Fatal("Should be implicit!")
	}
}
