package soundio

/*
#include <soundio/soundio.h>
*/
import "C"

type SampleRateRange struct {
	ptr *C.struct_SoundIoSampleRateRange
}

// fields

func (r *SampleRateRange) GetMin() int {
	return int(r.ptr.min)
}

func (r *SampleRateRange) GetMax() int {
	return int(r.ptr.max)
}
