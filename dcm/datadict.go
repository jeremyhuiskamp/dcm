// Copyright Jeremy Huiskamp 2015
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package dcm

import "sync"

// Specifications for data dictionaries with support for the standard dictionary
// (generated into stddict.go) and private creator dictionaries.

type ElementSpec struct {
	tag      Tag
	maxValue Tag
	vr       VR
	minVM    int
	maxVM    int
	retired  bool
	desc     string
	keyword  string
}

func (e ElementSpec) GetDesc() string {
	return e.desc
}

type DataDictionary struct {
	specsByTag        map[Tag]ElementSpec
	specsByName       map[string]ElementSpec
	privateCreatorUID string
}

func (dd *DataDictionary) FindElementSpec(tag Tag) *ElementSpec {
	if spec, ok := dd.specsByTag[tag]; ok {
		return &spec
	}

	return nil
}

func (dd *DataDictionary) GetPrivateCreatorUID() string {
	return dd.privateCreatorUID
}

func (dd *DataDictionary) FindElementSpecByName(name string) *ElementSpec {
	if spec, ok := dd.specsByName[name]; ok {
		return &spec
	}

	return nil
}

var (
	// privateDicts maps a privateCreatorUID to a private dictionary.
	// It's probably ok for this to be global, because there should only
	// ever be one static dictionary definition per private creator uid.
	privateDicts    = make(map[string]DataDictionary)
	privateDictsMtx sync.Mutex
)

func NewDataDictionary(privateCreatorUID string, specs map[Tag]ElementSpec) DataDictionary {
	specsByName := make(map[string]ElementSpec, len(specs))
	// TODO: populate specsByName??

	dd := DataDictionary{
		specsByTag:        specs,
		specsByName:       specsByName,
		privateCreatorUID: privateCreatorUID,
	}

	if privateCreatorUID != "" {
		privateDictsMtx.Lock()
		defer privateDictsMtx.Unlock()
		privateDicts[privateCreatorUID] = dd
	}

	return dd
}

func GetPrivateDictionary(creatorUID string) *DataDictionary {
	privateDictsMtx.Lock()
	defer privateDictsMtx.Unlock()
	if dict, ok := privateDicts[creatorUID]; ok {
		return &dict
	}
	return nil
}

func SpecForTag(privateCreatorUID string, tag Tag) *ElementSpec {
	var dd *DataDictionary

	if privateCreatorUID == "" {
		dd = &stddict
	} else {
		dd = GetPrivateDictionary(privateCreatorUID)
	}

	if dd == nil {
		return nil
	}

	return dd.FindElementSpec(tag)
}

func VRForTag(privateCreatorUID string, tag Tag) VR {
	spec := SpecForTag(privateCreatorUID, tag)

	if spec == nil {
		return UN
	}

	return spec.vr
}
