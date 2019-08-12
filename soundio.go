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
import "unsafe"

const (
	MaxChannels int = C.SOUNDIO_MAX_CHANNELS
)

type SoundIo struct {
	ptr                 *C.struct_SoundIo
	appNamePtr          *C.char
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

func (s *SoundIo) GetCurrentBackend() Backend {
	return Backend(int(s.ptr.current_backend))
}

func (s *SoundIo) GetAppName() string {
	return C.GoString(s.ptr.app_name)
}

func (s *SoundIo) SetAppName(name string) {
	if s.appNamePtr != nil {
		C.free(unsafe.Pointer(s.appNamePtr))
	}
	s.ptr.app_name = C.CString(name)
}

// functions

// Version ... the version number string of libsoundio.
func Version() string {
	return C.GoString(C.soundio_version_string())
}

// VersionMajor ... the major version number of libsoundio.
func VersionMajor() int {
	return int(C.soundio_version_major())
}

// VersionMinor ... the minor version number of libsoundio.
func VersionMinor() int {
	return int(C.soundio_version_minor())
}

// VersionPatch ... the patch version number of libsoundio.
func VersionPatch() int {
	return int(C.soundio_version_patch())
}

func GetBytesPerSample(format Format) int {
	return int(C.soundio_get_bytes_per_sample(uint32(format)))
}

func GetBytesPerFrame(format Format, channelCount int) int {
	return int(C.soundio_get_bytes_per_frame(uint32(format), C.int(channelCount)))
}

func GetBytesPerSecond(format Format, channelCount int, sampleRate int) int {
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
	return io
}

func (s *SoundIo) Destroy() {
	if s.appNamePtr != nil {
		C.free(unsafe.Pointer(s.appNamePtr))
	}
	s.appNamePtr = nil

	if s.ptr != nil {
		s.ptr.userdata = nil
		C.soundio_destroy(s.ptr)
	}
	s.ptr = nil
}

// Tries ::soundio_connect_backend on all available backends in order.
func (s *SoundIo) Connect() error {
	return convertToError(C.soundio_connect(s.ptr))
}

// Instead of calling ::soundio_connect you may call this function to try a specific backend.
func (s *SoundIo) ConnectBackend(backend Backend) error {
	return convertToError(C.soundio_connect_backend(s.ptr, uint32(backend)))
}

func (s *SoundIo) Disconnect() {
	C.soundio_disconnect(s.ptr)
}

// Returns the number of available backends.
func (s *SoundIo) BackendCount() int {
	return int(C.soundio_backend_count(s.ptr))
}

// get the available backend at the specified index (0 <= index < ::soundio_backend_count)
func (s *SoundIo) GetBackend(index int) Backend {
	return Backend(C.soundio_get_backend(s.ptr, C.int(index)))
}

// Atomically update information for all connected devices.
func (s *SoundIo) FlushEvents() {
	C.soundio_flush_events(s.ptr)
}

// This function calls FlushEvents then blocks until another event
// is ready or you call Wakeup. Be ready for spurious wakeups.
func (s *SoundIo) WaitEvents() {
	C.soundio_wait_events(s.ptr)
}

// Makes WaitEvents stop blocking.
func (s *SoundIo) Wakeup() {
	C.soundio_wakeup(s.ptr)
}

func (s *SoundIo) ForceDeviceScan() {
	C.soundio_force_device_scan(s.ptr)
}

func (s *SoundIo) InputDeviceCount() int {
	return int(C.soundio_input_device_count(s.ptr))
}

func (s *SoundIo) OutputDeviceCount() int {
	return int(C.soundio_output_device_count(s.ptr))
}

func (s *SoundIo) GetInputDevice(index int) *Device {
	return &Device{
		ptr: C.soundio_get_input_device(s.ptr, C.int(index)),
	}
}

func (s *SoundIo) GetOutputDevice(index int) *Device {
	return &Device{
		ptr: C.soundio_get_output_device(s.ptr, C.int(index)),
	}
}

func (s *SoundIo) DefaultInputDeviceIndex() int {
	return int(C.soundio_default_input_device_index(s.ptr))
}

func (s *SoundIo) DefaultOutputDeviceIndex() int {
	return int(C.soundio_default_output_device_index(s.ptr))
}

func (s *SoundIo) RingBufferCreate(requestedCapacity int) *RingBuffer {
	return &RingBuffer{
		ptr: C.soundio_ring_buffer_create(s.ptr, C.int(requestedCapacity)),
	}
}
