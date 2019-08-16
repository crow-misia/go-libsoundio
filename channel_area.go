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

type ChannelArea struct {
	ptr       uintptr
	step      int
	frameSize int
}

// fields

// GetBuffer returns base address of buffer.
func (a *ChannelArea) GetBuffer() []byte {
	size := a.frameSize

	sh := &reflect.SliceHeader{
		Data: a.ptr,
		Len:  size,
		Cap:  size,
	}

	return *(*[]byte)(unsafe.Pointer(sh))
}

// GetStep returns ow many bytes it takes to get from the beginning of one sample to
// the beginning of the next sample.
func (a *ChannelArea) GetStep() int {
	return a.step
}

func newChannelArea(areas *ChannelAreas, channel int) *ChannelArea {
	size := C.sizeof_struct_SoundIoChannelArea
	ptr := areas.ptr + uintptr(channel*size)
	area := (*C.struct_SoundIoChannelArea)(unsafe.Pointer(ptr))
	areaStep := int(area.step)

	return &ChannelArea{
		ptr:       uintptr(unsafe.Pointer(area.ptr)),
		step:      areaStep,
		frameSize: areas.frameCount * areaStep,
	}
}
