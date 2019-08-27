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

// Backend type.
type Backend uint32

// Backend enumeration.
const (
	BackendNone       = Backend(C.SoundIoBackendNone)       // None
	BackendJack       = Backend(C.SoundIoBackendJack)       // Jack
	BackendPulseAudio = Backend(C.SoundIoBackendPulseAudio) // PulseAudio
	BackendAlsa       = Backend(C.SoundIoBackendAlsa)       // ALSA
	BackendCoreAudio  = Backend(C.SoundIoBackendCoreAudio)  // CoreAudio
	BackendWasapi     = Backend(C.SoundIoBackendWasapi)     // WASAPI
	BackendDummy      = Backend(C.SoundIoBackendDummy)      // Dummy
)

func (b Backend) String() string {
	return C.GoString(C.soundio_backend_name(uint32(b)))
}

// functions

// Have returns whether libsoundio was compiled with backend.
func (b Backend) Have() bool {
	return bool(C.soundio_have_backend(uint32(b)))
}
