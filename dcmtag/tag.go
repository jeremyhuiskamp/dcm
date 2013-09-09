package dcmtag

// TODO: abstract the idea of a tag dictionary.
// Need to be able to support pluggable dictionaries.

import (
    "github.com/kamper/dcm/dcm"
    "fmt"
)

type Tag struct {
    Group   uint16
    Tag     uint16
    VR      dcm.VR
    MinVM   int
    MaxVM   int
    Retired bool
    Desc    string
}

func (t Tag) LegalMultiplicity(vm int) bool {
    if vm < t.MinVM {
        return false
    }

    if t.MaxVM < 0 {
        // if there is no max, it just has to be a
        // multiple of the min
        return vm % t.MinVM == 0
    }

    return vm <= t.MaxVM
}

func (t Tag) String() string {
    return fmt.Sprintf("%s (%04X,%04X)", t.Desc, t.Group, t.Tag)
}

func (t *Tag) EqTag(tag uint32) bool {
    if t == nil {
        return false
    }
    return combine(t.Group, t.Tag) == tag
}

func (t *Tag) Eq(group, tag uint16) bool {
    return t.EqTag(combine(group, tag))
}

func combine(group, tag uint16) uint32 {
    return uint32(group) << 16 | uint32(tag)
}

var tags = make(map[uint32]Tag)

func addTag(tag Tag) Tag {
    tags[combine(tag.Group, tag.Tag)] = tag
    return tag
}

func TagVR(val uint32) dcm.VR {
    if tag, ok := tags[val]; ok {
        return tag.VR
    }
    return dcm.UN
}

func GroupTagVR(group, tag uint16) dcm.VR {
    return TagVR(combine(group, tag))
}

func GetTag(val uint32) *Tag {
    if tag, ok := tags[val]; ok {
        return &tag
    }
    return nil
}

func GetGroupTag(group, tag uint16) *Tag {
    return GetTag(combine(group, tag))
}

