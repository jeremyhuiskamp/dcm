package dcm

// In-memory dicom object structures.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// gah, should try to outsource this sorting somewhere else
type tagSlice []Tag

func (p tagSlice) Len() int           { return len(p) }
func (p tagSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p tagSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Element interface {
	GetTag() Tag
	String() string
}

type Object struct {
	elements map[Tag]Element
	// TODO: endianness, specific character set and default time zone
	// These are needed for interpreting the rest of the data.
	// For nested items, need to either replicate those, or reference parent.
	// Replication is sucky because it implies that those values could be
	// changed on the children, which is often not really the case.  Also, if
	// someone sets a value on the top-level (eg, while creating an object),
	// we'd have to consider replicating that value down to any children, if
	// they exist yet...
}

func (o Object) GetString(tag Tag) string {
	if el, ok := o.elements[tag]; ok {
		if se, ok := el.(SimpleElement); ok {
			return strings.TrimRight(string(se.Data), "\x00")
		}
	}

	return ""
}

// TODO: return an ok bool like a map lookup?
func (o Object) GetUint16(tag Tag) uint16 {
	if el, ok := o.elements[tag]; ok {
		if se, ok := el.(SimpleElement); ok {
			// TODO: honour actual endianness
			return binary.BigEndian.Uint16(se.Data[:2])
		}
	}

	return 0
}

func (o Object) String() string {
	var buf bytes.Buffer

	o.ForEach(func(tag Tag, e Element) bool {
		buf.WriteString(o.elements[tag].String())
		buf.WriteString("\n")
		return true
	})

	return buf.String()
}

// Iterate over the elements in the object.
func (o Object) ForEach(f func(Tag, Element) bool) {
	keys := make(tagSlice, 0, len(o.elements))
	for key := range o.elements {
		keys = append(keys, key)
	}
	sort.Sort(keys)

	for _, tag := range keys {
		if !f(tag, o.elements[tag]) {
			break
		}
	}
}

func (o Object) Put(e Element) {
	o.elements[e.GetTag()] = e
}

func (o Object) Get(tag Tag) *Element {
	if e, ok := o.elements[tag]; ok {
		return &e
	}

	return nil
}

func NewObject() Object {
	return Object{make(map[Tag]Element)}
}

type SimpleElement struct {
	Tag Tag

	// Tag also has a VR, but that's the standard value, and we could have a
	// different one in a real object
	VR VR

	Data []byte
}

func (se SimpleElement) GetTag() Tag {
	return se.Tag
}

func (se SimpleElement) String() string {
	return fmt.Sprintf("%s %s", se.Tag, se.VR)
}

type SequenceElement struct {
	Tag     Tag
	Objects []Object
}

func (se SequenceElement) GetTag() Tag {
	return se.Tag
}

func (se SequenceElement) String() string {
	return fmt.Sprintf("%s SQ", se.Tag)
}

type EncapsulatedElement struct {
	Tag  Tag
	VR   VR
	Data [][]byte
}
