package dcm

import (
	"fmt"
)

//go:generate dcmtaggen

const (
	GroupLengthElement uint16 = 0
)

type Tag uint32

// Get the group number of the tag
func (t Tag) Group() uint16 {
	return uint16(t >> 16)
}

// Get the element number of the tag
func (t Tag) Element() uint16 {
	return uint16(t)
}

// Check if the tag is a group length tag.
// That is, if the element number is 0.
func (t Tag) IsGroupLength() bool {
	return t.Element() == GroupLengthElement
}

func (t Tag) String() string {
	return fmt.Sprintf("(%04X,%04X)", t.Group(), t.Element())
}

// Check if the tag has a value representation.
// Only Item (FFFE,E000), ItemDelimitationItem (FFFE,E00D) and
// SequenceDelimitationItem (FFFE,E0DD) do not.
// See PS 3.5, seciton 7.5.
func (t Tag) HasVR() bool {
	return t != Item && t != ItemDelimitationItem && t != SequenceDelimitationItem
}

func NewTag(group, element uint16) Tag {
	return Tag((uint32(group) << 16) + uint32(element))
}

// TODO:
// IsCommandElement
// IsFileMetaInfoElement (TODO: check standard name for this)
// IsPrivateData
// IsPrivateCreator
// GetPrivateCreator
