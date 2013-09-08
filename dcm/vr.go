package dcm

// this might be better represented as an interface...
type VR struct {
    Code uint16
    Padding uint32
    HeaderLength uint32
}

var vrmap map[uint16]VR = make(map[uint16]VR)

func regvr(vr VR) VR {
    vrmap[vr.Code] = vr
    return vr
}

// this is mutable, which isn't great
// seems the cleaner way to do this is to implement an interface with
// a private type that would have only getters
var AE VR = regvr(VR{0x4145, ' ',  8})
var OB VR = regvr(VR{0x4F42,   0, 12})
var SH VR = regvr(VR{0x5348, ' ',  8})
var UI VR = regvr(VR{0x5549,   0,  8})
var UL VR = regvr(VR{0x554C,   0,  8})
var UN VR = regvr(VR{0x554E,   0, 12})

func GetVR(code uint16) VR {
    if vr, ok := vrmap[code]; ok {
        return vr
    }

    return UN
}

func TagHasVR(group, tag uint16) bool {
    return !(group == 0xFFFE && (tag == 0xE000 || tag == 0xE00D || tag == 0xE0DD))
}

func VRName(vr VR) string {
    // seems like a rather hackish way to convert a uint16 to a string...
    return string([]byte{byte(vr.Code >> 8), byte(vr.Code)})
}

// TODO: write ToString(io.Reader, VR, maxlen, charset)
// if charset == null, charset == ascii
// read bytes and transform into a string until max number of characters
// EOF is not an error except when the length is not a proper multiple
// of the size of each item (eg, floats need to be multiples of 4 bytes)
// To do this, we probably need each vr to be its own type.

