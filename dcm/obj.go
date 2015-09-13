package dcm

// In-memory dicom object structures.

import (
	"bytes"
	"fmt"
	"sort"
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
	// We probably don't want to expose this in the api, might want to switch
	// to an underlying structure that is more efficient for in-order access.
	Elements map[Tag]Element
}

func (o Object) String() string {
	var buf bytes.Buffer

	o.ForEach(func(tag Tag, e Element) bool {
		buf.WriteString(o.Elements[tag].String())
		buf.WriteString("\n")
		return true
	})

	return buf.String()
}

// Iterate over the elements in the object.
func (o Object) ForEach(f func(Tag, Element) bool) {
	keys := make(tagSlice, 0, len(o.Elements))
	for key := range o.Elements {
		keys = append(keys, key)
	}
	sort.Sort(keys)

	for _, tag := range keys {
		if !f(tag, o.Elements[tag]) {
			break
		}
	}
}

func (o Object) Put(e Element) {
	o.Elements[e.GetTag()] = e
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
	return fmt.Sprintf("%s %s", se.Tag, se.VR /*, se.Tag.desc*/)
}

type SequenceElement struct {
	Tag     Tag
	Objects []Object
}

func (se SequenceElement) GetTag() Tag {
	return se.Tag
}

func (se SequenceElement) String() string {
	return fmt.Sprintf("%s SQ", se.Tag /*, se.Tag.desc*/)
}

type EncapsulatedElement struct {
	Tag  Tag
	VR   VR
	Data [][]byte
}
