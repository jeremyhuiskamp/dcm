package dcm

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestTag(t *testing.T) {
	RegisterTestingT(t)

	tag := Tag(0xffff0000)
	Expect(tag.Group()).To(Equal(uint16(0xffff)))
	Expect(tag.Element()).To(Equal(uint16(0x0000)))
	Expect(uint32(tag)).To(Equal(uint32(0xffff0000)))

	tag = Tag(0)
	Expect(tag.Group()).To(Equal(uint16(0)))
	Expect(tag.Element()).To(Equal(uint16(0)))
	Expect(uint32(tag)).To(Equal(uint32(0)))

	tag = Tag(0xffff)
	Expect(tag.Group()).To(Equal(uint16(0)))
	Expect(tag.Element()).To(Equal(uint16(0xffff)))
	Expect(uint32(tag)).To(Equal(uint32(0xffff)))
}

func TestGroupLength(t *testing.T) {
	RegisterTestingT(t)

	Expect(Tag(0xffff0000).IsGroupLength()).To(BeTrue())
	Expect(Tag(0xffff0001).IsGroupLength()).To(BeFalse())
}

func TestCommandElement(t *testing.T) {
	RegisterTestingT(t)

	Expect(Tag(0x00000008).IsCommandElement()).To(BeTrue())
	Expect(Tag(0x00020008).IsCommandElement()).To(BeFalse())
}

func TestFileMetaInfoElement(t *testing.T) {
	RegisterTestingT(t)

	Expect(Tag(0x00000008).IsFileMetaInfoElement()).To(BeFalse())
	Expect(Tag(0x00020008).IsFileMetaInfoElement()).To(BeTrue())
	Expect(Tag(0x00040008).IsFileMetaInfoElement()).To(BeFalse())
}

func TestNewTag(t *testing.T) {
	RegisterTestingT(t)

	Expect(NewTag(0x0000, 0x0000)).To(Equal(Tag(0x0)))
	Expect(NewTag(0xffff, 0xffff)).To(Equal(Tag(0xffffffff)))
	Expect(NewTag(0x0123, 0x4567)).To(Equal(Tag(0x01234567)))
}
