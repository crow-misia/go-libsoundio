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
	"log"
	"unsafe"
)

// OutStream is Output Stream.
type OutStream struct {
	p                 uintptr
	d                 Device
	writeCallback     func(*OutStream, int, int)
	underflowCallback func(*OutStream)
	errorCallback     func(*OutStream, error)
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
func (s *OutStream) Device() Device {
	return s.d
}

// Format returns format of stream.
func (s *OutStream) Format() Format {
	p := s.cptr()
	return Format(p.format)
}

// SetFormat sets format of stream.
func (s *OutStream) SetFormat(format Format) {
	p := s.cptr()
	p.format = uint32(format)
}

// SampleRate returns sample rate of stream.
func (s *OutStream) SampleRate() int {
	p := s.cptr()
	return int(p.sample_rate)
}

// SetSampleRate sets sample rate of stream.
func (s *OutStream) SetSampleRate(sampleRate int) {
	p := s.cptr()
	p.sample_rate = C.int(sampleRate)
}

// Layout returns layout of stream.
func (s *OutStream) Layout() *ChannelLayout {
	p := s.cptr()
	return newChannelLayout(&p.layout)
}

// SetLayout sets layout of stream.
func (s *OutStream) SetLayout(layout *ChannelLayout) {
	p := s.cptr()
	C.memcpy(unsafe.Pointer(&p.layout), unsafe.Pointer(layout.cptr()), C.sizeof_struct_SoundIoChannelLayout)
}

// SoftwareLatency returns software latency of stream.
func (s *OutStream) SoftwareLatency() float64 {
	p := s.cptr()
	return float64(p.software_latency)
}

// SetSoftwareLatency sets software latency of stream.
func (s *OutStream) SetSoftwareLatency(latency float64) {
	p := s.cptr()
	p.software_latency = C.double(latency)
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

// SetName sets name of stream.
func (s *OutStream) SetName(name string) {
	p := s.cptr()
	if p.name != nil {
		C.free(unsafe.Pointer(p.name))
	}
	p.name = C.CString(name)
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

// Open opens stream.
// After you call this function, SoftwareLatency is set to the correct value.
// The next thing to do is call Start function.
// If this function returns an error, the outstream is in an invalid state and
// you must call Destroy function on it.
//
// Possible errors:
// * ErrorInvalid
//   device aim is not DeviceAimOutput
//   format is not valid
//   requested layout channel count > MaxChannels
// * ErrorNoMem
// * ErrorOpeningDevice
// * ErrorBackendDisconnected
// * ErrorSystemResources
// * ErrorNoSuchClient - when JACK returns `JackNoSuchClient`
// * ErrorIncompatibleBackend - SoundIoOutStream::channel_count is
//   greater than the number of channels the backend can handle.
// * ErrorIncompatibleDevice - stream parameters requested are not
//   compatible with the chosen device.
func (s *OutStream) Open() error {
	p := s.cptr()
	return convertToError(C.soundio_outstream_open(p))
}

// Destroy releases resources.
func (s *OutStream) Destroy() {
	log.Println("destroy OutStream")
	p := s.cptr()
	if p != nil {
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

func newOutStream(p *C.struct_SoundIoOutStream, d Device) *OutStream {
	s := &OutStream{
		p: uintptr(unsafe.Pointer(p)),
		d: d,
	}
	p.userdata = unsafe.Pointer(s)
	C.setOutStreamCallback(p)
	return s
}

func (s *OutStream) cptr() *C.struct_SoundIoOutStream {
	if s.p == 0 {
		return nil
	}
	return (*C.struct_SoundIoOutStream)(unsafe.Pointer(s.p))
}
