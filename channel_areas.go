package soundio

/*
#include <soundio/soundio.h>
*/
import "C"
import "unsafe"

type ChannelAreas struct {
	ptr          *C.struct_SoundIoChannelArea
	channelCount int
	frameCount   int
}

func (a *ChannelAreas) GetChannelCount() int {
	return a.channelCount
}

func (a *ChannelAreas) GetFrameCount() int {
	return a.frameCount
}

func (a *ChannelAreas) GetArea(channel int) *ChannelArea {
	size := unsafe.Sizeof(*a.ptr)
	return &ChannelArea{
		ptr: (*C.struct_SoundIoChannelArea)(unsafe.Pointer(uintptr(unsafe.Pointer(a.ptr)) + uintptr(channel)*size)),
	}
}
