package dcm

import (
	"fmt"
)

//go:generate dcmtaggen

const (
	// GroupLengthElement is the tag element number indicating
	// that the tag represents the length of the group.
	GroupLengthElement uint16 = 0x0000

	// CommandGroup is the group number for a message command set.
	CommandGroup uint16 = 0x0000

	// FileMetaInfoGroup is the group number for file meta information.
	FileMetaInfoGroup uint16 = 0x0002
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

// IsCommandElement returns true if the tag is in the command group.
// See PS 3.7, Section 6.3.
func (t Tag) IsCommandElement() bool {
	return t.Group() == CommandGroup
}

// IsFileMetaInfoElement returns true if the tag is file meta information.
// See PS 3.10, Section 7.1
func (t Tag) IsFileMetaInfoElement() bool {
	return t.Group() == FileMetaInfoGroup
}

func NewTag(group, element uint16) Tag {
	return Tag((uint32(group) << 16) + uint32(element))
}

// TODO:
// IsPrivateData
// IsPrivateCreator
// GetPrivateCreator
