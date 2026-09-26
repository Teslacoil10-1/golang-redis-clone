package aof

/*
#cgo LDFLAGS: -L../../aof_engine/target/release -laof_engine -lpthread -ldl
#include <stdlib.h>

extern void write_to_aof(const char* command);
*/
import "C"
import "unsafe"

func RecordCommand(cmd string) {
	c_cmd := C.CString(cmd)
	defer C.free(unsafe.Pointer(c_cmd))
	C.write_to_aof(c_cmd)
}
