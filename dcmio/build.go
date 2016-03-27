package dcmio

import (
	"github.com/jeremyhuiskamp/dcm/dcm"
	"io/ioutil"
	"log"
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

		log.Printf("Read tag %s with vr %s\n", tag.Tag, vr.Name)

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

			log.Printf("Read %d bytes value\n", len(data))

			el := dcm.SimpleElement{
				Tag:  tag.Tag,
				VR:   *vr,
				Data: data,
			}

			obj.Put(el)
		}
	}
}
