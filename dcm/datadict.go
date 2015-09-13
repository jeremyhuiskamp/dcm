// Copyright Jeremy Huiskamp 2015
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package dcm

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

func NewDataDictionary(privateCreatorUID string, specs map[Tag]ElementSpec) DataDictionary {
	specsByName := make(map[string]ElementSpec, len(specs))

	dd := DataDictionary{
		specsByTag:        specs,
		specsByName:       specsByName,
		privateCreatorUID: privateCreatorUID,
	}
	
	if privateCreatorUID != "" {
		privateDicts[privateCreatorUID] = dd
	}

	return dd
}

var privateDicts = make(map[string]DataDictionary)

func GetPrivateDictionary(creatorUID string) *DataDictionary {
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
