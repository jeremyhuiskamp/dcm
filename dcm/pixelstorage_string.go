// Code generated by "stringer -type PixelStorage"; DO NOT EDIT.

package dcm

import "fmt"

const _PixelStorage_name = "EncapsulatedNative"

var _PixelStorage_index = [...]uint8{0, 12, 18}

func (i PixelStorage) String() string {
	if i < 0 || i >= PixelStorage(len(_PixelStorage_index)-1) {
		return fmt.Sprintf("PixelStorage(%d)", i)
	}
	return _PixelStorage_name[_PixelStorage_index[i]:_PixelStorage_index[i+1]]
}
