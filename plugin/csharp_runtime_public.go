//go:build cgo

package plugin

/*
#include <stdint.h>
*/
import "C"

func CSharpHostContextByID(id uintptr) (*csharpHostContext, bool) {
	return csharpHostContextByID(id)
}

func GoCString(v *C.char) string {
	return goCString(v)
}

func BoolInt(v bool) C.int {
	return boolInt(v)
}

func CString(v string) *C.char {
	return cString(v)
}
