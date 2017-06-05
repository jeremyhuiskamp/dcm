// Package stream defines an abstraction over a stream of data where the source
// of the data may prefer to push the data or allow it to be pulled and the
// consumer of the data may prefer to pull the data or allow it to be pushed.
package stream

import (
	"bytes"
	"io"
)

// Stream represents a stream of data that can be either pushed or pulled.
// Typically the data may only be used once, and hence only one of the methods
// may be used.
// It is typically easier for a producer of data to implement io.WriterTo but
// easier for a consumer to use an io.Reader.  Since the consumer has the choice,
// it should try to use io.WriterTo if it is able.
//
// TODO: add io.Closer interface for safety
type Stream interface {
	io.Reader
	io.WriterTo
}

// NewReaderStream returns a Stream backed by an io.Reader.
// It implements WriteTo by copying data from the Reader to the Writer.
func NewReaderStream(reader io.Reader) Stream {
	return readerStream{reader}
}

type readerStream struct {
	io.Reader
}

func (rs readerStream) WriteTo(w io.Writer) (int64, error) {
	// NB: don't pass rs, that results in infinite recursion!
	return io.Copy(w, rs.Reader)
}

// NewBufferedStream returns a Stream backed by an io.WriterTo.
// It implements io.Reader by buffering all the data in memory.  It is simple
// but not suitable for large amounts of data.
func NewBufferedStream(wt io.WriterTo) Stream {
	// if the WriterTo already implements io.Reader, just use it directly
	// instead of buffering the data ourselves (possibly double-buffering):
	if stream, ok := wt.(Stream); ok {
		return stream
	}

	var buf bytes.Buffer
	wt.WriteTo(&buf)
	return &buf
}

// NewPipedStream returns a Stream backed by an io.WriterTo.
// It implements io.Reader by using io.Pipe() and having the backing WriterTo
// feed the pipe in a background go-routine.  It is scalable, but probably
// error prone.
// TODO: additional error handling and testing
func NewPipedStream(wt io.WriterTo) Stream {
	// if the WriterTo already implements io.Reader, just use it directly
	// instead of buffering the data ourselves (possibly double-buffering):
	if stream, ok := wt.(Stream); ok {
		return stream
	}

	return &pipedStream{wt, nil}
}

type pipedStream struct {
	io.WriterTo
	reader io.Reader
}

func (ps *pipedStream) Read(buf []byte) (int, error) {
	// first time setup (don't do this in constructor, since it precludes use
	// as WriterTo):
	if ps.reader == nil {
		reader, writer := io.Pipe()
		ps.reader = reader
		// TODO: provide a mechanism to make sure this goroutine will stop
		// Currently won't if caller doesn't consume all the data.
		go func() {
			_, err := ps.WriteTo(writer)
			writer.CloseWithError(err)
		}()
	}

	return ps.reader.Read(buf)
}

// TODO: can intercept calls to WriteTo and panic if ps.reader != nil ?
// alternatively, if ps.reader != nil, could just call io.Copy()...

type nodata struct{}

func (nd nodata) Read(buf []byte) (int, error) {
	return 0, io.EOF
}

func (nd nodata) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

// NoData is a dummy Stream that contains nothing.
var NoData Stream = nodata{}
