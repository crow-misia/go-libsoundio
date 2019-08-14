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
	ptr *C.struct_SoundIoChannelArea
}

// fields

// GetBuffer returns base address of buffer.
func (a *ChannelArea) GetBuffer() uintptr {
	return uintptr(unsafe.Pointer(a.ptr.ptr))
}

// GetStep returns ow many bytes it takes to get from the beginning of one sample to
// the beginning of the next sample.
func (a *ChannelArea) GetStep() int {
	return int(a.ptr.step)
}
