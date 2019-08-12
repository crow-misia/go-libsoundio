package soundio

/*
#include <soundio/soundio.h>
*/
import "C"
import "unsafe"

type ChannelArea struct {
	ptr *C.struct_SoundIoChannelArea
}

// fields

func (a *ChannelArea) GetBuffer() uintptr {
	return uintptr(unsafe.Pointer(a.ptr.ptr))
}

func (a *ChannelArea) GetStep() int {
	return int(a.ptr.step)
}
