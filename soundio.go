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
	"runtime"
	"unsafe"
)

const (
	// MaxChannels is suppor channel max count.
	MaxChannels int = C.SOUNDIO_MAX_CHANNELS
)

// SoundIo is used for selecting and initializing the relevant backends.
type SoundIo struct {
	ptr                 *C.struct_SoundIo
	appNamePtr          unsafe.Pointer
	onDevicesChange     func(io *SoundIo)
	onBackendDisconnect func(io *SoundIo, err error)
	onEventsSignal      func(io *SoundIo)
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

// SetAppName sets application name.
// PulseAudio uses this for "application name".
// JACK uses this for `client_name`.
// Must not contain a colon (":").
func (s *SoundIo) SetAppName(name string) {
	newPtr := C.CString(name)
	oldPtr := s.appNamePtr
	if oldPtr != nil {
		C.free(oldPtr)
	}
	s.appNamePtr = unsafe.Pointer(newPtr)
	s.ptr.app_name = newPtr
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
func Create() *SoundIo {
	ptr := C.soundio_create()
	io := &SoundIo{
		ptr: ptr,
	}
	ptr.userdata = unsafe.Pointer(io)
	C.setSoundIoCallback(ptr)

	runtime.SetFinalizer(io, destroySoundIo)
	return io
}

// destroySoundIo releases resources.
func destroySoundIo(s *SoundIo) {
	if s.appNamePtr != nil {
		C.free(unsafe.Pointer(s.appNamePtr))
		s.ptr = nil
	}
	if s.ptr != nil {
		C.soundio_destroy(s.ptr)
		s.ptr = nil
	}
}

// fields

// SetOnDevicesChange is onDeviceChange callback setter.
func (s *SoundIo) SetOnDevicesChange(callback func(*SoundIo)) {
	s.onDevicesChange = callback
}

// SetOnBackendDisconnect is onBackendDisconnect callback setter.
func (s *SoundIo) SetOnBackendDisconnect(callback func(*SoundIo, error)) {
	s.onBackendDisconnect = callback
}

// SetOnEventsSignal is onEventsSignal callback setter.
func (s *SoundIo) SetOnEventsSignal(callback func(*SoundIo)) {
	s.onEventsSignal = callback
}

// functions

// Connect tries to connect on all available backends in order.
func (s *SoundIo) Connect() error {
	return convertToError(C.soundio_connect(s.ptr))
}

// ConnectBackend connect to backend.
// Instead of calling Connect function you may call this function to try a specific backend.
func (s *SoundIo) ConnectBackend(backend Backend) error {
	return convertToError(C.soundio_connect_backend(s.ptr, uint32(backend)))
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

// WaitEvents calls FlushEvents then blocks until another event
// is ready or you call Wakeup. Be ready for spurious wakeups.
func (s *SoundIo) WaitEvents() {
	C.soundio_wait_events(s.ptr)
}

// Wakeup makes WaitEvents stop blocking.
func (s *SoundIo) Wakeup() {
	C.soundio_wakeup(s.ptr)
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
