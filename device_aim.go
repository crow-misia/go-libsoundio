package soundio

/*
#include <soundio/soundio.h>
*/
import "C"

type DeviceAim uint32

const (
	DeviceAimInput  DeviceAim = C.SoundIoDeviceAimInput  // capture / recording
	DeviceAimOutput           = C.SoundIoDeviceAimOutput // playback
)

func (a DeviceAim) String() string {
	switch a {
	case DeviceAimInput:
		return "Input"
	case DeviceAimOutput:
		return "Output"
	default:
		return ""
	}
}
