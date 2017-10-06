package dcmio

import (
	"os"
	"strings"
	"testing"

	"github.com/jeremyhuiskamp/dcm/dcm"
)

func TestParseCEchoReqCmd(t *testing.T) {
	f, err := os.Open("testdata/cecho_req_cmd.bin")
	if err != nil {
		t.Fatal(err)
	}

	p := NewStreamParser(f, dcm.ImplicitVRLittleEndian)
	obj, err := Build(p)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Parsed:\n%s\n", obj)

	if got := getString(dcm.AffectedSOPClassUID, obj); got != "1.2.840.10008.1.1" {
		t.Errorf("unexpected affected sop class uid: %q", got)
	}

	for tag, value := range map[dcm.Tag]int{
		dcm.CommandField:       0x30,
		dcm.MessageID:          0x01,
		dcm.CommandDataSetType: 0x0101,
	} {
		if got := getInt(tag, obj, dcm.ExplicitVRLittleEndian); got != value {
			t.Errorf("unexpected value %+v for tag %s", got, tag)
		}
	}
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
