package dcmnet

import (
    "bytes"
    "encoding/binary"
    "io"
    "io/ioutil"
)

type PresentationContext struct {
    Id uint32
    Result uint32
    AbstractSyntax string
    TransferSyntaxes []string
}

func (pc PresentationContext) Write(dst io.Writer) {
    // item type
    binary.Write(dst, binary.LittleEndian, uint16(0x20))

    buf := new(bytes.Buffer)

    binary.Write(buf, binary.LittleEndian, pc.Id)

    binary.Write(buf, binary.LittleEndian, uint16(0x30))
    binary.Write(buf, binary.BigEndian, uint16(len(pc.AbstractSyntax)))
    buf.WriteString(pc.AbstractSyntax)

    for _, transferSyntax := range pc.TransferSyntaxes {
        binary.Write(buf, binary.LittleEndian, uint16(0x40))
        binary.Write(buf, binary.BigEndian, uint16(len(transferSyntax)))
        buf.WriteString(transferSyntax)
    }

    binary.Write(dst, binary.BigEndian, uint16(buf.Len()))
    dst.Write(buf.Bytes())
}

func readString(src io.Reader) (string) {
    bytes, _ := ioutil.ReadAll(src)
    return string(bytes)
}

func (pc *PresentationContext) Read(src io.Reader) {
    binary.Read(src, binary.LittleEndian, &pc.Id)
    //debug("Read presentation context id %d", pc.id)

    for {
        var itemType uint16
        err := binary.Read(src, binary.LittleEndian, &itemType)
        if err == io.EOF {
            break
        }
        //debug("Read item type 0x%X", itemType)

        var itemLength uint16
        binary.Read(src, binary.BigEndian, &itemLength)
        //debug("Read item length 0x%X", itemLength)

        itemSrc := io.LimitReader(src, int64(itemLength))

        switch itemType {
        case 0x30:
            pc.AbstractSyntax = readString(itemSrc)
            //debug("Read abstract syntax: %s", pc.abstractSyntax)
        case 0x40:
            ts := readString(itemSrc)
            //debug("Read transfer syntax: %s", ts)
            pc.TransferSyntaxes = append(pc.TransferSyntaxes, ts)
        default:
            //debug("Unknown item type in presentation context: 0x%X", itemType)
            readString(itemSrc)
        }
    }
}


