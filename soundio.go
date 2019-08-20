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
#include <soundio/soundio.h>
#include <stdlib.h>

extern void soundioOnDevicesChange(struct SoundIo *);
extern void soundioOnBackendDisconnect(struct SoundIo *, int);
extern void soundioOnEventsSignal(struct SoundIo *);

static void setSoundIoCallback(struct SoundIo *io) {
	io->on_devices_change = soundioOnDevicesChange;
	io->on_backend_disconnect = soundioOnBackendDisconnect;
	io->on_events_signal = soundioOnEventsSignal;
}
*/
import "C"
import (
	"sync/atomic"
	"unsafe"
)

const (
	// MaxChannels is suppor channel max count.
	MaxChannels int = C.SOUNDIO_MAX_CHANNELS
)

type SoundIo struct {
	ptr                 uintptr
	appNamePtr          uintptr
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
	p := s.pointer()
	if p == nil {
		return BackendNone
	}
	return Backend(int(p.current_backend))
}

// AppName returns application name.
func (s *SoundIo) AppName() string {
	p := s.pointer()
	if p == nil {
		return ""
	}
	return C.GoString(p.app_name)
}

// SetAppName sets application name.
// PulseAudio uses this for "application name".
// JACK uses this for `client_name`.
// Must not contain a colon (":").
func (s *SoundIo) SetAppName(name string) {
	p := s.pointer()
	if p == nil {
		return
	}
	if s.appNamePtr != 0 {
		C.free(unsafe.Pointer(s.appNamePtr))
	}
	ptr := C.CString(name)
	p.app_name = ptr
	s.appNamePtr = uintptr(unsafe.Pointer(ptr))
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
		ptr: uintptr(unsafe.Pointer(ptr)),
	}
	ptr.userdata = unsafe.Pointer(io)
	C.setSoundIoCallback(ptr)
	return io
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

// Destroy releases resources.
func (s *SoundIo) Destroy() {
	ptr := atomic.SwapUintptr(&s.ptr, 0)

	if ptr != 0 {
		p := (*C.struct_SoundIo)(unsafe.Pointer(ptr))
		C.soundio_destroy(p)

		if s.appNamePtr != 0 {
			C.free(unsafe.Pointer(s.appNamePtr))
		}
	}
}

// Connect tries to connect on all available backends in order.
func (s *SoundIo) Connect() error {
	p := s.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(C.soundio_connect(p))
}

// ConnectBackend connect to backend.
// Instead of calling Connect function you may call this function to try a specific backend.
func (s *SoundIo) ConnectBackend(backend Backend) error {
	p := s.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(C.soundio_connect_backend(p, uint32(backend)))
}

// Disconnect disconnect from backend.
func (s *SoundIo) Disconnect() {
	p := s.pointer()
	if p == nil {
		return
	}
	C.soundio_disconnect(p)
}

// BackendCount returns the number of available backends.
func (s *SoundIo) BackendCount() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(C.soundio_backend_count(p))
}

// Backend returns the available backend at the specified index (0 <= index < BackendCount)
func (s *SoundIo) Backend(index int) Backend {
	p := s.pointer()
	if p == nil {
		return BackendNone
	}
	return Backend(C.soundio_get_backend(p, C.int(index)))
}

// FlushEvents atomically updates information for all connected devices.
func (s *SoundIo) FlushEvents() {
	p := s.pointer()
	if p == nil {
		return
	}
	C.soundio_flush_events(p)
}

// WaitEvents calls FlushEvents then blocks until another event
// is ready or you call Wakeup. Be ready for spurious wakeups.
func (s *SoundIo) WaitEvents() {
	p := s.pointer()
	if p == nil {
		return
	}
	C.soundio_wait_events(p)
}

// Wakeup makes WaitEvents stop blocking.
func (s *SoundIo) Wakeup() {
	p := s.pointer()
	if p == nil {
		return
	}
	C.soundio_wakeup(p)
}

// ForceDeviceScan rescan device If necessary.
func (s *SoundIo) ForceDeviceScan() {
	p := s.pointer()
	if p == nil {
		return
	}
	C.soundio_force_device_scan(p)
}

// InputDeviceCount returns the number of input devices.
// Returns -1 if you never called FlushEvents.
func (s *SoundIo) InputDeviceCount() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(C.soundio_input_device_count(p))
}

// OutputDeviceCount returns the number of output devices.
// Returns -1 if you never called FlushEvents.
func (s *SoundIo) OutputDeviceCount() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(C.soundio_output_device_count(p))
}

// InputDevice returns a device.
// Call RemoveReference when done.
// `index` must be 0 <= index < InputDeviceCount.
func (s *SoundIo) InputDevice(index int) (*Device, error) {
	p := s.pointer()
	if p == nil {
		return nil, errorUninitialized
	}
	return newDevice(C.soundio_get_input_device(p, C.int(index))), nil
}

// OutputDevice returns a device.
// Call RemoveReference when done.
// `index` must be 0 <= index < OutputDeviceCount
func (s *SoundIo) OutputDevice(index int) (*Device, error) {
	p := s.pointer()
	if p == nil {
		return nil, errorUninitialized
	}
	return newDevice(C.soundio_get_output_device(p, C.int(index))), nil
}

// DefaultInputDeviceIndex returns the index of the default input device
// returns -1 if there are no devices or if you never called FlushEvents.
func (s *SoundIo) DefaultInputDeviceIndex() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(C.soundio_default_input_device_index(p))
}

// DefaultOutputDeviceIndex returns the index of the default output device
// returns -1 if there are no devices or if you never called FlushEvents.
func (s *SoundIo) DefaultOutputDeviceIndex() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(C.soundio_default_output_device_index(p))
}

func (s *SoundIo) pointer() *C.struct_SoundIo {
	if s == nil {
		return nil
	}
	p := atomic.LoadUintptr(&s.ptr)
	if p == 0 {
		return nil
	}
	return (*C.struct_SoundIo)(unsafe.Pointer(p))
}
