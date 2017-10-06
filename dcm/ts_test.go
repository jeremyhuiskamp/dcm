package dcm

import "testing"

func TestIVRLE(t *testing.T) {
	for _, test := range []struct {
		ts TransferSyntax
		vr Explicitness
	}{
		{ImplicitVRLittleEndian, Implicit},
		{GetTransferSyntax("1.2.840.10008.1.2"), Implicit},
	} {
		if got := test.ts.VR(); got != test.vr {
			t.Errorf("unexpected vr %s for syntax %s", got, test.ts)
		}
	}
}
