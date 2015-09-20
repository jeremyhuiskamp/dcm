package dcmnet

// These are not tests, but utils for writing tests.  Not sure what the naming
// convention for such files should be in go.

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
)

// TODO: utils for inserting errors at arbitrary points in the data
// perhaps we want io.Reader to be our basic unit of data, with io.MultiReader
// to concatenate

// bufcat concatenates a bunch of buffers
func bufcat(bufs ...interface{}) (buf bytes.Buffer) {
	for _, bufx := range bufs {
		b := toBuffer(bufx)
		buf.ReadFrom(&b)
	}
	return buf
}

// bufpdu creates a pdu in a buffer
func bufpdu(typ PDUType, pdvs ...interface{}) (buf bytes.Buffer) {
	payload := bufcat(pdvs...)
	header := make([]byte, 6)
	header[0] = byte(typ)
	binary.BigEndian.PutUint32(header[2:6], uint32(payload.Len()))
	buf.Write(header)
	buf.ReadFrom(&payload)

	return buf
}

// bufpdv creates a pdv in a buffer
func bufpdv(context uint8, tipe PDVType, last bool, data interface{}) (buf bytes.Buffer) {
	dataBuf := toBuffer(data)
	header := make([]byte, 6)
	binary.BigEndian.PutUint32(header[0:4], uint32(dataBuf.Len()+2))
	header[4] = context

	var pdv PDV
	pdv.SetType(tipe)
	pdv.SetLast(last)
	header[5] = pdv.Flags

	buf.Write(header)
	buf.ReadFrom(&dataBuf)

	return buf
}

// toBuffer co-erces a thing into a Buffer
func toBuffer(thing interface{}) bytes.Buffer {
	if buf, ok := thing.(bytes.Buffer); ok {
		return buf
	}

	if str, ok := thing.(string); ok {
		thing = []byte(str)
	}

	if reader, ok := thing.(io.Reader); ok {
		t, err := ioutil.ReadAll(reader)
		if err != nil {
			panic("Unable to read thing")
		}

		thing = t
	}

	if b, ok := thing.([]byte); ok {
		return *bytes.NewBuffer(b)
	}

	panic("Could not convert thing to bytes.Buffer")
}

func toBytes(thing interface{}) []byte {
	b := toBuffer(thing)
	return b.Bytes()
}

func toString(thing interface{}) string {
	b := toBuffer(thing)
	return b.String()
}
