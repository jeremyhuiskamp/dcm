package dcmnet

import (
	"github.com/kamper/dcm/stream"
	"io"
	"io/ioutil"
)

// StreamDecoder is a helper for parsing streams that are segmented into chunks
// using a header which includes a length field
type StreamDecoder struct {
	stream    io.Reader
	lastChunk io.Reader
}

// ParseHeaderFunc parses a header and returns the length of the following data.
// Should it report returning an error if the header is invalid?
type ParseHeaderFunc func(header []byte) int64

func (s *StreamDecoder) NextChunk(headerLen int, parse ParseHeaderFunc) (stream.Stream, error) {
	if s.lastChunk != nil {
		// is it feasible to try seeking here?  that may be more efficient.
		// however, we're likely wrapped in several layers of LimitedReaders,
		// which don't support seeking, and unwrapping them is probably not
		// worth the trouble
		_, err := io.Copy(ioutil.Discard, s.lastChunk)
		if err != nil {
			return nil, err
		}
	}

	header := make([]byte, headerLen)
	len, err := io.ReadFull(s.stream, header)
	if len == 0 {
		// no more chunks
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	dataLen := parse(header)

	// TODO: sanity check the length?
	s.lastChunk = io.LimitReader(s.stream, dataLen)

	return stream.NewReaderStream(s.lastChunk), nil
}
