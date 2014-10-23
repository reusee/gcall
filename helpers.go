package gcall

import "unsafe"

func New(class string, args ...interface{}) unsafe.Pointer {
	return Call(class+".new", args...)[0].(unsafe.Pointer)
}
