package dcm

// TODO: consider representing this as an interface instead
type VR struct {
    // The two-character name of the VR
    Name    string

    // Used to pad odd-length values.
    Padding byte

    // Whether the VR requires 4 bytes to encode its
    // length
    Long    bool
}

var vrmap map[string]VR = make(map[string]VR)

func vr(name string, padding byte, long bool) VR {
    vr := VR{name, padding, long}
    vrmap[name] = vr
    return vr
}

const (
    // padding values:
    text byte = ' '
    bin       = 0

    // Long values:
    long bool = true
    short     = false
)

var (
    // These need to be updated with methods to actually
    // parse values
    AE = vr("AE", text, short)
    AS = vr("AS", text, short)
    AT = vr("AT",  bin, short)
    CS = vr("CS", text, short)
    DA = vr("DA", text, short)
    DS = vr("DS", text, short)
    DT = vr("DT", text, short)
    FD = vr("FD",  bin, short)
    FL = vr("FL",  bin, short)
    IS = vr("IS", text, short)
    LO = vr("LO", text, short)
    LT = vr("LT", text, short)
    OB = vr("OB",  bin,  long)
    OF = vr("OF",  bin,  long)
    OW = vr("OW",  bin,  long)
    PN = vr("PN", text, short)
    SH = vr("SH", text, short)
    SL = vr("SL", text, short)
    SQ = vr("SQ",  bin,  long)
    SS = vr("SS",  bin, short)
    ST = vr("ST", text, short)
    TM = vr("TM", text, short)
    UI = vr("UI",  bin, short)
    UL = vr("UL",  bin, short)
    UN = vr("UN",  bin,  long)
    US = vr("US",  bin, short)
    UT = vr("UT", text,  long)
    //?? = vr("??",  bin,  long)
)

func GetVR(name string) VR {
    if vr, ok := vrmap[name]; ok {
        return vr
    }

    return UN
}

func TagHasVR(group, tag uint16) bool {
    return !(group == 0xFFFE && (tag == 0xE000 || tag == 0xE00D || tag == 0xE0DD))
}

