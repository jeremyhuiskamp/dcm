package dcm

import (
    "encoding/binary"
)

type Explicitness bool
const (
    Explicit Explicitness = true
    Implicit Explicitness = false
)

type Deflation bool
const (
    Deflated Deflation = true
    Inflated Deflation = false
)

type PixelStorage int
const (
    Encapsulated PixelStorage = iota
    Native
)

type TransferSyntax interface {
    UID()          string
    ByteOrder()    binary.ByteOrder
    VR()           Explicitness
    Deflation()    Deflation
    PixelStorage() PixelStorage
}

type transferSyntax struct {
    uid          string
    order        binary.ByteOrder
    vr           Explicitness
    deflation    Deflation
    pixelStorage PixelStorage
}

func (ts *transferSyntax) UID() string {
    return ts.uid
}

func (ts *transferSyntax) ByteOrder() binary.ByteOrder {
    return ts.order
}

func (ts *transferSyntax) VR() Explicitness {
    return ts.vr
}

func (ts *transferSyntax) Deflation() Deflation {
    return ts.deflation
}

func (ts *transferSyntax) PixelStorage() PixelStorage {
    return ts.pixelStorage
}

var tsmap map[string]TransferSyntax = make(map[string]TransferSyntax)

func regts(ts TransferSyntax) TransferSyntax {
    tsmap[ts.UID()] = ts
    return ts
}

// Known transfer syntaxes:
var (
    ExplicitVRLittleEndian = regts(&transferSyntax{
            "1.2.840.10008.1.2.1",
            binary.LittleEndian,
            Explicit,
            Inflated,
            Native,
        })
    ImplicitVRLittleEndian = regts(&transferSyntax{
            "1.2.840.10008.1.2",
            binary.LittleEndian,
            Implicit,
            Inflated,
            Native,
        })
    ExplicitVRBigEndian = regts(&transferSyntax{
            "1.2.840.10008.1.2.2",
            binary.BigEndian,
            Explicit,
            Inflated,
            Native,
        })
    ImplicitVRBigEndian = regts(&transferSyntax{
            "1.2.840.113619.5.2",
            binary.BigEndian,
            Implicit,
            Inflated,
            Native,
        })

    // TODO: add other known syntaxes
)

// Get the transfer syntax defined by the given uid.
// If the uid is unknown, default properties are returned.
func GetTransferSyntax(uid string) TransferSyntax {
    if ts, ok := tsmap[uid]; ok {
        return ts
    }

    // I believe these are the defaults for all non-raw
    // image types...
    return &transferSyntax{
            uid,
            binary.LittleEndian,
            Explicit,
            Inflated,
            Encapsulated,
        }
}

