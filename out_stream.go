package soundio

/*
#include <soundio/soundio.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

type OutStream struct {
	ptr               *C.struct_SoundIoOutStream
	namePtr           *C.char
	writeCallback     func(stream *OutStream, frameCountMin int, frameCountMax int)
	underflowCallback func(stream *OutStream)
	errorCallback     func(stream *OutStream, err error)
}

// fields

func (s *OutStream) GetDevice() *Device {
	return &Device{
		ptr: s.ptr.device,
	}
}

func (s *OutStream) GetFormat() Format {
	return Format(s.ptr.format)
}

func (s *OutStream) SetFormat(format Format) {
	s.ptr.format = uint32(format)
}

func (s *OutStream) GetSampleRate() int {
	return int(s.ptr.sample_rate)
}

func (s *OutStream) SetSampleRate(sampleRate int) {
	s.ptr.sample_rate = C.int(sampleRate)
}

func (s *OutStream) GetLayout() *ChannelLayout {
	return &ChannelLayout{
		ptr: &(s.ptr.layout),
	}
}

func (s *OutStream) SetLayout(layout *ChannelLayout) {
	s.ptr.layout = *(layout.ptr)
}

func (s *OutStream) GetSoftwareLatency() float64 {
	return float64(s.ptr.software_latency)
}

func (s *OutStream) SetSoftwareLatency(latency float64) {
	s.ptr.software_latency = C.double(latency)
}

func (s *OutStream) GetVolume() float32 {
	return float32(s.ptr.volume)
}

func (s *OutStream) SetVolume(volume float64) error {
	return convertToError(C.soundio_outstream_set_volume(s.ptr, C.double(volume)))
}

func (s *OutStream) GetName() string {
	return C.GoString(s.ptr.name)
}

func (s *OutStream) SetName(name string) {
	if s.namePtr != nil {
		C.free(unsafe.Pointer(s.namePtr))
	}
	s.ptr.name = C.CString(name)
}

func (s *OutStream) GetNonTerminalHint() bool {
	return bool(s.ptr.non_terminal_hint)
}

func (s *OutStream) GetBytesPerFrame() int {
	return int(s.ptr.bytes_per_frame)
}

func (s *OutStream) GetBytesPerSample() int {
	return int(s.ptr.bytes_per_sample)
}

func (s *OutStream) GetLayoutError() error {
	return convertToError(s.ptr.layout_error)
}

func (s *OutStream) SetWriteCallback(callback func(outStream *OutStream, frameCountMix int, frameCountMax int)) {
	s.writeCallback = callback
}

// functions

func (s *OutStream) Destroy() {
	if s.namePtr != nil {
		C.free(unsafe.Pointer(s.namePtr))
	}
	s.namePtr = nil

	if s.ptr != nil {
		s.ptr.userdata = nil
		C.soundio_outstream_destroy(s.ptr)
	}
	s.ptr = nil
}

func (s *OutStream) Open() error {
	return convertToError(C.soundio_outstream_open(s.ptr))
}

func (s *OutStream) Start() error {
	return convertToError(C.soundio_outstream_start(s.ptr))
}

func (s *OutStream) BeginWrite(frameCount *int) (*ChannelAreas, error) {
	var ptrs *C.struct_SoundIoChannelArea
	var nativeFrameCount C.int
	nativeFrameCount = C.int(*frameCount)
	err := convertToError(C.soundio_outstream_begin_write(s.ptr, &ptrs, &nativeFrameCount))
	*frameCount = int(nativeFrameCount)

	return &ChannelAreas{
		ptr:          ptrs,
		channelCount: s.GetLayout().GetChannelCount(),
		frameCount:   *frameCount,
	}, err
}

func (s *OutStream) EndWrite() error {
	return convertToError(C.soundio_outstream_end_write(s.ptr))
}

func (s *OutStream) ClearBuffer() error {
	return convertToError(C.soundio_outstream_clear_buffer(s.ptr))
}

func (s *OutStream) Pause(pause bool) error {
	return convertToError(C.soundio_outstream_pause(s.ptr, C.bool(pause)))
}

func (s *OutStream) GetLatency(outLatency float64) (float64, error) {
	var latency C.double
	latency = C.double(outLatency)
	err := convertToError(C.soundio_outstream_get_latency(s.ptr, &latency))
	return float64(latency), err
}
