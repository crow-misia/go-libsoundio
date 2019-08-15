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

type ChannelArea struct {
	ptr uintptr
}

// fields

// GetBuffer returns base address of buffer.
func (a *ChannelArea) GetBuffer() uintptr {
	p := a.getPointer()
	return uintptr(unsafe.Pointer(p.ptr))
}

// GetStep returns ow many bytes it takes to get from the beginning of one sample to
// the beginning of the next sample.
func (a *ChannelArea) GetStep() int {
	p := a.getPointer()
	return int(p.step)
}

func (a *ChannelArea) getPointer() *C.struct_SoundIoChannelArea {
	return (*C.struct_SoundIoChannelArea)(unsafe.Pointer(a.ptr))
}
