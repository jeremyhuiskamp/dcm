package dcm

import "testing"

func TestTag(t *testing.T) {
	for _, test := range []struct {
		tag     Tag
		group   uint16
		element uint16
	}{
		{Tag(0xffff0000), 0xffff, 0x0000},
		{Tag(0), 0, 0},
		{Tag(0xffff), 0, 0xffff},
	} {
		if group := test.tag.Group(); group != test.group {
			t.Errorf("tag %s has unexpected group 0x%04x", test.tag, group)
		}
		if element := test.tag.Element(); element != test.element {
			t.Errorf("tag %s has unexpected element 0x%04x", test.tag, element)
		}
	}
}

func TestNewTag(t *testing.T) {
	for _, test := range []struct {
		group   uint16
		element uint16
		tag     Tag
	}{
		{0, 0, Tag(0)},
		{0xffff, 0xffff, Tag(0xffffffff)},
		{0x0123, 0x4567, Tag(0x01234567)},
	} {
		if tag := NewTag(test.group, test.element); tag != test.tag {
			t.Errorf("group 0x%04x and element 0x%04x give unexpected tag %s",
				test.group, test.element, tag)
		}
	}
}

func TestGroupLength(t *testing.T) {
	for tag, exp := range map[Tag]bool{
		Tag(0xffff0000): true,
		Tag(0xffff0001): false,
	} {
		if got := tag.IsGroupLength(); got != exp {
			t.Errorf("unexpected group length %t for tag %s", got, tag)
		}
	}
}

func TestCommandElement(t *testing.T) {
	for tag, exp := range map[Tag]bool{
		Tag(0x00000008): true,
		Tag(0x00020008): false,
	} {
		if got := tag.IsCommandElement(); got != exp {
			t.Errorf("unexpected command element %t for tag %s", got, tag)
		}
	}
}

func TestFileMetaInfoElement(t *testing.T) {
	for tag, exp := range map[Tag]bool{
		Tag(0x00000008): false,
		Tag(0x00020008): true,
		Tag(0x00040008): false,
	} {
		if got := tag.IsFileMetaInfoElement(); got != exp {
			t.Errorf("unexpected file meta info %t for tag %s", got, tag)
		}
	}
}
