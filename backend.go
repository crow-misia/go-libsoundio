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

type Backend uint32

const (
	BackendNone       Backend = C.SoundIoBackendNone       // None
	BackendJack               = C.SoundIoBackendJack       // Jack
	BackendPulseAudio         = C.SoundIoBackendPulseAudio // PulseAudio
	BackendAlsa               = C.SoundIoBackendAlsa       // ALSA
	BackendCoreAudio          = C.SoundIoBackendCoreAudio  // CoreAudio
	BackendWasapi             = C.SoundIoBackendWasapi     // WASAPI
	BackendDummy              = C.SoundIoBackendDummy      // Dummy
)

func (b Backend) String() string {
	return C.GoString(C.soundio_backend_name(uint32(b)))
}

// functions

// Have returns whether libsoundio was compiled with backend.
func (b Backend) Have() bool {
	return HaveBackend(b)
}

// HaveBackend returns whether libsoundio was compiled with backend.
func HaveBackend(backend Backend) bool {
	return bool(C.soundio_have_backend(uint32(backend)))
}
