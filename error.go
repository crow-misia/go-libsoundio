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

type Error int

// libsoundio error
const (
	ErrorNone                = Error(C.SoundIoErrorNone)
	ErrorNoMem               = Error(C.SoundIoErrorNoMem)            // Out of memory
	ErrorInitAudioBackend    = Error(C.SoundIoErrorInitAudioBackend) // The backend does not appear to be active or running
	ErrorSystemResources     = Error(C.SoundIoErrorSystemResources)  // A system resource other than memory was not available
	ErrorOpeningDevice       = Error(C.SoundIoErrorOpeningDevice)    // Attempted to open a device and failed
	ErrorNoSuchDevice        = Error(C.SoundIoErrorNoSuchDevice)
	ErrorInvalid             = Error(C.SoundIoErrorInvalid)             // The programmer did not comply with the API
	ErrorBackendUnavailable  = Error(C.SoundIoErrorBackendUnavailable)  // libsoundio was compiled without support for that backend
	ErrorStreaming           = Error(C.SoundIoErrorStreaming)           // An open stream had an error that can only be recovered from by destroying the stream and creating it again
	ErrorIncompatibleDevice  = Error(C.SoundIoErrorIncompatibleDevice)  // Attempted to use a device with parameters it cannot support
	ErrorNoSuchClient        = Error(C.SoundIoErrorNoSuchClient)        // When JACK returns `JackNoSuchClient`
	ErrorIncompatibleBackend = Error(C.SoundIoErrorIncompatibleBackend) // Attempted to use parameters that the backend cannot support.
	ErrorBackendDisconnected = Error(C.SoundIoErrorBackendDisconnected) // Backend server shutdown or became inactive
	ErrorInterrupted         = Error(C.SoundIoErrorInterrupted)
	ErrorUnderflow           = Error(C.SoundIoErrorUnderflow)      // Buffer underrun occurred
	ErrorEncodingString      = Error(C.SoundIoErrorEncodingString) // Unable to convert to or from UTF-8 to the native string format
)

func (e Error) Error() string {
	return C.GoString(C.soundio_strerror(C.int(e)))
}

func convertToError(err C.int) error {
	if err == C.SoundIoErrorNone {
		return nil
	}
	return Error(err)
}
