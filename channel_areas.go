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

type ChannelAreas struct {
	ptr          uintptr
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
	size := C.sizeof_struct_SoundIoChannelArea
	return &ChannelArea{
		ptr: a.ptr + uintptr(channel*size),
	}
}
