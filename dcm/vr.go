package dcm

// this might be better represented as an interface...
type VR struct {
    Name string

    // Used to pad odd-length values.
    Padding byte

    // TODO: this is silly, replace it with a bool flag.
    // All it means is whether the vallen can fit into
    // 2 bytes or if it requires 4
    HeaderLength uint32
}

var vrmap map[string]VR = make(map[string]VR)

func vr(name string, padding byte, headerLength uint32) VR {
    vr := VR{name, padding, headerLength}
    vrmap[vr.Name] = vr
    return vr
}

// These are mutable, which isn't great.
// Seems the cleaner way to do this is to implement an interface with
// a private type that would have only getters.
var (
    AE VR = vr("AE", ' ',  8)
    AS VR = vr("AS", ' ',  8)
    AT VR = vr("AT",   0,  8)
    CS VR = vr("CS", ' ',  8)
    DA VR = vr("DA", ' ',  8)
    DS VR = vr("DS", ' ',  8)
    DT VR = vr("DT", ' ',  8)
    FD VR = vr("FD",   0,  8)
    FL VR = vr("FL",   0,  8)
    IS VR = vr("IS", ' ',  8)
    LO VR = vr("LO", ' ',  8)
    LT VR = vr("LT", ' ',  8)
    OB VR = vr("OB",   0, 12)
    OF VR = vr("OF",   0, 12)
    OW VR = vr("OW",   0, 12)
    PN VR = vr("PN", ' ',  8)
    SH VR = vr("SH", ' ',  8)
    SL VR = vr("SL", ' ',  8)
    SQ VR = vr("SQ",   0, 12)
    SS VR = vr("SS",   0,  8)
    ST VR = vr("ST", ' ',  8)
    TM VR = vr("TM", ' ',  8)
    UI VR = vr("UI",   0,  8)
    UL VR = vr("UL",   0,  8)
    UN VR = vr("UN",   0, 12)
    US VR = vr("US",   0,  8)
    UT VR = vr("UT", ' ', 12)
)

func GetVR(code uint16) VR {
    name := string([]byte{byte(code >> 8), byte(code)})
    if vr, ok := vrmap[name]; ok {
        return vr
    }

    return UN
}

func TagHasVR(group, tag uint16) bool {
    return !(group == 0xFFFE && (tag == 0xE000 || tag == 0xE00D || tag == 0xE0DD))
}

// TODO: write ToString(io.Reader, VR, maxlen, charset)
// if charset == null, charset == ascii
// read bytes and transform into a string until max number of characters
// EOF is not an error except when the length is not a proper multiple
// of the size of each item (eg, floats need to be multiples of 4 bytes)
// To do this, we probably need each vr to be its own type.

