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
import "unsafe"

type OutStream struct {
	ptr               uintptr
	device            *Device
	writeCallback     func(*OutStream, int, int)
	underflowCallback func(*OutStream)
	errorCallback     func(*OutStream, error)
}

// fields

// GetDevice returns device to which the stream belongs.
func (s *OutStream) GetDevice() *Device {
	return s.device
}

// GetFormat returns format of stream.
func (s *OutStream) GetFormat() Format {
	p := s.getPointer()
	return Format(p.format)
}

// SetFormat sets format of stream.
func (s *OutStream) SetFormat(format Format) {
	p := s.getPointer()
	p.format = uint32(format)
}

// GetSampleRate returns sample rate of stream.
func (s *OutStream) GetSampleRate() int {
	p := s.getPointer()
	return int(p.sample_rate)
}

// SetSampleRate sets sample rate of stream.
func (s *OutStream) SetSampleRate(sampleRate int) {
	p := s.getPointer()
	p.sample_rate = C.int(sampleRate)
}

// GetLayout returns layout of stream.
func (s *OutStream) GetLayout() *ChannelLayout {
	p := s.getPointer()
	return createChannelLayout(&p.layout)
}

// SetLayout sets layout of stream.
func (s *OutStream) SetLayout(layout *ChannelLayout) {
	p := s.getPointer()
	C.memcpy(unsafe.Pointer(&p.layout), unsafe.Pointer(layout.ptr), C.sizeof_struct_SoundIoChannelLayout)
}

// GetSoftwareLatency returns software latency of stream.
func (s *OutStream) GetSoftwareLatency() float64 {
	p := s.getPointer()
	return float64(p.software_latency)
}

// SetSoftwareLatency sets software latency of stream.
func (s *OutStream) SetSoftwareLatency(latency float64) {
	p := s.getPointer()
	p.software_latency = C.double(latency)
}

// GetVolume returns volume of stream.
func (s *OutStream) GetVolume() float32 {
	p := s.getPointer()
	return float32(p.volume)
}

// SetVolume sets volume of stream.
func (s *OutStream) SetVolume(volume float64) error {
	return convertToError(C.soundio_outstream_set_volume(s.getPointer(), C.double(volume)))
}

// GetName returns name of stream.
func (s *OutStream) GetName() string {
	p := s.getPointer()
	return C.GoString(p.name)
}

// SetName sets name of stream.
func (s *OutStream) SetName(name string) {
	p := s.getPointer()
	if p.name != nil {
		C.free(unsafe.Pointer(p.name))
	}
	p.name = C.CString(name)
}

// GetNonTerminalHint returns hint that this output stream is nonterminal.
// This is used by JACK and it means that the output stream data originates from an input
// stream. Defaults to `false`.
func (s *OutStream) GetNonTerminalHint() bool {
	p := s.getPointer()
	return bool(p.non_terminal_hint)
}

// GetBytesPerFrame returns bytes per frame.
func (s *OutStream) GetBytesPerFrame() int {
	p := s.getPointer()
	return int(p.bytes_per_frame)
}

// GetBytesPerSample returns bytes per sample.
func (s *OutStream) GetBytesPerSample() int {
	p := s.getPointer()
	return int(p.bytes_per_sample)
}

// GetLayoutError returns error If setting the channel layout fails for some reason.
// Possible error:
// * SoundIoErrorIncompatibleDevice
func (s *OutStream) GetLayoutError() error {
	p := s.getPointer()
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
	if s.ptr != 0 {
		C.soundio_outstream_destroy(s.getPointer())
		s.ptr = 0
	}
}

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
	return convertToError(C.soundio_outstream_open(s.getPointer()))
}

// Start starts playback.
// After you call this function, WriteCallback will be called.
func (s *OutStream) Start() error {
	return convertToError(C.soundio_outstream_start(s.getPointer()))
}

// BeginWrite called when you are ready to begin writing to the device buffer.
func (s *OutStream) BeginWrite(frameCount *int) (*ChannelAreas, error) {
	var ptrs *C.struct_SoundIoChannelArea

	nativeFrameCount := C.int(*frameCount)
	err := convertToError(C.soundio_outstream_begin_write(s.getPointer(), &ptrs, &nativeFrameCount))
	*frameCount = int(nativeFrameCount)
	if err != nil {
		return nil, err
	}
	return newChannelAreas(ptrs, s.GetLayout().GetChannelCount(), *frameCount), nil
}

// EndWrite commits the write that you began with BeginWrite.
func (s *OutStream) EndWrite() error {
	return convertToError(C.soundio_outstream_end_write(s.getPointer()))
}

// ClearBuffer clears the output stream buffer.
func (s *OutStream) ClearBuffer() error {
	return convertToError(C.soundio_outstream_clear_buffer(s.getPointer()))
}

// Pause pauses the stream If the underlying backend and device support pausing.
func (s *OutStream) Pause(pause bool) error {
	return convertToError(C.soundio_outstream_pause(s.getPointer(), C.bool(pause)))
}

// GetLatency returns the total number of seconds that the next frame written after the
// last frame written with EndWrite will take to become audible.
func (s *OutStream) GetLatency(outLatency float64) (float64, error) {
	latency := C.double(outLatency)
	err := convertToError(C.soundio_outstream_get_latency(s.getPointer(), &latency))
	return float64(latency), err
}

func (s *OutStream) getPointer() *C.struct_SoundIoOutStream {
	return (*C.struct_SoundIoOutStream)(unsafe.Pointer(s.ptr))
}
