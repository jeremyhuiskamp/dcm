package main

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestParseTag(t *testing.T) {
	RegisterTestingT(t)

	tag := Element{Tag: "(0123,4567)"}
	Expect(*tag.GetTagLowValue()).To(Equal(uint32(0x01234567)))
	Expect(*tag.GetTagHighValue()).To(Equal(uint32(0x01234567)))

	tag.Tag = "(fefe,fefe)"
	Expect(*tag.GetTagLowValue()).To(Equal(uint32(0xfefefefe)))
	Expect(*tag.GetTagHighValue()).To(Equal(uint32(0xfefefefe)))

	tag.Tag = "(fexx,fxfx)"
	Expect(*tag.GetTagLowValue()).To(Equal(uint32(0xfe00f0f0)))
	Expect(*tag.GetTagHighValue()).To(Equal(uint32(0xfeffffff)))
}

func TestParseKeyword(t *testing.T) {
	RegisterTestingT(t)

	tag := Element{Keyword: "asdf"}
	Expect(tag.GetKeyword()).To(Equal("asdf"))

	tag.Keyword = " a s d f "
	Expect(tag.GetKeyword()).To(Equal("asdf"))

	tag.Keyword = "as\u200Bdf"
	Expect(tag.GetKeyword()).To(Equal("asdf"))
}

func TestParseVR(t *testing.T) {
	RegisterTestingT(t)

	type input [][]string
	data := input{
		[]string{"US", "US"},
		[]string{"US or OB", "US"},
		[]string{"OB or US", "OB"},
		[]string{"", "UN"},
	}

	el := Element{Tag: "(0010,0020)"}

	for _, test := range data {
		el.VR = test[0]
		Expect(el.GetVR()).To(Equal(test[1]))
	}

	el.Tag = "(FFFE,E0DD)"
	el.VR = "See Note 2"
	Expect(el.GetVR()).To(Equal("UN"))
}
