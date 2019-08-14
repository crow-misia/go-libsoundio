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
*/
import "C"
import "unsafe"

type InStream struct {
	ptr              *C.struct_SoundIoInStream
	namePtr          *C.char
	readCallback     func(*InStream, int, int)
	overflowCallback func(*InStream)
	errorCallback    func(*InStream, error)
}

// fields

// GetDevice returns device to which the stream belongs.
func (s *InStream) GetDevice() *Device {
	return &Device{
		ptr: s.ptr.device,
	}
}

// GetFormat returns format of stream.
func (s *InStream) GetFormat() Format {
	return Format(s.ptr.format)
}

// SetFormat sets format of stream.
func (s *InStream) SetFormat(format Format) {
	s.ptr.format = uint32(format)
}

// GetSampleRate returns sample rate of stream.
func (s *InStream) GetSampleRate() int {
	return int(s.ptr.sample_rate)
}

// SetSampleRate sets sample rate of stream.
func (s *InStream) SetSampleRate(sampleRate int) {
	s.ptr.sample_rate = C.int(sampleRate)
}

// GetLayout returns layout of stream.
func (s *InStream) GetLayout() *ChannelLayout {
	return &ChannelLayout{
		ptr: &(s.ptr.layout),
	}
}

// SetLayout sets layout of stream.
func (s *InStream) SetLayout(layout *ChannelLayout) {
	s.ptr.layout = *(layout.ptr)
}

// GetSoftwareLatency returns software latency of stream.
func (s *InStream) GetSoftwareLatency() float64 {
	return float64(s.ptr.software_latency)
}

// SetSoftwareLatency sets software latency of stream.
func (s *InStream) SetSoftwareLatency(latency float64) {
	s.ptr.software_latency = C.double(latency)
}

// GetName returns name of stream.
func (s *InStream) GetName() string {
	return C.GoString(s.ptr.name)
}

// SetName sets name of stream.
func (s *InStream) SetName(name string) {
	if s.namePtr != nil {
		C.free(unsafe.Pointer(s.namePtr))
	}
	s.ptr.name = C.CString(name)
}

// GetNonTerminalHint returns hint that this input stream is nonterminal.
// This is used by JACK and it means that the data received by the stream will be
// passed on or made available to another stream. Defaults to `false`.
func (s *InStream) GetNonTerminalHint() bool {
	return bool(s.ptr.non_terminal_hint)
}

// GetBytesPerFrame returns bytes per frame.
func (s *InStream) GetBytesPerFrame() int {
	return int(s.ptr.bytes_per_frame)
}

// GetBytesPerSample returns bytes per sample.
func (s *InStream) GetBytesPerSample() int {
	return int(s.ptr.bytes_per_sample)
}

// GetLayoutError returns error If setting the channel layout fails for some reason.
// Possible error:
// * SoundIoErrorIncompatibleDevice
func (s *InStream) GetLayoutError() error {
	return convertToError(s.ptr.layout_error)
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
	if s.namePtr != nil {
		C.free(unsafe.Pointer(s.namePtr))
	}
	s.namePtr = nil

	if s.ptr != nil {
		s.ptr.userdata = nil
		C.soundio_instream_destroy(s.ptr)
	}
	s.ptr = nil
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
	return convertToError(C.soundio_instream_open(s.ptr))
}

// Start starts recording.
// After you call this function, ReadCallback will be called.
func (s *InStream) Start() error {
	return convertToError(C.soundio_instream_start(s.ptr))
}

// BeginRead called when you are ready to begin reading from the device buffer.
func (s *InStream) BeginRead(frameCount *int) (*ChannelAreas, error) {
	var ptrs *C.struct_SoundIoChannelArea

	nativeFrameCount := C.int(*frameCount)
	err := convertToError(C.soundio_instream_begin_read(s.ptr, &ptrs, &nativeFrameCount))
	*frameCount = int(nativeFrameCount)

	return &ChannelAreas{
		ptr:          ptrs,
		channelCount: s.GetLayout().GetChannelCount(),
		frameCount:   *frameCount,
	}, err
}

// EndRead will drop all of the frames from when you called.
func (s *InStream) EndRead() error {
	return convertToError(C.soundio_instream_end_read(s.ptr))
}

// Pause pauses the stream and prevents ReadCallback from being called
// If the underyling device supports pausing.
func (s *InStream) Pause(pause bool) error {
	return convertToError(C.soundio_instream_pause(s.ptr, C.bool(pause)))
}

// GetLatency returns the number of seconds that the next frame of sound being
// captured will take to arrive in the buffer, plus the amount of time that is
// represented in the buffer.
// This includes both software and hardware latency.
func (s *InStream) GetLatency() (float64, error) {
	var latency C.double
	err := convertToError(C.soundio_instream_get_latency(s.ptr, &latency))
	return float64(latency), err
}
