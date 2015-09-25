package dcmnet

import (
	"github.com/Sirupsen/logrus"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"testing"
)

func init() {
	msgLog.Logger.Level = logrus.DebugLevel
}

// TODO: tests for unexpected errors

func TestSingleMessageSinglePDV(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(bufpdv(1, Data, true, "data"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessage(md, 1, Data, "data")
	expectNoMoreMessages(md)
}

func TestSingleMessageTwoPDVs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(1, Data, true, "data2"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessage(md, 1, Data, "data1data2")
	expectNoMoreMessages(md)
}

func TestTwoMessages(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, true, "data"),
		bufpdv(2, Command, true, "command"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessage(md, 1, Data, "data")
	expectMessage(md, 2, Command, "command")
	expectNoMoreMessages(md)
}

func TestDrainMessage(t *testing.T) {
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

func TestMessageUnexpectedPDVType(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data"),
		bufpdv(1, Command, true, "command"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessageError(md, "unexpected type")
}

func TestMessageUnexpectedPresentationContext(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat(
		bufpdv(1, Data, false, "data1"),
		bufpdv(2, Data, true, "data2"))

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectMessageError(md, "unexpected presentation context")
}

func TestMessageUnexpectedMissingPDV(t *testing.T) {
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

func TestMessageWithNoPDVs(t *testing.T) {
	RegisterTestingT(t)

	data := bufcat()

	md := NewMessageDecoder(NewPDVDecoder(&data))

	expectNoMoreMessages(md)
}

func expectMessage(msgs MessageDecoder, context uint8, tipe PDVType, value string) {
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
