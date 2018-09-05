package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/jeremyhuiskamp/dcm/dcm"
	"github.com/jeremyhuiskamp/dcm/dcmnet"
)

const (
	logDebug = true
)

// Fprintf to stderr, putting our program name on the front.
func warnf(format string, elems ...interface{}) {
	fmt.Fprintf(os.Stderr, "cecho: "+format, elems...)
}

// Fprintln to stderr, putting our program name on the front.
func warnln(elems ...interface{}) {
	e := append([]interface{}{"cecho:"}, elems...)
	fmt.Fprintln(os.Stderr, e...)
}

func debug(format string, elems ...interface{}) {
	if logDebug {
		fmt.Fprintf(os.Stderr, format+"\n", elems...)
	}
}

func writePDU(dst io.Writer, pduType uint16, src bytes.Buffer) {
	_ = binary.Write(dst, binary.LittleEndian, pduType)
	_ = binary.Write(dst, binary.BigEndian, uint32(src.Len()))
	_, _ = dst.Write(src.Bytes())
}

func readPDU(src io.Reader) (pduType uint16, pdu io.Reader) {
	binary.Read(src, binary.LittleEndian, &pduType)
	debug("Read pdu type %X", pduType)
	var pduLength uint32
	binary.Read(src, binary.BigEndian, &pduLength)
	debug("Read pdu length %X", pduLength)
	return pduType, io.LimitReader(src, int64(pduLength))
}

func main() {
	var callingAE, calledAE, addr string

	flag.StringVar(&callingAE, "calling", "CECHO", "Calling AE Title")
	flag.StringVar(&calledAE, "called", "CECHO", "Called AE Title")
	flag.StringVar(&addr, "d", "", "host:port of SCP")

	flag.Parse()

	fmt.Printf("Calling AE: %s\n", callingAE)
	fmt.Printf("Called AE: %s\n", calledAE)
	fmt.Printf("Address: %s\n", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		warnf("%s\n", err)
		return
	}

	rq := dcmnet.AssociateRQAC{
		ProtocolVersion: 1,
		CalledAE:        calledAE,
		CallingAE:       callingAE,
		PresentationContexts: []dcmnet.PresentationContext{
			{
				ID:               1,
				AbstractSyntax:   "1.2.840.10008.1.1",
				TransferSyntaxes: []dcm.TransferSyntax{dcm.ImplicitVRLittleEndian},
			},
		},
	}
	fmt.Printf("Sending %v\n", rq)

	var buf bytes.Buffer
	rq.Write(&buf)
	writePDU(conn, 1, buf)

	_, pduSrc := readPDU(conn)

	ac := dcmnet.AssociateRQAC{}
	ac.Read(pduSrc)
	fmt.Printf("Read %v\n", ac)

	var release bytes.Buffer
	binary.Write(&release, binary.LittleEndian, uint16(0))
	binary.Write(&release, binary.LittleEndian, uint16(0))
	writePDU(conn, uint16(dcmnet.PDUReleaseRQ), release)

	conn.Close()
}
