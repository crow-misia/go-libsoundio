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
	readCallback     func(stream *InStream, frameCountMin int, frameCountMax int)
	overflowCallback func(stream *InStream)
	errorCallback    func(stream *InStream, err error)
}

// fields

func (s *InStream) GetDevice() *Device {
	return &Device{
		ptr: s.ptr.device,
	}
}

func (s *InStream) GetFormat() Format {
	return Format(s.ptr.format)
}

func (s *InStream) SetFormat(format Format) {
	s.ptr.format = uint32(format)
}

func (s *InStream) GetSampleRate() int {
	return int(s.ptr.sample_rate)
}

func (s *InStream) SetSampleRate(sampleRate int) {
	s.ptr.sample_rate = C.int(sampleRate)
}

func (s *InStream) GetLayout() *ChannelLayout {
	return &ChannelLayout{
		ptr: &(s.ptr.layout),
	}
}

func (s *InStream) SetLayout(layout *ChannelLayout) {
	s.ptr.layout = *(layout.ptr)
}

func (s *InStream) GetSoftwareLatency() float64 {
	return float64(s.ptr.software_latency)
}

func (s *InStream) SetSoftwareLatency(latency float64) {
	s.ptr.software_latency = C.double(latency)
}

func (s *InStream) GetName() string {
	return C.GoString(s.ptr.name)
}

func (s *InStream) SetName(name string) {
	if s.namePtr != nil {
		C.free(unsafe.Pointer(s.namePtr))
	}
	s.ptr.name = C.CString(name)
}

func (s *InStream) GetNonTerminalHint() bool {
	return bool(s.ptr.non_terminal_hint)
}

func (s *InStream) GetBytesPerFrame() int {
	return int(s.ptr.bytes_per_frame)
}

func (s *InStream) GetBytesPerSample() int {
	return int(s.ptr.bytes_per_sample)
}

func (s *InStream) GetLayoutError() error {
	return convertToError(s.ptr.layout_error)
}

// functions

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

func (s *InStream) Open() error {
	return convertToError(C.soundio_instream_open(s.ptr))
}

func (s *InStream) Start() error {
	return convertToError(C.soundio_instream_start(s.ptr))
}

func (s *InStream) BeginRead(frameCount *int) (*ChannelAreas, error) {
	var ptrs *C.struct_SoundIoChannelArea
	var nativeFrameCount C.int
	nativeFrameCount = C.int(*frameCount)
	err := convertToError(C.soundio_instream_begin_read(s.ptr, &ptrs, &nativeFrameCount))
	*frameCount = int(nativeFrameCount)

	return &ChannelAreas{
		ptr:          ptrs,
		channelCount: s.GetLayout().GetChannelCount(),
		frameCount:   *frameCount,
	}, err
}

func (s *InStream) EndRead() error {
	return convertToError(C.soundio_instream_end_read(s.ptr))
}

func (s *InStream) Pause(pause bool) error {
	return convertToError(C.soundio_instream_pause(s.ptr, C.bool(pause)))
}

func (s *InStream) GetLatency() (float64, error) {
	var latency C.double
	err := convertToError(C.soundio_instream_get_latency(s.ptr, &latency))
	return float64(latency), err
}
