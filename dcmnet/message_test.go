package dcmnet

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/jeremyhuiskamp/dcm/dcm"
	. "github.com/onsi/gomega"
)

func init() {
	msgLog.Logger.Level = logrus.DebugLevel
}

// TODO: tests for unexpected errors

func TestReadSingleMessageElementSinglePDV(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(bufpdv(1, Data, true, "data"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElement(med, 1, Data, "data")
	expectNoMoreMessageElements(med)
}

func TestReadSingleMessageElementTwoPDVs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(1, Data, true, "data2"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElement(med, 1, Data, "data1data2")
	expectNoMoreMessageElements(med)
}

func TestReadTwoMessageElements(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, true, "data"),
		bufpdv(2, Command, true, "command"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElement(med, 1, Data, "data")
	expectMessageElement(med, 2, Command, "command")
	expectNoMoreMessageElements(med)
}

func TestReadDrainMessageElement(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, true, "data"),
		bufpdv(2, Command, true, "command"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	msg, err := med.NextMessageElement()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())
	// don't read contents...

	expectMessageElement(med, 2, Command, "command")
	expectNoMoreMessageElements(med)
}

func TestReadMessageElementUnexpectedPDVType(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data"),
		bufpdv(1, Command, true, "command"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElementError(med, "unexpected type")
}

func TestReadMessageElementUnexpectedPresentationContext(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(2, Data, true, "data2"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectMessageElementError(med, "unexpected presentation context")
}

func TestReadMessageElementUnexpectedMissingPDV(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Command, false, "data1"))

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	msg, err := med.NextMessageElement()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())

	_, err = ioutil.ReadAll(msg.Data)
	Expect(err).To(Equal(io.ErrUnexpectedEOF))
}

func TestReadMessageElementWithNoPDVs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat()

	med := NewMessageElementDecoder(NewPDVDecoder(&data))

	expectNoMoreMessageElements(med)
}

func TestWriteOneMessageElement(t *testing.T) {
	RegisterTestingT(t)

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
				pduType, pduData, err := getpdu(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(pduType).To(Equal(PDUPresentationData))
				context, pdvType, thislast, pdvData, err := getpdv(&pduData)
				last = thislast
				Expect(err).ToNot(HaveOccurred())
				Expect(context).To(Equal(PCID(1)))
				Expect(pdvType).To(Equal(Command))
				output.Write(pdvData.Bytes())
			}

			Expect(data.Len()).To(Equal(0))

			if len(curinput) == 0 {
				Expect(last).To(BeFalse())
			} else {
				Expect(last).To(BeTrue())
			}

			Expect(output.String()).To(Equal(curinput),
				"pdulen=%d, inputlen=%d", pdulen, inputlen)

			if t.Failed() {
				return
			}
		}
	}
}

func TestWriteMultipleMessageElements(t *testing.T) {
	RegisterTestingT(t)

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
			pduType, pduData, err := getpdu(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(pduType).To(Equal(PDUPresentationData))

			actualContext, pdvType, pdvLast, pdvData, err := getpdv(&pduData)
			last = pdvLast
			Expect(err).ToNot(HaveOccurred())
			Expect(actualContext).To(Equal(context))
			Expect(pdvType).To(Equal(typ))
			output.Write(pdvData.Bytes())
		}

		return output.String()
	}

	for context, value := range inputs {
		commandValue := expectMessageElement(PCID(context), Command)
		Expect(commandValue).To(Equal("command" + value))

		dataValue := expectMessageElement(PCID(context), Data)
		Expect(dataValue).To(Equal("data" + value))
	}

	Expect(data.Len()).To(Equal(0))
}

func TestReadNoMessages(t *testing.T) {
	RegisterTestingT(t)

	var data bytes.Buffer
	var pcs PresentationContexts
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	msg, err := md.NextMessage()
	Expect(msg).To(BeNil())
	Expect(err).To(BeNil())
}

func TestReadUnexpectedData(t *testing.T) {
	RegisterTestingT(t)

	data := bufpdv(1, Data, true, "xxxx")
	var pcs PresentationContexts
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	_, err := md.NextMessage()
	Expect(err).To(MatchError(ContainSubstring("Expected a command message")))
}

func TestReadUnexpectedPCID(t *testing.T) {
	RegisterTestingT(t)

	data := bufpdv(1, Command, true, "xxx")
	var pcs PresentationContexts
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	_, err := md.NextMessage()
	Expect(err).To(MatchError(ContainSubstring("Unrecognized presentation context")))
}

func TestExpectDatasetButFindNone(t *testing.T) {
	RegisterTestingT(t)

	pcs := PresentationContexts{
		Requested: []PresentationContext{
			{
				ID:             1,
				AbstractSyntax: "??",
			},
		},
		Accepted: []PresentationContext{
			{
				ID:     1,
				Result: PCAcceptance,
				TransferSyntaxes: []dcm.TransferSyntax{
					dcm.ImplicitVRLittleEndian,
				},
			},
		},
	}

	data := bufpdv(1, Command, true,
		// data set type element indicating that data should be expected:
		[]byte{0x00, 0x00, 0x00, 0x08, 0x02, 0x00, 0x00, 0x00, 0x02, 0x01})
	md := NewMessageDecoder(pcs, NewMessageElementDecoder(NewPDVDecoder(&data)))

	_, err := md.NextMessage()
	Expect(err).To(Equal(io.ErrUnexpectedEOF))
}

func expectMessageElement(msgs MessageElementDecoder, context PCID,
	tipe PDVType, value string) {
	msg, err := msgs.NextMessageElement()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())
	Expect(msg.Context).To(Equal(context))
	Expect(msg.Type).To(Equal(tipe))
	Expect(toString(msg.Data)).To(Equal(value))
}

func expectMessageElementError(msgs MessageElementDecoder, errSubstring string) {
	msg, err := msgs.NextMessageElement()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())

	_, err = ioutil.ReadAll(msg.Data)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring(errSubstring))
}

func expectNoMoreMessageElements(msgs MessageElementDecoder) {
	msg, err := msgs.NextMessageElement()
	Expect(err).To(BeNil())
	Expect(msg).To(BeNil())
}
