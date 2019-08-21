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
	min int
	max int
}

// fields

// Min returns sample rate minimal.
func (r *SampleRateRange) Min() int {
	if r == nil {
		return 0
	}
	return r.min
}

// Max returns sample rate maximal.
func (r *SampleRateRange) Max() int {
	if r == nil {
		return 0
	}
	return r.max
}

func newSampleRateRange(p uintptr) SampleRateRange {
	r := (*C.struct_SoundIoSampleRateRange)(unsafe.Pointer(p))
	return SampleRateRange{
		min: int(r.min),
		max: int(r.max),
	}
}
