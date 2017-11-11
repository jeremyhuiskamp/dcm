package dcmnet

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jeremyhuiskamp/dcm/dcm"
)

func TestParseAssocRQCEcho(t *testing.T) {
	exp := AssociateRQAC{
		ProtocolVersion:    1,
		CalledAE:           "RIALTO",
		CallingAE:          "DCMECHO",
		ApplicationContext: "1.2.840.10008.3.1.1.1",
		PresentationContexts: []PresentationContext{{
			ID:             PCID(1),
			Result:         PCAcceptance,
			AbstractSyntax: "1.2.840.10008.1.1",
			TransferSyntaxes: []dcm.TransferSyntax{
				dcm.ImplicitVRLittleEndian,
			},
		}},
		MaxPDULength:           16384,
		ImplementationClassUID: "1.2.40.0.13.1.1",
		ImplementationVersion:  "dcm4che-2.0",
		// hmm, not a good test, since these are the defaults:
		MaxOperationsInvoked:   0,
		MaxOperationsPerformed: 0,
	}
	got := readAssocRQAC(t, "testdata/assocrq_cecho.bin", PDUAssociateRQ)

	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected\n%#v\ngot\n%#v", exp, got)
	}
}

func TestParseAssocACCEcho(t *testing.T) {
	exp := AssociateRQAC{
		ProtocolVersion:    1,
		CalledAE:           "RIALTO",
		CallingAE:          "DCMECHO",
		ApplicationContext: "1.2.840.10008.3.1.1.1",
		PresentationContexts: []PresentationContext{{
			ID:             1,
			Result:         PCAcceptance,
			AbstractSyntax: "",
			TransferSyntaxes: []dcm.TransferSyntax{
				dcm.ImplicitVRLittleEndian,
			},
		}},
		MaxPDULength:           16384,
		ImplementationClassUID: "1.2.40.0.13.1.1",
		ImplementationVersion:  "dcm4che-2.0",
		// hmm, not a good test, since these are the defaults:
		MaxOperationsInvoked:   0,
		MaxOperationsPerformed: 0,
	}
	got := readAssocRQAC(t, "testdata/assocac_cecho.bin", PDUAssociateAC)

	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected\n%#v\ngot\n%#v", exp, got)
	}
}

func readAssocRQAC(t *testing.T, file string, pduType PDUType) AssociateRQAC {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	buf.Write(b)
	buf.WriteString("don't read me bro")

	pduDecoder := NewPDUDecoder(&buf)
	pdu, err := pduDecoder.NextPDU()
	if err != nil {
		t.Fatal(err)
	}
	if pdu == nil {
		t.Fatal("no pdu found")
	}

	if pduType != pdu.Type {
		t.Fatalf("unexpected pdu type, expected %s, got %s",
			pduType, pdu.Type)
	}
	var rqac AssociateRQAC
	rqac.Read(pdu.Data)

	if "don't read me bro" != buf.String() {
		t.Fatal("didn't stop reading at the right place")
	}

	return rqac
}
