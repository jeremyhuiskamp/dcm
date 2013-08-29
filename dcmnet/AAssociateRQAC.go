package dcmnet

import (
    "encoding/binary"
	"fmt"
    "io"
    "io/ioutil"
    "strings"
)

type AAssociateRQAC struct {
    ProtocolVersion int16
    CalledAE, CallingAE string
    PresentationContexts []PresentationContext
}

func (rq AAssociateRQAC) Write(dst io.Writer) {
    binary.Write(dst, binary.BigEndian, rq.ProtocolVersion)
    // padding:
    binary.Write(dst, binary.BigEndian, uint16(0))

    fmt.Fprintf(dst, "%-16s", rq.CalledAE)
    fmt.Fprintf(dst, "%-16s", rq.CallingAE)

    // reserved bytes:
    var reserved [32]byte
    dst.Write(reserved[:])

    for _, presentationContext := range rq.PresentationContexts {
        presentationContext.Write(dst)
    }
}

func readAE(src io.Reader) (string) {
    var bytes [16]byte
    io.ReadFull(src, bytes[:])
    return strings.TrimSpace(string(bytes[:]))
}

func (rq *AAssociateRQAC) Read(src io.Reader) {
    binary.Read(src, binary.BigEndian, &rq.ProtocolVersion)
    //debug("Read protocol version %d", rq.protocolVersion)

    var padding uint16
    binary.Read(src, binary.BigEndian, &padding)
    //debug("Read padding")


    rq.CalledAE = readAE(src)
    //debug("Read called AE %s", rq.calledAE)

    rq.CallingAE = readAE(src)
    //debug("Read calling AE %s", rq.callingAE)

    var reserved [32]byte
    io.ReadFull(src, reserved[:])

    for {
        var itemType uint16
        err := binary.Read(src, binary.LittleEndian, &itemType)
        if err == io.EOF {
            break
        }

        var itemLength uint16
        binary.Read(src, binary.BigEndian, &itemLength)

        itemSrc := io.LimitReader(src, int64(itemLength))

        switch itemType {
        case 0x21:
            pc := PresentationContext{}
            pc.Read(itemSrc)
            rq.PresentationContexts = append(rq.PresentationContexts, pc)
        default:
            //debug("Skipping unknown item type 0x%X of length %d", itemType, itemLength)
            io.Copy(ioutil.Discard, itemSrc)
        }
    }
}

