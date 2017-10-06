package stream

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

// tests readerStream backed by a Reader that is not also a WriterTo
func TestReaderExclusive(t *testing.T) {
	newReader := func(content []byte) io.Reader {
		return readerOnly{bytes.NewBuffer(content)}
	}

	if _, ok := newReader([]byte{}).(io.WriterTo); ok {
		t.Fatal("test is not valid")
	}

	newStream := func(content []byte) Stream {
		return NewReaderStream(newReader(content))
	}

	test(t, newStream)
}

// tests readerStream backed by a Reader that is also a WriterTo
func TestReaderAlsoWriterTo(t *testing.T) {
	newStream := func(content []byte) Stream {
		return NewReaderStream(bytes.NewBuffer(content))
	}

	test(t, newStream)
}

// tests NewBufferedStream with a WriterTo that is also a Reader
func TestBufferAlreadyStream(t *testing.T) {
	test(t, func(content []byte) Stream {
		return NewBufferedStream(bytes.NewBuffer(content))
	})
}

// tests NewBufferedStream with a WriterTo that is not also a Reader
func TestBufferWriterTo(t *testing.T) {
	newWriter := func(content []byte) io.WriterTo {
		return writerToOnly{bytes.NewBuffer(content)}
	}

	if _, ok := newWriter([]byte{}).(io.Reader); ok {
		t.Fatal("test is not valid")
	}

	test(t, func(content []byte) Stream {
		return NewBufferedStream(newWriter(content))
	})
}

// tests pipedStream with a WriterTo that is also a Reader
func TestPipeAlreadyStream(t *testing.T) {
	test(t, func(content []byte) Stream {
		return NewPipedStream(bytes.NewBuffer(content))
	})
}

// tests pipedStream with a WriterTo that is not also a Reader
func TestPipeWriterTo(t *testing.T) {
	newWriter := func(content []byte) io.WriterTo {
		return writerToOnly{bytes.NewBuffer(content)}
	}

	if _, ok := newWriter([]byte{}).(io.Reader); ok {
		t.Fatal("test is not valid")
	}

	test(t, func(content []byte) Stream {
		return NewPipedStream(newWriter(content))
	})
}

func TestNoData(t *testing.T) {
	if red, err := read(NoData); err != nil {
		t.Fatal(err)
	} else if red != "" {
		t.Fatalf("reading NoData produced %q", red)
	}
	if written, err := write(NoData); err != nil {
		t.Fatal(err)
	} else if written != "" {
		t.Fatalf("writeToing NoData produced %q", written)
	}
}

// test that the Stream returned by the func contains the given content
// through both its io.Reader and io.WriterTo implementations
func test(t *testing.T, newStream func([]byte) Stream) {
	content := "hai!"

	if red, err := read(newStream([]byte(content))); err != nil {
		t.Fatal(err)
	} else if red != content {
		t.Fatalf("unexpected read content: %q", red)
	}

	if written, err := write(newStream([]byte(content))); err != nil {
		t.Fatal(err)
	} else if written != content {
		t.Fatalf("unexpected written content: %q", written)
	}
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
