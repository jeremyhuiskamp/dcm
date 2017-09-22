package main

import "testing"

func TestParseTag(t *testing.T) {
	for input, exp := range map[string]struct {
		low, high uint32
	}{
		"(0123,4567)": {0x01234567, 0x01234567},
		"(fefe,fefe)": {0xfefefefe, 0xfefefefe},
		"(fexx,fxfx)": {0xfe00f0f0, 0xfeffffff},
	} {
		tag := Element{Tag: input}
		if got := *tag.GetTagLowValue(); got != exp.low {
			t.Errorf("unexpected low value %x for %s", got, input)
		}
		if got := *tag.GetTagHighValue(); got != exp.high {
			t.Errorf("unexpected high value %x for %s", got, input)
		}
	}
}

func TestParseKeyword(t *testing.T) {
	for _, input := range []string{
		"asdf", " a s d f", "as\u200Bdf",
	} {
		tag := Element{Keyword: input}
		if got := tag.GetKeyword(); got != "asdf" {
			t.Errorf("unexpected keyword %q", got)
		}
	}
}

func TestParseVR(t *testing.T) {
	for _, test := range []struct {
		tag   string
		vr    string
		expVR string
	}{
		{
			tag:   "(0010,0020)",
			vr:    "US",
			expVR: "US",
		},
		{
			tag:   "(0010,0020)",
			vr:    "US or OB",
			expVR: "US",
		},
		{
			tag:   "(0010,0020)",
			vr:    "OB or US",
			expVR: "OB",
		},
		{
			tag:   "(FFFE,E0DD)",
			vr:    "See Note 2",
			expVR: "UN",
		},
	} {
		el := Element{
			Tag: test.tag,
			VR:  test.vr,
		}
		if got := el.GetVR(); got != test.expVR {
			t.Errorf("unexpected vr %q for %s/%q",
				got, test.tag, test.vr)
		}
	}
}
