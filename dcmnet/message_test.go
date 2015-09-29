package dcmnet

import (
	"bytes"
	"github.com/Sirupsen/logrus"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func init() {
	msgLog.Logger.Level = logrus.DebugLevel
}

// TODO: tests for unexpected errors

func TestReadSingleMessageSinglePDV(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(bufpdv(1, Data, true, "data"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessage(md, 1, Data, "data")
	expectNoMoreMessages(md)
}

func TestReadSingleMessageTwoPDVs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(1, Data, true, "data2"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessage(md, 1, Data, "data1data2")
	expectNoMoreMessages(md)
}

func TestReadTwoMessages(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, true, "data"),
		bufpdv(2, Command, true, "command"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessage(md, 1, Data, "data")
	expectMessage(md, 2, Command, "command")
	expectNoMoreMessages(md)
}

func TestReadDrainMessage(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, true, "data"),
		bufpdv(2, Command, true, "command"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	msg, err := md.NextMessage()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())
	// don't read contents...

	expectMessage(md, 2, Command, "command")
	expectNoMoreMessages(md)
}

func TestReadMessageUnexpectedPDVType(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data"),
		bufpdv(1, Command, true, "command"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessageError(md, "unexpected type")
}

func TestReadMessageUnexpectedPresentationContext(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(2, Data, true, "data2"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessageError(md, "unexpected presentation context")
}

func TestReadMessageUnexpectedMissingPDV(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Command, false, "data1"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	msg, err := md.NextMessage()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())

	_, err = ioutil.ReadAll(msg.Data)
	Expect(err).To(Equal(io.ErrUnexpectedEOF))
}

func TestReadMessageWithNoPDVs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat()

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectNoMoreMessages(md)
}

func TestWriteOneMessage(t *testing.T) {
	RegisterTestingT(t)

	input := "abcdefghijklmnopqrstuvwxyz0123456789"

	// hits edge cases in pdu lengths
	// nb: a pdu that can't contain any data after the pdv header is illegal:
	for pdulen := uint32(pdvHeaderLen + 1); pdulen < 20; pdulen++ {
		for inputlen := 0; inputlen <= len(input); inputlen++ {
			curinput := input[:inputlen]

			data := new(bytes.Buffer)
			pdus := NewPDUEncoder(data)
			msgs := NewMessageEncoder(pdus, pdulen)

			msgs.NextMessage(Message{
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

func TestWriteMultipleMessages(t *testing.T) {
	RegisterTestingT(t)

	inputs := []string{
		"short",
		strings.Repeat("long", 5),
		strings.Repeat("reallylong", 30),
	}

	data := new(bytes.Buffer)
	pdus := NewPDUEncoder(data)
	msgs := NewMessageEncoder(pdus, 30)

	for context, value := range inputs {
		msgs.NextMessage(Message{
			Context: PCID(context),
			Type:    Command,
			Data:    toBufferP("command" + value),
		})
		msgs.NextMessage(Message{
			Context: PCID(context),
			Type:    Data,
			Data:    toBufferP("data" + value),
		})
	}

	expectMessage := func(context PCID, typ PDVType) string {
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
		commandValue := expectMessage(PCID(context), Command)
		Expect(commandValue).To(Equal("command" + value))

		dataValue := expectMessage(PCID(context), Data)
		Expect(dataValue).To(Equal("data" + value))
	}

	Expect(data.Len()).To(Equal(0))
}

func expectMessage(msgs MessageDecoder, context PCID, tipe PDVType, value string) {
	msg, err := msgs.NextMessage()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())
	Expect(msg.Context).To(Equal(context))
	Expect(msg.Type).To(Equal(tipe))
	Expect(toString(msg.Data)).To(Equal(value))
}

func expectMessageError(msgs MessageDecoder, errSubstring string) {
	msg, err := msgs.NextMessage()
	Expect(err).To(BeNil())
	Expect(msg).ToNot(BeNil())

	_, err = ioutil.ReadAll(msg.Data)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring(errSubstring))
}

func expectNoMoreMessages(msgs MessageDecoder) {
	msg, err := msgs.NextMessage()
	Expect(err).To(BeNil())
	Expect(msg).To(BeNil())
}
