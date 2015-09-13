// Copyright Jeremy Huiskamp 2015
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dcm

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestStandardVR(t *testing.T) {
	RegisterTestingT(t)

	// with constant:
	Expect(VRForTag("", PatientID)).To(Equal(LO))
	// with literal:
	Expect(VRForTag("", Tag(0x00100020))).To(Equal(LO))
}

func TestPrivateVR(t *testing.T) {
	RegisterTestingT(t)

	tag := Tag(0x00010001)

	NewDataDictionary("private", map[Tag]ElementSpec{
		tag: ElementSpec{tag: tag, vr: UC},
	})

	Expect(VRForTag("private", tag)).To(Equal(UC))
}
