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
	areas        []*ChannelArea
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
	return a.areas[channel]
}

// GetBuffer returns ChannelArea buffer.
func (a *ChannelAreas) GetBuffer(channel int, frame int) []byte {
	return a.areas[channel].getBuffer(frame)

}
func newChannelAreas(ptr *C.struct_SoundIoChannelArea, chanelCount int, frameCount int) *ChannelAreas {
	areasPtr := uintptr(unsafe.Pointer(ptr))
	areas := make([]*ChannelArea, chanelCount)

	for ch := 0; ch < chanelCount; ch++ {
		areas[ch] = newChannelArea(areasPtr, ch, frameCount)
	}

	return &ChannelAreas{
		areas:        areas,
		channelCount: chanelCount,
		frameCount:   frameCount,
	}
}
