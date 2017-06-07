package dcmio

import (
	"os"
	"strings"
	"testing"

	"github.com/jeremyhuiskamp/dcm/dcm"
	. "github.com/onsi/gomega"
)

func TestParseCEchoReqCmd(t *testing.T) {
	RegisterTestingT(t)

	f, err := os.Open("testdata/cecho_req_cmd.bin")
	Expect(err).To(BeNil())

	p := NewStreamParser(f, dcm.ImplicitVRLittleEndian)
	obj, err := Build(p)
	Expect(err).To(BeNil())
	Expect(obj).ToNot(BeNil())

	t.Logf("Parsed:\n%s\n", obj)

	Expect(getString(dcm.AffectedSOPClassUID, obj)).To(Equal("1.2.840.10008.1.1"))

	Expect(getInt(dcm.CommandField, obj, dcm.ExplicitVRLittleEndian)).To(Equal(0x30))

	Expect(getInt(dcm.MessageID, obj, dcm.ExplicitVRLittleEndian)).To(Equal(0x01))

	Expect(
		getInt(dcm.CommandDataSetType, obj, dcm.ExplicitVRLittleEndian),
	).To(Equal(0x0101))
}

// TODO these should be defined in dcm package!

func getString(tag dcm.Tag, obj dcm.Object) string {
	el := obj.Get(tag)
	if el == nil {
		return ""
	}

	if se, ok := (*el).(dcm.SimpleElement); ok {
		return strings.TrimRight(string(se.Data), "\x00")
	}

	return ""
}

func getInt(tag dcm.Tag, obj dcm.Object, ts dcm.TransferSyntax) int {
	el := obj.Get(tag)
	if el == nil {
		return -1
	}

	if se, ok := (*el).(dcm.SimpleElement); ok {
		switch se.VR.Name {
		case dcm.US.Name:
			return int(ts.ByteOrder().Uint16(se.Data))
		}
	}

	return -1
}
