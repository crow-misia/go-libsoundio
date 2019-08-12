package soundio

/*
#include <soundio/soundio.h>
*/
import "C"

type Backend uint32

const (
	BackendNone       Backend = C.SoundIoBackendNone
	BackendJack               = C.SoundIoBackendJack
	BackendPulseAudio         = C.SoundIoBackendPulseAudio
	BackendAlsa               = C.SoundIoBackendAlsa
	BackendCoreAudio          = C.SoundIoBackendCoreAudio
	BackendWasapi             = C.SoundIoBackendWasapi
	BackendDummy              = C.SoundIoBackendDummy
)

// Get a string representation of a #SoundIoBackend
func (b Backend) String() string {
	return C.GoString(C.soundio_backend_name(uint32(b)))
}

// functions

func (b Backend) Have() bool {
	return HaveBackend(b)
}

func HaveBackend(backend Backend) bool {
	return bool(C.soundio_have_backend(uint32(backend)))
}
