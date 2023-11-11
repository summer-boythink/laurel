package laurel

import (
	"fmt"
	"reflect"
)

func IsZeroPage(arr [PAGE_SIZE]byte) bool {
	// TODO:how to better compare ?
	return reflect.Zero(reflect.TypeOf(arr)).Interface() == arr
}

func SetZeroPage(arr [PAGE_SIZE]byte) error {
	for i := 0; i < len(arr); i++ {
		arr[i] = 0
	}
	return nil
}

func CopyPage(dst *[PAGE_SIZE]byte, src []byte) {
	copy(dst[:], src)
}

func PrintMsgf(msg string, args ...any) {
	ResMsg <- fmt.Sprintf(msg, args...)
}

func indent(level uint32) {
	for i := uint32(0); i < level; i++ {
		PrintMsgf("  ")
	}
}
