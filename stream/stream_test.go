package stream

import (
	"bytes"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"testing"
)

// tests readerStream backed by a Reader that is not also a WriterTo
func TestReaderExclusive(t *testing.T) {
	RegisterTestingT(t)

	newReader := func(content []byte) io.Reader {
		return readerOnly{bytes.NewBuffer(content)}
	}

	// just to assert that this test is actually valid:
	Expect(newReader([]byte("x"))).ToNot(BeAssignableToTypeOf((*io.WriterTo)(nil)))

	newStream := func(content []byte) Stream {
		return NewReaderStream(newReader(content))
	}

	test(newStream)
}

// tests readerStream backed by a Reader that is also a WriterTo
func TestReaderAlsoWriterTo(t *testing.T) {
	RegisterTestingT(t)

	newStream := func(content []byte) Stream {
		return NewReaderStream(bytes.NewBuffer(content))
	}

	test(newStream)
}

// tests NewBufferedStream with a WriterTo that is also a Reader
func TestBufferAlreadyStream(t *testing.T) {
	RegisterTestingT(t)

	test(func(content []byte) Stream {
		return NewBufferedStream(bytes.NewBuffer(content))
	})
}

// tests NewBufferedStream with a WriterTo that is not also a Reader
func TestBufferWriterTo(t *testing.T) {
	RegisterTestingT(t)

	newWriter := func(content []byte) io.WriterTo {
		return writerToOnly{bytes.NewBuffer(content)}
	}

	// just to assert that this test is actually valid:
	Expect(newWriter([]byte("x"))).ToNot(BeAssignableToTypeOf((*io.Reader)(nil)))

	test(func(content []byte) Stream {
		return NewBufferedStream(newWriter(content))
	})
}

// tests pipedStream with a WriterTo that is also a Reader
func TestPipeAlreadyStream(t *testing.T) {
	RegisterTestingT(t)

	test(func(content []byte) Stream {
		return NewPipedStream(bytes.NewBuffer(content))
	})
}

// tests pipedStream with a WriterTo that is not also a Reader
func TestPipeWriterTo(t *testing.T) {
	RegisterTestingT(t)

	newWriter := func(content []byte) io.WriterTo {
		return writerToOnly{bytes.NewBuffer(content)}
	}

	// just to assert that this test is actually valid:
	Expect(newWriter([]byte("x"))).ToNot(BeAssignableToTypeOf((*io.Reader)(nil)))

	test(func(content []byte) Stream {
		return NewPipedStream(newWriter(content))
	})
}

// test that the Stream returned by the func contains the given content
// through both its io.Reader and io.WriterTo implementations
func test(newStream func([]byte) Stream) {
	content := "hai!"
	Expect(read(newStream([]byte(content)))).To(Equal(content))
	Expect(write(newStream([]byte(content)))).To(Equal(content))
}

// read the content of the Stream via its io.Reader implementation
func read(stream Stream) (string, error) {
	bytes, err := ioutil.ReadAll(stream)
	return string(bytes), err
}

// read the content of the Stream via its io.WriterTo implementation
func write(stream Stream) (string, error) {
	var buf bytes.Buffer
	_, err := stream.WriteTo(&buf)
	return buf.String(), err
}

// readerOnly wraps another io.Reader to hide any other interfaces it might
// satisfy
type readerOnly struct {
	reader io.Reader
}

func (ro readerOnly) Read(buf []byte) (int, error) {
	return ro.reader.Read(buf)
}

// writerToOnly wraps another io.WriterTo to hide any other interfaces it might
// satisfy
type writerToOnly struct {
	writerTo io.WriterTo
}

func (wto writerToOnly) WriteTo(w io.Writer) (int64, error) {
	return wto.writerTo.WriteTo(w)
}
