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
import (
	"reflect"
	"unsafe"
)

// ChannelArea contain sound data.
type ChannelArea struct {
	buffer         []byte
	step           int
	bytesPerSample int
}

// fields

// Buffer returns buffer.
func (a *ChannelArea) Buffer() []byte {
	return a.buffer
}

func (a *ChannelArea) bufferWithFrame(frame int) []byte {
	offset := frame * a.step
	return a.buffer[offset : offset+a.bytesPerSample]
}

// Step returns ow many bytes it takes to get from the beginning of one sample to
// the beginning of the next sample.
func (a *ChannelArea) Step() int {
	return a.step
}

func newChannelArea(ptr uintptr, format Format, channel int, frameCount int) *ChannelArea {
	size := C.sizeof_struct_SoundIoChannelArea
	areaPtr := ptr + uintptr(channel*size)
	area := (*C.struct_SoundIoChannelArea)(unsafe.Pointer(areaPtr))
	areaStep := int(area.step)
	frameSize := frameCount * areaStep

	sh := &reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(area.ptr)),
		Len:  frameSize,
		Cap:  frameSize,
	}
	buffer := *(*[]byte)(unsafe.Pointer(sh))

	return &ChannelArea{
		buffer:         buffer,
		step:           areaStep,
		bytesPerSample: BytesPerSample(format),
	}
}
