package soundio

/*
#include <soundio/soundio.h>
*/
import "C"

type ErrorCode int32

const (
	ErrorNone                ErrorCode = C.SoundIoErrorNone
	ErrorNoMem                         = C.SoundIoErrorNoMem            // Out of memory
	ErrorInitAudioBackend              = C.SoundIoErrorInitAudioBackend // The backend does not appear to be active or running
	ErrorSystemResources               = C.SoundIoErrorSystemResources  // A system resource other than memory was not available
	ErrorOpeningDevice                 = C.SoundIoErrorOpeningDevice    // Attempted to open a device and failed
	ErrorNoSuchDevice                  = C.SoundIoErrorNoSuchDevice
	ErrorInvalid                       = C.SoundIoErrorInvalid             // The programmer did not comply with the API
	ErrorBackendUnavailable            = C.SoundIoErrorBackendUnavailable  // libsoundio was compiled without support for that backend
	ErrorStreaming                     = C.SoundIoErrorStreaming           // An open stream had an error that can only be recovered from by destroying the stream and creating it again
	ErrorIncompatibleDevice            = C.SoundIoErrorIncompatibleDevice  // Attempted to use a device with parameters it cannot support
	ErrorNoSuchClient                  = C.SoundIoErrorNoSuchClient        // When JACK returns `JackNoSuchClient`
	ErrorIncompatibleBackend           = C.SoundIoErrorIncompatibleBackend // Attempted to use parameters that the backend cannot support.
	ErrorBackendDisconnected           = C.SoundIoErrorBackendDisconnected // Backend server shutdown or became inactive
	ErrorInterrupted                   = C.SoundIoErrorInterrupted
	ErrorUnderflow                     = C.SoundIoErrorUnderflow      // Buffer underrun occurred
	ErrorEncodingString                = C.SoundIoErrorEncodingString // Unable to convert to or from UTF-8 to the native string format
)

type Error struct {
	Code ErrorCode
}

func (err *Error) Error() string {
	return C.GoString(C.soundio_strerror(C.int(err.Code)))
}

func convertToError(err C.int) error {
	code := ErrorCode(err)
	if code != ErrorNone {
		return &Error{Code: code}
	}
	return nil
}
