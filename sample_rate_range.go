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
	"sync/atomic"
	"unsafe"
)

type SampleRateRange struct {
	ptr uintptr
}

// fields

// Min returns sample rate minimal.
func (r *SampleRateRange) Min() int {
	p := r.pointer()
	if p == nil {
		return 0
	}
	return int(p.min)
}

// Max returns sample rate maximal.
func (r *SampleRateRange) Max() int {
	p := r.pointer()
	if p == nil {
		return 0
	}
	return int(p.max)
}

func (r *SampleRateRange) pointer() *C.struct_SoundIoSampleRateRange {
	if r == nil {
		return nil
	}
	p := atomic.LoadUintptr(&r.ptr)
	if p == 0 {
		return nil
	}
	return (*C.struct_SoundIoSampleRateRange)(unsafe.Pointer(p))
}
