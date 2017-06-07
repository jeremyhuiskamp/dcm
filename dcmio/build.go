package dcmio

import (
	"io/ioutil"

	"github.com/jeremyhuiskamp/dcm/dcm"
)

func Build(parser Parser) (obj dcm.Object, err error) {
	obj = dcm.NewObject()

	for {
		tag, err := parser.NextTag()
		if err != nil || tag == nil || dcm.ItemDelimitationItem == tag.Tag {
			return obj, err
		}

		if tag.Tag.IsGroupLength() {
			// group length
			continue
		}

		vr := tag.VR
		if vr == nil {
			// TODO: private creator support
			tagvr := dcm.VRForTag("", tag.Tag)
			if !dcm.VREq(&tagvr, &dcm.UN) {
				vr = &tagvr
			}
		}

		if vr == nil && tag.ValueLength == -1 {
			vr = &dcm.SQ
		}

		if dcm.VREq(vr, &dcm.SQ) {
			sq := dcm.SequenceElement{
				Tag: tag.Tag,
			}

			obj.Put(sq)

			// TODO iterate through items
		} else {
			data, err := ioutil.ReadAll(tag.Value)
			if err != nil {
				return obj, err
			}

			el := dcm.SimpleElement{
				Tag:  tag.Tag,
				VR:   *vr,
				Data: data,
			}

			obj.Put(el)
		}
	}
}
