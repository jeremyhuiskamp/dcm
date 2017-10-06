// Copyright Jeremy Huiskamp 2015
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dcm

import "testing"

func TestVRForTag(t *testing.T) {
	privateTag := Tag(0x00010001)
	NewDataDictionary("private", map[Tag]ElementSpec{
		privateTag: {tag: privateTag, vr: UC},
	})

	for _, test := range []struct {
		privateCreatorUID string
		tag               Tag
		vr                VR
	}{
		{"", PatientID, LO},
		{"", privateTag, UN},
		{"private", privateTag, UC},
		{"private", PatientID, UN},
	} {
		if vr := VRForTag(test.privateCreatorUID, test.tag); vr != test.vr {
			t.Errorf("unexpected vr %s for %s/%s",
				vr, test.privateCreatorUID, test.tag)
		}
	}

}
