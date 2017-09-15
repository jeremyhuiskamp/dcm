package dcmnet

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestReadOnePDU(t *testing.T) {
	decoder := pduDecoder(bufpdu(0x01, "hai!"))
	expectNextPDU(t, decoder, 0x01, "hai!")
}

func TestReadTwoPDUs(t *testing.T) {
	decoder := pduDecoder(bufpdu(0x01, "one"), bufpdu(0x02, "two"))
	expectNextPDU(t, decoder, 0x01, "one")
	expectNextPDU(t, decoder, 0x02, "two")
}

func TestReadPDUNilForEOF(t *testing.T) {
	decoder := pduDecoder()
	pdu, err := decoder.NextPDU()
	if pdu != nil {
		t.Error("unexpected pdu")
	}
	if err != nil {
		t.Error(err)
	}
}

func TestReadDrainFirstPDUWhenAskedForSecond(t *testing.T) {
	decoder := pduDecoder(bufpdu(0x01, "one"), bufpdu(0x02, "two"))
	// not reading value...
	decoder.NextPDU()
	expectNextPDU(t, decoder, 0x02, "two")
}

func TestWriteOnePDU(t *testing.T) {
	data := new(bytes.Buffer)
	encoder := NewPDUEncoder(data)
	encoder.NextPDU(toPDU(PDUPresentationData, "data"))

	content := expectPDU(t, data, PDUPresentationData)
	if got := toString(content); got != "data" {
		t.Fatalf("unexpected content: %q", got)
	}

	if data.Len() != 0 {
		t.Fatalf("unexpected data left over: %q", data.String())
	}
}

func TestWriteTwoPDUs(t *testing.T) {
	data := new(bytes.Buffer)
	encoder := NewPDUEncoder(data)
	encoder.NextPDU(toPDU(PDUPresentationData, "data1"))
	encoder.NextPDU(toPDU(PDUType(25), "data2"))

	content := expectPDU(t, data, PDUPresentationData)
	if got := toString(content); got != "data1" {
		t.Fatalf("unexpected content: %q", got)
	}

	content = expectPDU(t, data, PDUType(25))
	if got := toString(content); got != "data2" {
		t.Fatalf("unexpected content: %q", got)
	}

	if data.Len() != 0 {
		t.Fatalf("unexpected data left over: %q", data.String())
	}
}

func TestWriteEmptyPDU(t *testing.T) {
	data := new(bytes.Buffer)
	encoder := NewPDUEncoder(data)
	encoder.NextPDU(toPDU(PDUPresentationData, ""))
	encoder.NextPDU(toPDU(PDUType(25), ""))

	content := expectPDU(t, data, PDUPresentationData)
	if got := toString(content); got != "" {
		t.Fatalf("unexpected content: %q", got)
	}

	content = expectPDU(t, data, PDUType(25))
	if got := toString(content); got != "" {
		t.Fatalf("unexpected content: %q", got)
	}

	if data.Len() != 0 {
		t.Fatalf("unexpected data left over: %q", data.String())
	}
}

func pduDecoder(pdus ...interface{}) PDUDecoder {
	b := bufcat(pdus...)
	return NewPDUDecoder(&b)
}

func expectNextPDU(t *testing.T, decoder PDUDecoder, pduType PDUType, value string) {
	pdu, err := decoder.NextPDU()

	if err != nil {
		t.Fatalf("error reading next pdu: %s", err)
	}
	if pdu == nil {
		t.Fatal("didn't get expected pdu")
	}

	if pdu.Type != pduType {
		t.Fatalf("unexpected pdu type: %s", pdu.Type)
	}
	if pdu.Length != uint32(len(value)) {
		t.Fatalf("unexpected pdu length: %d", pdu.Length)
	}

	actualValue, err := ioutil.ReadAll(pdu.Data)
	if err != nil {
		t.Fatalf("error reading pdu data: %s", err)
	}
	if got := string(actualValue); got != value {
		t.Fatalf("unexpected pdu value: %q", got)
	}
}

func expectPDU(t *testing.T, in *bytes.Buffer, expType PDUType) bytes.Buffer {
	gotType, content, err := getpdu(in)
	if err != nil {
		t.Fatal(err)
	}
	if expType != gotType {
		t.Fatalf("expected %s, got %s", expType, gotType)
	}
	return content
}
