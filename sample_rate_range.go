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

type SampleRateRange struct {
	ptr *C.struct_SoundIoSampleRateRange
}

// fields

// GetMin returns sample rate minimal.
func (r *SampleRateRange) GetMin() int {
	return int(r.ptr.min)
}

// GetMax returns sample rate maximal.
func (r *SampleRateRange) GetMax() int {
	return int(r.ptr.max)
}
