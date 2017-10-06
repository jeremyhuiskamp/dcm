package dcmnet

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestPDataReaderSinglePData(t *testing.T) {
	data := bufpdu(PDUPresentationData, "data")
	data, finalPDU := readPDUs(t, data)

	if got := string(data.Bytes()); got != "data" {
		t.Errorf("unexpected data: %q", data)
	}
	if finalPDU != nil {
		t.Error("unexpected final pdu")
	}
}

func TestPDataReaderTwoPData(t *testing.T) {
	data := bufcat(
		bufpdu(PDUPresentationData, "data1"),
		bufpdu(PDUPresentationData, "data2"))

	data, finalPDU := readPDUs(t, data)

	if got := string(data.Bytes()); got != "data1data2" {
		t.Errorf("unexpected data: %q", got)
	}
	if finalPDU != nil {
		t.Error("unexpected final pdu")
	}
}

func TestPDataReaderOnePDataThenAbort(t *testing.T) {
	data := bufcat(
		bufpdu(PDUPresentationData, "data"),
		bufpdu(PDUAbort, "notdata"))

	data, abort := readPDUs(t, data)

	if got := string(data.Bytes()); got != "data" {
		t.Errorf("unexpected data: %q", data)
	}
	if abort == nil {
		t.Error("didn't get expected abort")
	} else if abort.Type != PDUAbort {
		t.Error("unexpected final pdu type: %d", abort.Type)
	}
}

func readPDUs(t *testing.T, data bytes.Buffer) (bytes.Buffer, *PDU) {
	pduDecoder := NewPDUDecoder(&data)
	pdataReader := NewPDataReader(pduDecoder)

	output, err := ioutil.ReadAll(&pdataReader)
	if err != nil {
		t.Fatalf("unable to read pdata: %s", err)
	}

	return *bytes.NewBuffer(output), pdataReader.GetFinalPDU()
}
