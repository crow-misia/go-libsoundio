/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */

// Package soundio is as set of bindings for the libsoundio sound library.
package soundio

/*
#cgo LDFLAGS: -lsoundio -lm
#include "soundio.h"
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"runtime"
	"unsafe"
)

const (
	// MaxChannels is support channel max count.
	MaxChannels int = C.SOUNDIO_MAX_CHANNELS
)

// SoundIo is used for selecting and initializing the relevant backends.
type SoundIo struct {
	backend             Backend
	ptr                 *C.struct_SoundIo
	appName             string
	onDevicesChange     func(*SoundIo)
	onBackendDisconnect func(*SoundIo, error)
	onEventsSignal      func(*SoundIo)
}

//export soundioOnDevicesChange
func soundioOnDevicesChange(nativeIo *C.struct_SoundIo) {
	io := (*SoundIo)(nativeIo.userdata)
	if io.onDevicesChange != nil {
		io.onDevicesChange(io)
	}
}

//export soundioOnBackendDisconnect
func soundioOnBackendDisconnect(nativeIo *C.struct_SoundIo, err C.int) {
	io := (*SoundIo)(nativeIo.userdata)
	if io.onBackendDisconnect != nil {
		io.onBackendDisconnect(io, convertToError(err))
	}
}

//export soundioOnEventsSignal
func soundioOnEventsSignal(nativeIo *C.struct_SoundIo) {
	io := (*SoundIo)(nativeIo.userdata)
	if io.onEventsSignal != nil {
		io.onEventsSignal(io)
	}
}

// fields

// CurrentBackend returns current backend.
func (s *SoundIo) CurrentBackend() Backend {
	return Backend(int(s.ptr.current_backend))
}

// AppName returns application name.
func (s *SoundIo) AppName() string {
	return C.GoString(s.ptr.app_name)
}

// functions

// Version returns the version number string of libsoundio.
func Version() string {
	return C.GoString(C.soundio_version_string())
}

// VersionMajor returns the major version number of libsoundio.
func VersionMajor() int {
	return int(C.soundio_version_major())
}

// VersionMinor returns the minor version number of libsoundio.
func VersionMinor() int {
	return int(C.soundio_version_minor())
}

// VersionPatch returns the patch version number of libsoundio.
func VersionPatch() int {
	return int(C.soundio_version_patch())
}

// BytesPerSample returns bytes per sample.
// Returns -1 on invalid format.
func BytesPerSample(format Format) int {
	return int(C.soundio_get_bytes_per_sample(uint32(format)))
}

// BytesPerFrame returns bytes per frame.
// A frame is one sample per channel.
func BytesPerFrame(format Format, channelCount int) int {
	return int(C.soundio_get_bytes_per_frame(uint32(format), C.int(channelCount)))
}

// BytesPerSecond returns bytes per second.
// Sample rate is the number of frames per second.
func BytesPerSecond(format Format, channelCount int, sampleRate int) int {
	return int(C.soundio_get_bytes_per_second(uint32(format), C.int(channelCount), C.int(sampleRate)))
}

// Create a SoundIo context. You may create multiple instances of this to connect to multiple backends. Sets all fields to defaults.
func Create(opts ...Option) *SoundIo {
	ptr := C.soundio_create()
	io := &SoundIo{
		backend: BackendNone,
		ptr:     ptr,
		appName: "SoundIo",
	}
	ptr.userdata = unsafe.Pointer(io)
	C.setSoundIoCallback(ptr)

	for _, opt := range opts {
		opt(io)
	}
	io.ptr.app_name = C.CString(io.appName)

	runtime.SetFinalizer(io, destroySoundIo)

	return io
}

// destroySoundIo releases resources.
func destroySoundIo(s *SoundIo) {
	if s.ptr != nil {
		C.free(unsafe.Pointer(s.ptr.app_name))
		C.soundio_destroy(s.ptr)
		s.ptr = nil
	}
}

// functions

// Connect tries to connect on all available backends in order.
func (s *SoundIo) Connect() error {
	var err error
	if s.backend == BackendNone {
		err = convertToError(C.soundio_connect(s.ptr))
	} else {
		err = convertToError(C.soundio_connect_backend(s.ptr, uint32(s.backend)))
	}

	if err == nil {
		s.FlushEvents()
	}

	return err
}

// Disconnect disconnect from backend.
func (s *SoundIo) Disconnect() {
	C.soundio_disconnect(s.ptr)
}

// BackendCount returns the number of available backends.
func (s *SoundIo) BackendCount() int {
	return int(C.soundio_backend_count(s.ptr))
}

// Backend returns the available backend at the specified index (0 <= index < BackendCount)
func (s *SoundIo) Backend(index int) Backend {
	return Backend(C.soundio_get_backend(s.ptr, C.int(index)))
}

// FlushEvents atomically updates information for all connected devices.
func (s *SoundIo) FlushEvents() {
	C.soundio_flush_events(s.ptr)
}

// WaitEvents calls FlushEvents then blocks until context canceled.
func (s *SoundIo) WaitEvents(ctx context.Context) error {
	go func() {
		select {
		case <-ctx.Done():
			if s.CurrentBackend() != BackendNone {
				C.soundio_wakeup(s.ptr)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if s.CurrentBackend() != BackendNone {
				C.soundio_wait_events(s.ptr)
			} else {
				return ctx.Err()
			}
		}
	}
}

// ForceDeviceScan rescan device If necessary.
func (s *SoundIo) ForceDeviceScan() {
	C.soundio_force_device_scan(s.ptr)
}

// InputDeviceCount returns the number of input devices.
// Returns -1 if you never called FlushEvents.
func (s *SoundIo) InputDeviceCount() int {
	return int(C.soundio_input_device_count(s.ptr))
}

// OutputDeviceCount returns the number of output devices.
// Returns -1 if you never called FlushEvents.
func (s *SoundIo) OutputDeviceCount() int {
	return int(C.soundio_output_device_count(s.ptr))
}

// InputDevice returns a device.
// Call RemoveReference when done.
// `index` must be 0 <= index < InputDeviceCount.
func (s *SoundIo) InputDevice(index int) *Device {
	return newDevice(C.soundio_get_input_device(s.ptr, C.int(index)))
}

// OutputDevice returns a device.
// Call RemoveReference when done.
// `index` must be 0 <= index < OutputDeviceCount
func (s *SoundIo) OutputDevice(index int) *Device {
	return newDevice(C.soundio_get_output_device(s.ptr, C.int(index)))
}

// DefaultInputDeviceIndex returns the index of the default input device
// returns -1 if there are no devices or if you never called FlushEvents.
func (s *SoundIo) DefaultInputDeviceIndex() int {
	return int(C.soundio_default_input_device_index(s.ptr))
}

// DefaultOutputDeviceIndex returns the index of the default output device
// returns -1 if there are no devices or if you never called FlushEvents.
func (s *SoundIo) DefaultOutputDeviceIndex() int {
	return int(C.soundio_default_output_device_index(s.ptr))
}
