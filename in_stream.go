/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */

package soundio

/*
#include <soundio/soundio.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"sync/atomic"
	"unsafe"
)

type InStream struct {
	ptr              uintptr
	device           *Device
	readCallback     func(*InStream, int, int)
	overflowCallback func(*InStream)
	errorCallback    func(*InStream, error)
}

// fields

// Device returns device to which the stream belongs.
func (s *InStream) Device() *Device {
	return s.device
}

// Format returns format of stream.
func (s *InStream) Format() Format {
	p := s.pointer()
	if p == nil {
		return FormatInvalid
	}
	return Format(p.format)
}

// SetFormat sets format of stream.
func (s *InStream) SetFormat(format Format) {
	p := s.pointer()
	if p == nil {
		return
	}
	p.format = uint32(format)
}

// SampleRate returns sample rate of stream.
func (s *InStream) SampleRate() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(p.sample_rate)
}

// SetSampleRate sets sample rate of stream.
func (s *InStream) SetSampleRate(sampleRate int) {
	p := s.pointer()
	if p == nil {
		return
	}
	p.sample_rate = C.int(sampleRate)
}

// Layout returns layout of stream.
func (s *InStream) Layout() *ChannelLayout {
	p := s.pointer()
	if p == nil {
		return nil
	}
	return newChannelLayout(&p.layout)
}

// SetLayout sets layout of stream.
func (s *InStream) SetLayout(layout *ChannelLayout) {
	if layout == nil || layout.ptr == 0 {
		return
	}
	p := s.pointer()
	if p == nil {
		return
	}
	C.memcpy(unsafe.Pointer(&p.layout), unsafe.Pointer(layout.ptr), C.sizeof_struct_SoundIoChannelLayout)
}

// SoftwareLatency returns software latency of stream.
func (s *InStream) SoftwareLatency() float64 {
	p := s.pointer()
	if p == nil {
		return 0.0
	}
	return float64(p.software_latency)
}

// SetSoftwareLatency sets software latency of stream.
func (s *InStream) SetSoftwareLatency(latency float64) {
	p := s.pointer()
	if p == nil {
		return
	}
	p.software_latency = C.double(latency)
}

// Name returns name of stream.
func (s *InStream) Name() string {
	p := s.pointer()
	if p == nil {
		return ""
	}
	return C.GoString(p.name)
}

// SetName sets name of stream.
func (s *InStream) SetName(name string) {
	p := s.pointer()
	if p == nil {
		return
	}
	if p.name != nil {
		C.free(unsafe.Pointer(p.name))
	}
	p.name = C.CString(name)
}

// NonTerminalHint returns hint that this input stream is nonterminal.
// This is used by JACK and it means that the data received by the stream will be
// passed on or made available to another stream. Defaults to `false`.
func (s *InStream) NonTerminalHint() bool {
	p := s.pointer()
	if p == nil {
		return false
	}
	return bool(p.non_terminal_hint)
}

// BytesPerFrame returns bytes per frame.
func (s *InStream) BytesPerFrame() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(p.bytes_per_frame)
}

// BytesPerSample returns bytes per sample.
func (s *InStream) BytesPerSample() int {
	p := s.pointer()
	if p == nil {
		return 0
	}
	return int(p.bytes_per_sample)
}

// LayoutError returns error If setting the channel layout fails for some reason.
// Possible error:
// * SoundIoErrorIncompatibleDevice
func (s *InStream) LayoutError() error {
	p := s.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(p.layout_error)
}

// SetReadCallback sets ReadCallback.
func (s *InStream) SetReadCallback(callback func(stream *InStream, frameCountMin int, frameCountMax int)) {
	s.readCallback = callback
}

// SetOverflowCallback sets OverflowCallback.
func (s *InStream) SetOverflowCallback(callback func(stream *InStream)) {
	s.overflowCallback = callback
}

// SetErrorCallback sets ErrorCallback.
func (s *InStream) SetErrorCallback(callback func(stream *InStream, err error)) {
	s.errorCallback = callback
}

// functions

// Destroy releases resources.
func (s *InStream) Destroy() {
	ptr := atomic.SwapUintptr(&s.ptr, 0)
	if ptr != 0 {
		p := (*C.struct_SoundIoInStream)(unsafe.Pointer(ptr))
		C.soundio_instream_destroy(p)
	}
}

// Open opens stream.
// After you call this function, SoftwareLatency is set to the correct value.
// The next thing to do is call Start function.
// If this function returns an error, the instream is in an invalid state and
// you must call Destroy function on it.
//
// Possible errors:
// * ErrorInvalid
//   device aim is not DeviceAimInput
//   format is not valid
//   requested layout channel count > MaxChannels
// * ErrorOpeningDevice
// * IoErrorNoMem
// * ErrorBackendDisconnected
// * ErrorSystemResources
// * ErrorNoSuchClient
// * ErrorIncompatibleBackend
// * ErrorIncompatibleDevice
func (s *InStream) Open() error {
	p := s.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(C.soundio_instream_open(p))
}

// Start starts recording.
// After you call this function, ReadCallback will be called.
func (s *InStream) Start() error {
	p := s.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(C.soundio_instream_start(p))
}

// BeginRead called when you are ready to begin reading from the device buffer.
func (s *InStream) BeginRead(frameCount *int) (*ChannelAreas, error) {
	p := s.pointer()
	if p == nil {
		return nil, errorUninitialized
	}

	var ptrs *C.struct_SoundIoChannelArea
	nativeFrameCount := C.int(*frameCount)
	err := convertToError(C.soundio_instream_begin_read(p, &ptrs, &nativeFrameCount))
	*frameCount = int(nativeFrameCount)
	if err != nil {
		return nil, err
	}
	if ptrs == nil {
		return nil, nil
	}
	return newChannelAreas(ptrs, s.Layout().ChannelCount(), *frameCount), nil
}

// EndRead will drop all of the frames from when you called.
func (s *InStream) EndRead() error {
	p := s.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(C.soundio_instream_end_read(p))
}

// Pause pauses the stream and prevents ReadCallback from being called
// If the underlying device supports pausing.
func (s *InStream) Pause(pause bool) error {
	p := s.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(C.soundio_instream_pause(p, C.bool(pause)))
}

// Latency returns the number of seconds that the next frame of sound being
// captured will take to arrive in the buffer, plus the amount of time that is
// represented in the buffer.
// This includes both software and hardware latency.
func (s *InStream) Latency() (float64, error) {
	var latency C.double
	err := convertToError(C.soundio_instream_get_latency(s.pointer(), &latency))
	return float64(latency), err
}

func (s *InStream) pointer() *C.struct_SoundIoInStream {
	if s == nil {
		return nil
	}
	p := atomic.LoadUintptr(&s.ptr)
	if p == 0 {
		return nil
	}
	return (*C.struct_SoundIoInStream)(unsafe.Pointer(p))
}
