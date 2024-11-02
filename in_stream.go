/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */

package soundio

/*
#include "soundio.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

// InStream is Input Stream.
type InStream struct {
	p                uintptr
	d                *Device
	readCallback     func(*InStream, int, int)
	overflowCallback func(*InStream)
	errorCallback    func(*InStream, error)
}

// InStreamConfig is config of input stream.
type InStreamConfig struct {
	// Format
	Format Format
	// SampleRate
	SampleRate int
	// Layout
	Layout *ChannelLayout
	// SoftwareLatency
	SoftwareLatency float64
	// Name
	Name string
}

//export instreamReadCallbackDelegate
func instreamReadCallbackDelegate(nativeStream *C.struct_SoundIoInStream, frameCountMin C.int, frameCountMax C.int) {
	stream := (*InStream)(nativeStream.userdata)
	if stream.readCallback != nil {
		stream.readCallback(stream, int(frameCountMin), int(frameCountMax))
	}
}

//export instreamOverflowCallbackDelegate
func instreamOverflowCallbackDelegate(nativeStream *C.struct_SoundIoInStream) {
	stream := (*InStream)(nativeStream.userdata)
	if stream.overflowCallback != nil {
		stream.overflowCallback(stream)
	}
}

//export instreamErrorCallbackDelegate
func instreamErrorCallbackDelegate(nativeStream *C.struct_SoundIoInStream, err C.int) {
	stream := (*InStream)(nativeStream.userdata)
	if stream.errorCallback != nil {
		stream.errorCallback(stream, convertToError(err))
	}
}

// fields

// Device returns device to which the stream belongs.
func (s *InStream) Device() *Device {
	return s.d
}

// Format returns format of stream.
func (s *InStream) Format() Format {
	p := s.cptr()
	return Format(p.format)
}

// SampleRate returns sample rate of stream.
func (s *InStream) SampleRate() int {
	p := s.cptr()
	return int(p.sample_rate)
}

// Layout returns layout of stream.
func (s *InStream) Layout() *ChannelLayout {
	p := s.cptr()
	return newChannelLayout(&p.layout)
}

// SoftwareLatency returns software latency of stream.
func (s *InStream) SoftwareLatency() float64 {
	p := s.cptr()
	return float64(p.software_latency)
}

// Name returns name of stream.
func (s *InStream) Name() string {
	p := s.cptr()
	return C.GoString(p.name)
}

// NonTerminalHint returns hint that this input stream is nonterminal.
// This is used by JACK and it means that the data received by the stream will be
// passed on or made available to another stream. Defaults to `false`.
func (s *InStream) NonTerminalHint() bool {
	p := s.cptr()
	return bool(p.non_terminal_hint)
}

// BytesPerFrame returns bytes per frame.
func (s *InStream) BytesPerFrame() int {
	p := s.cptr()
	return int(p.bytes_per_frame)
}

// BytesPerSample returns bytes per sample.
func (s *InStream) BytesPerSample() int {
	p := s.cptr()
	return int(p.bytes_per_sample)
}

// LayoutError returns error If setting the channel layout fails for some reason.
// Possible error:
// * SoundIoErrorIncompatibleDevice
func (s *InStream) LayoutError() error {
	p := s.cptr()
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
	p := s.cptr()
	if p != nil {
		C.free(unsafe.Pointer(p.name))
		C.soundio_instream_destroy(p)
		s.p = 0
	}
}

// Start starts recording.
// After you call this function, ReadCallback will be called.
func (s *InStream) Start() error {
	p := s.cptr()
	return convertToError(C.soundio_instream_start(p))
}

// BeginRead called when you are ready to begin reading from the device buffer.
func (s *InStream) BeginRead(frameCount *int) (*ChannelAreas, error) {
	p := s.cptr()
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
	p := s.cptr()
	return convertToError(C.soundio_instream_end_read(p))
}

// Pause pauses the stream and prevents ReadCallback from being called
// If the underlying device supports pausing.
func (s *InStream) Pause(pause bool) error {
	p := s.cptr()
	return convertToError(C.soundio_instream_pause(p, C.bool(pause)))
}

// Latency returns the number of seconds that the next frame of sound being
// captured will take to arrive in the buffer, plus the amount of time that is
// represented in the buffer.
// This includes both software and hardware latency.
func (s *InStream) Latency() (float64, error) {
	p := s.cptr()
	var latency C.double
	err := convertToError(C.soundio_instream_get_latency(p, &latency))
	return float64(latency), err
}

func newInStream(d *Device, config *InStreamConfig) (*InStream, error) {
	p := C.soundio_instream_create(d.cptr())
	s := &InStream{
		p: uintptr(unsafe.Pointer(p)),
		d: d,
	}

	p.userdata = unsafe.Pointer(s)
	C.setInStreamCallback(p)

	// set config values
	if config.Format != FormatInvalid {
		p.format = uint32(config.Format)
	}
	if config.SampleRate > 0 {
		p.sample_rate = C.int(config.SampleRate)
	}
	if config.Layout != nil {
		C.memcpy(unsafe.Pointer(&p.layout), unsafe.Pointer(config.Layout.cptr()), C.sizeof_struct_SoundIoChannelLayout)
	}
	if config.SoftwareLatency > 0.0 {
		p.software_latency = C.double(config.SoftwareLatency)
	}
	if config.Name != "" {
		p.name = C.CString(config.Name)
	}

	err := convertToError(C.soundio_instream_open(p))
	if err != nil {
		C.soundio_instream_destroy(p)
		return nil, err
	}

	return s, nil
}

func (s *InStream) cptr() *C.struct_SoundIoInStream {
	if s.p == 0 {
		return nil
	}
	return (*C.struct_SoundIoInStream)(unsafe.Pointer(s.p))
}
