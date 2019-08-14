/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */

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

// GetChannelCount returns channel count.
func (a *ChannelAreas) GetChannelCount() int {
	return a.channelCount
}

// GetFrameCount returns frame count.
func (a *ChannelAreas) GetFrameCount() int {
	return a.frameCount
}

// GetArea returns ChannelArea.
func (a *ChannelAreas) GetArea(channel int) *ChannelArea {
	size := unsafe.Sizeof(*a.ptr)
	return &ChannelArea{
		ptr: (*C.struct_SoundIoChannelArea)(unsafe.Pointer(uintptr(unsafe.Pointer(a.ptr)) + uintptr(channel)*size)),
	}
}
