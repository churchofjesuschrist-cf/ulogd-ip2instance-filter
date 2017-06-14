// +build diego

package plugin

import "C"

func registrationName() *C.char {
	// TODO: remember to C.free() the returned value in the consumer
	return C.CString("DIEGOINSTANCE")
}
