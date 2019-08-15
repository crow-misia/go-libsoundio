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

type SampleRateRange struct {
	ptr uintptr
}

// fields

// GetMin returns sample rate minimal.
func (r *SampleRateRange) GetMin() int {
	p := r.getPointer()
	return int(p.min)
}

// GetMax returns sample rate maximal.
func (r *SampleRateRange) GetMax() int {
	p := r.getPointer()
	return int(p.max)
}

func (r *SampleRateRange) getPointer() *C.struct_SoundIoSampleRateRange {
	return (*C.struct_SoundIoSampleRateRange)(unsafe.Pointer(r.ptr))
}
