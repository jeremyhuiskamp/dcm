package dcmnet

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jeremyhuiskamp/dcm/dcm"
)

// TODO: tests for unexpected errors

func TestReadSingleMessageElementSinglePDV(t *testing.T) {
	data := bufcat(bufpdv(1, Data, true, "data"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElement(t, med, 1, Data, "data")
	expectNoMoreMessageElements(t, med)
}

func TestReadSingleMessageElementTwoPDVs(t *testing.T) {
	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(1, Data, true, "data2"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElement(t, med, 1, Data, "data1data2")
	expectNoMoreMessageElements(t, med)
}

func TestReadTwoMessageElements(t *testing.T) {
	data := bufcat(
		bufpdv(1, Data, true, "data"),
		bufpdv(2, Command, true, "command"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElement(t, med, 1, Data, "data")
	expectMessageElement(t, med, 2, Command, "command")
	expectNoMoreMessageElements(t, med)
}

func TestReadDrainMessageElement(t *testing.T) {
	data := bufcat(
		bufpdv(1, Data, true, "data"),
		bufpdv(2, Command, true, "command"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	// don't read contents...
	_ = expectNextElement(t, med)

	expectMessageElement(t, med, 2, Command, "command")
	expectNoMoreMessageElements(t, med)
}

func TestReadMessageElementUnexpectedPDVType(t *testing.T) {
	data := bufcat(
		bufpdv(1, Data, false, "data"),
		bufpdv(1, Command, true, "command"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElementError(t, med, "unexpected type")
}

func TestReadMessageElementUnexpectedPresentationContext(t *testing.T) {
	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(2, Data, true, "data2"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElementError(t, med, "unexpected presentation context")
}

func TestReadMessageElementUnexpectedMissingPDV(t *testing.T) {
	data := bufcat(
		bufpdv(1, Command, false, "data1"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	msg := expectNextElement(t, med)
	_, err := ioutil.ReadAll(msg.Data)
	if io.ErrUnexpectedEOF != err {
		t.Errorf("expected unexpected eof but got %q", err)
	}
}

func TestReadMessageElementWithNoPDVs(t *testing.T) {
	data := bufcat()
	med := NewMessageElementDecoder(NewPDVDecoder(&data))
	expectNoMoreMessageElements(t, med)
}

func TestWriteOneMessageElement(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyz0123456789"

	// hits edge cases in pdu lengths
	// nb: a pdu that can't contain any data after the pdv header is illegal:
	for pdulen := uint32(pdvHeaderLen + 1); pdulen < 20; pdulen++ {
		for inputlen := 0; inputlen <= len(input); inputlen++ {
			curinput := input[:inputlen]

			data := new(bytes.Buffer)
			pdus := NewPDUEncoder(data)
			msgs := NewMessageElementEncoder(pdus, pdulen)

			msgs.NextMessageElement(MessageElement{
				Context: 1,
				Type:    Command,
				Data:    toBufferP(curinput),
			})

			output := new(bytes.Buffer)
			last := false
			for data.Len() > 0 && !last {
				pduData := expectPDU(t, data, PDUPresentationData)
				pdvLast, pdvData := expectPDV(t, &pduData, PCID(1), Command)
				output.Write(pdvData.Bytes())
				last = pdvLast
			}

			if data.Len() != 0 {
				t.Fatal("expected no data left in pdu")
			}

			if (len(curinput) == 0) == last {
				// TODO: explain what this means in error message
				t.Fatalf("last: %t, len(curinput): %d",
					last, len(curinput))
			}

			if curinput != output.String() {
				t.Fatalf("expected output %q but got %q, "+
					"where pdulen=%d and inputlen=%d",
					curinput, output, pdulen, inputlen)
			}
		}
	}
}

func TestWriteMultipleMessageElements(t *testing.T) {
	inputs := []string{
		"short",
		strings.Repeat("long", 5),
		strings.Repeat("reallylong", 30),
	}

	data := new(bytes.Buffer)
	pdus := NewPDUEncoder(data)
	msgs := NewMessageElementEncoder(pdus, 30)

	for context, value := range inputs {
		msgs.NextMessageElement(MessageElement{
			Context: PCID(context),
			Type:    Command,
			Data:    toBufferP("command" + value),
		})
		msgs.NextMessageElement(MessageElement{
			Context: PCID(context),
			Type:    Data,
			Data:    toBufferP("data" + value),
		})
	}

	expectMessageElement := func(context PCID, typ PDVType) string {
		last := false
		output := new(bytes.Buffer)

		for !last {
			pduData := expectPDU(t, data, PDUPresentationData)
			pdvLast, pdvData := expectPDV(t, &pduData, context, typ)
			output.Write(pdvData.Bytes())
			last = pdvLast
		}

		return output.String()
	}

	for context, value := range inputs {
		expCmdValue := "command" + value
		gotCmdValue := expectMessageElement(PCID(context), Command)
		if expCmdValue != gotCmdValue {
			t.Fatalf("expected %q, got %q", expCmdValue, gotCmdValue)
		}

		expDataValue := "data" + value
		gotDataValue := expectMessageElement(PCID(context), Data)
		if expDataValue != gotDataValue {
			t.Fatalf("expected %q, got %q", expDataValue, gotDataValue)
		}
	}

	if data.Len() != 0 {
		t.Fatalf("expected no more data but got %q", data)
	}
}

func TestReadNoMessages(t *testing.T) {
	var data bytes.Buffer
	var pcs PresentationContexts
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	msg, err := md.NextMessage()
	if msg != nil || err != nil {
		t.Fatal("expected neither msg (%V) nor error (%q)", msg, err)
	}
}

func TestReadUnexpectedData(t *testing.T) {
	data := bufpdv(1, Data, true, "xxxx")
	var pcs PresentationContexts
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	_, err := md.NextMessage()
	if err == nil || !strings.Contains(err.Error(), "Expected a command message") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestReadUnexpectedPCID(t *testing.T) {
	data := bufpdv(1, Command, true, "xxx")
	var pcs PresentationContexts
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	_, err := md.NextMessage()
	if err == nil || !strings.Contains(err.Error(), "Unrecognized presentation context") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestExpectDatasetButFindNone(t *testing.T) {
	pcs := PresentationContexts{
		Requested: []PresentationContext{{
			ID:             1,
			AbstractSyntax: "??",
		}},
		Accepted: []PresentationContext{{
			ID:     1,
			Result: PCAcceptance,
			TransferSyntaxes: []dcm.TransferSyntax{
				dcm.ImplicitVRLittleEndian,
			},
		}},
	}

	data := bufpdv(1, Command, true,
		// data set type element indicating that data should be expected:
		[]byte{0x00, 0x00, 0x00, 0x08, 0x02, 0x00, 0x00, 0x00, 0x02, 0x01})
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	if _, err := md.NextMessage(); err != io.ErrUnexpectedEOF {
		t.Fatalf("expected eof but got: %s", err)
	}
}

func expectNextElement(
	t *testing.T,
	msgs *MessageElementDecoder,
) *MessageElement {
	msg, err := msgs.NextMessageElement()
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected message not found")
	}
	return msg
}

func expectMessageElement(
	t *testing.T,
	msgs *MessageElementDecoder,
	context PCID,
	tipe PDVType,
	value string,
) {
	msg := expectNextElement(t, msgs)
	if context != msg.Context {
		t.Errorf("expected presentation context %d but got %d",
			context, msg.Context)
	}
	if tipe != msg.Type {
		t.Errorf("expected pdv type %s but got %s", tipe, msg.Type)
	}
	gotValue := toString(msg.Data)
	if value != gotValue {
		t.Errorf("expected message data %q but got %q", value, gotValue)
	}
}

func expectMessageElementError(
	t *testing.T,
	msgs *MessageElementDecoder,
	errSubstring string,
) {
	msg, err := msgs.NextMessageElement()
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected message not found")
	}

	_, err = ioutil.ReadAll(msg.Data)
	if err == nil {
		t.Fatal("expected error while reading message data")
	}
	if !strings.Contains(err.Error(), errSubstring) {
		t.Error("expected %q to contain %q", err.Error(), errSubstring)
	}
}

func expectNoMoreMessageElements(t *testing.T, msgs *MessageElementDecoder) {
	msg, err := msgs.NextMessageElement()
	if err != nil {
		t.Fatal(err)
	}
	if msg != nil {
		t.Fatalf("expected no more elements, but got %+V", msg)
	}
}

func expectPDV(
	t *testing.T,
	in *bytes.Buffer,
	expCtx PCID,
	expType PDVType,
) (bool, bytes.Buffer) {
	gotCtx, gotType, last, data, err := getpdv(in)
	if err != nil {
		t.Fatal(err)
	}
	if expCtx != gotCtx {
		t.Fatalf("expected presentation context %d, got %d", expCtx, gotCtx)
	}
	if expType != gotType {
		t.Fatalf("expected %s, got %s", expType, gotType)
	}
	return last, data
}
