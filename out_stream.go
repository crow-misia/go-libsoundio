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

// OutStream is Output Stream.
type OutStream struct {
	p                 uintptr
	d                 *Device
	writeCallback     func(*OutStream, int, int)
	underflowCallback func(*OutStream)
	errorCallback     func(*OutStream, error)
}

// OutStreamConfig is config of output stream.
type OutStreamConfig struct {
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

//export outstreamWriteCallbackDelegate
func outstreamWriteCallbackDelegate(nativeStream *C.struct_SoundIoOutStream, frameCountMin C.int, frameCountMax C.int) {
	stream := (*OutStream)(nativeStream.userdata)
	if stream.writeCallback != nil {
		stream.writeCallback(stream, int(frameCountMin), int(frameCountMax))
	}
}

//export outstreamUnderflowCallbackDelegate
func outstreamUnderflowCallbackDelegate(nativeStream *C.struct_SoundIoOutStream) {
	stream := (*OutStream)(nativeStream.userdata)
	if stream.underflowCallback != nil {
		stream.underflowCallback(stream)
	}
}

//export outstreamErrorCallbackDelegate
func outstreamErrorCallbackDelegate(nativeStream *C.struct_SoundIoOutStream, err C.int) {
	stream := (*OutStream)(nativeStream.userdata)
	if stream.errorCallback != nil {
		stream.errorCallback(stream, convertToError(err))
	}
}

// fields

// Device returns device to which the stream belongs.
func (s *OutStream) Device() *Device {
	return s.d
}

// Format returns format of stream.
func (s *OutStream) Format() Format {
	p := s.cptr()
	return Format(p.format)
}

// SampleRate returns sample rate of stream.
func (s *OutStream) SampleRate() int {
	p := s.cptr()
	return int(p.sample_rate)
}

// Layout returns layout of stream.
func (s *OutStream) Layout() *ChannelLayout {
	p := s.cptr()
	return newChannelLayout(&p.layout)
}

// SoftwareLatency returns software latency of stream.
func (s *OutStream) SoftwareLatency() float64 {
	p := s.cptr()
	return float64(p.software_latency)
}

// Volume returns volume of stream.
func (s *OutStream) Volume() float32 {
	p := s.cptr()
	return float32(p.volume)
}

// SetVolume sets volume of stream.
func (s *OutStream) SetVolume(volume float64) error {
	p := s.cptr()
	return convertToError(C.soundio_outstream_set_volume(p, C.double(volume)))
}

// Name returns name of stream.
func (s *OutStream) Name() string {
	p := s.cptr()
	return C.GoString(p.name)
}

// NonTerminalHint returns hint that this output stream is nonterminal.
// This is used by JACK and it means that the output stream data originates from an input
// stream. Defaults to `false`.
func (s *OutStream) NonTerminalHint() bool {
	p := s.cptr()
	return bool(p.non_terminal_hint)
}

// BytesPerFrame returns bytes per frame.
func (s *OutStream) BytesPerFrame() int {
	p := s.cptr()
	return int(p.bytes_per_frame)
}

// BytesPerSample returns bytes per sample.
func (s *OutStream) BytesPerSample() int {
	p := s.cptr()
	return int(p.bytes_per_sample)
}

// LayoutError returns error If setting the channel layout fails for some reason.
// Possible error:
// * SoundIoErrorIncompatibleDevice
func (s *OutStream) LayoutError() error {
	p := s.cptr()
	return convertToError(p.layout_error)
}

// SetWriteCallback sets WriteCallback.
func (s *OutStream) SetWriteCallback(callback func(stream *OutStream, frameCountMin int, frameCountMax int)) {
	s.writeCallback = callback
}

// SetUnderflowCallback sets UnderflowCallback.
func (s *OutStream) SetUnderflowCallback(callback func(stream *OutStream)) {
	s.underflowCallback = callback
}

// SetErrorCallback sets ErrorCallback.
func (s *OutStream) SetErrorCallback(callback func(stream *OutStream, err error)) {
	s.errorCallback = callback
}

// functions

// Destroy releases resources.
func (s *OutStream) Destroy() {
	p := s.cptr()
	if p != nil {
		C.free(unsafe.Pointer(p.name))
		C.soundio_outstream_destroy(p)
		s.p = 0
	}
}

// Start starts playback.
// After you call this function, WriteCallback will be called.
func (s *OutStream) Start() error {
	p := s.cptr()
	return convertToError(C.soundio_outstream_start(p))
}

// BeginWrite called when you are ready to begin writing to the device buffer.
func (s *OutStream) BeginWrite(frameCount *int) (*ChannelAreas, error) {
	p := s.cptr()
	var ptrs *C.struct_SoundIoChannelArea
	nativeFrameCount := C.int(*frameCount)
	err := convertToError(C.soundio_outstream_begin_write(p, &ptrs, &nativeFrameCount))
	*frameCount = int(nativeFrameCount)
	if err != nil {
		return nil, err
	}
	if ptrs == nil {
		return nil, nil
	}
	return newChannelAreas(ptrs, s.Layout().ChannelCount(), *frameCount), nil
}

// EndWrite commits the write that you began with BeginWrite.
func (s *OutStream) EndWrite() error {
	p := s.cptr()
	return convertToError(C.soundio_outstream_end_write(p))
}

// ClearBuffer clears the output stream buffer.
func (s *OutStream) ClearBuffer() error {
	p := s.cptr()
	return convertToError(C.soundio_outstream_clear_buffer(p))
}

// Pause pauses the stream If the underlying backend and device support pausing.
func (s *OutStream) Pause(pause bool) error {
	p := s.cptr()
	return convertToError(C.soundio_outstream_pause(p, C.bool(pause)))
}

// Latency returns the total number of seconds that the next frame written after the
// last frame written with EndWrite will take to become audible.
func (s *OutStream) Latency(outLatency float64) (float64, error) {
	p := s.cptr()
	latency := C.double(outLatency)
	err := convertToError(C.soundio_outstream_get_latency(p, &latency))
	return float64(latency), err
}

func newOutStream(d *Device, config *OutStreamConfig) (*OutStream, error) {
	p := C.soundio_outstream_create(d.cptr())
	s := &OutStream{
		p: uintptr(unsafe.Pointer(p)),
		d: d,
	}

	p.userdata = unsafe.Pointer(s)
	C.setOutStreamCallback(p)

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
	if config.Name == "" {
		p.name = C.CString("SoundIoOutStream")
	} else {
		p.name = C.CString(config.Name)
	}

	err := convertToError(C.soundio_outstream_open(p))
	if err != nil {
		C.soundio_outstream_destroy(p)
		return nil, err
	}

	return s, nil
}

func (s *OutStream) cptr() *C.struct_SoundIoOutStream {
	if s.p == 0 {
		return nil
	}
	return (*C.struct_SoundIoOutStream)(unsafe.Pointer(s.p))
}
