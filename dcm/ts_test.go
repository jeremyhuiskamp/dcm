package dcm

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestIVRLE(t *testing.T) {
	RegisterTestingT(t)

	Expect(ImplicitVRLittleEndian.VR()).NotTo(Equal(Explicit))
	Expect(GetTransferSyntax("1.2.840.10008.1.2").VR()).To(Equal(Implicit))
}
