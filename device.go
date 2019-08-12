package soundio

/*
#include <soundio/soundio.h>
#include <stdlib.h>

extern void instreamReadCallbackDelegate(struct SoundIoInStream *, int, int);
extern void instreamOverflowCallbackDelegate(struct SoundIoInStream *);
extern void instreamErrorCallbackDelegate(struct SoundIoInStream *, int);

extern void outstreamWriteCallbackDelegate(struct SoundIoOutStream *, int, int);
extern void outstreamUnderflowCallbackDelegate(struct SoundIoOutStream *);
extern void outstreamErrorCallbackDelegate(struct SoundIoOutStream *, int);

static void setInstreamCallback(struct SoundIoInStream *instream) {
	instream->read_callback = instreamReadCallbackDelegate;
	instream->overflow_callback = instreamOverflowCallbackDelegate;
	instream->error_callback = instreamErrorCallbackDelegate;
}

static void setOutstreamCallback(struct SoundIoOutStream *outstream) {
	outstream->write_callback = outstreamWriteCallbackDelegate;
	outstream->underflow_callback = outstreamUnderflowCallbackDelegate;
	outstream->error_callback = outstreamErrorCallbackDelegate;
}
*/
import "C"
import "unsafe"

type Device struct {
	ptr *C.struct_SoundIoDevice
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

func (d *Device) GetID() string {
	return C.GoString(d.ptr.id)
}

func (d *Device) GetName() string {
	return C.GoString(d.ptr.name)
}

func (d *Device) GetAim() DeviceAim {
	return DeviceAim(uint32(d.ptr.aim))
}

func (d *Device) GetLayouts() []*ChannelLayout {
	count := uintptr(d.ptr.layout_count)
	layouts := make([]*ChannelLayout, count)
	size := unsafe.Sizeof(*d.ptr.layouts)
	var i uintptr
	for i = 0; i < count; i++ {
		l := (*C.struct_SoundIoChannelLayout)(unsafe.Pointer(uintptr(unsafe.Pointer(d.ptr.layouts)) + i*size))
		layouts[i] = &ChannelLayout{
			ptr: l,
		}
	}
	return layouts
}

func (d *Device) GetLayoutCount() int {
	return int(d.ptr.layout_count)
}

func (d *Device) GetCurrentLayout() *ChannelLayout {
	return &ChannelLayout{
		ptr: &(d.ptr.current_layout),
	}
}

func (d *Device) GetFormats() []Format {
	count := uintptr(d.ptr.format_count)
	formats := make([]Format, count)
	size := unsafe.Sizeof(*d.ptr.formats)
	var i uintptr
	for i = 0; i < count; i++ {
		f := (*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(d.ptr.formats)) + i*size))
		formats[i] = Format(*f)
	}
	return formats
}

func (d *Device) GetFormatCount() int {
	return int(d.ptr.format_count)
}

func (d *Device) GetCurrentFormat() Format {
	return Format(uint32(d.ptr.current_format))
}

func (d *Device) GetSampleRates() *SampleRateRange {
	return &SampleRateRange{
		ptr: d.ptr.sample_rates,
	}
}

func (d *Device) GetSampleRateCount() int {
	return int(d.ptr.sample_rate_count)
}

func (d *Device) GetSampleRateCurrent() int {
	return int(d.ptr.sample_rate_current)
}

func (d *Device) GetSoftwareLatencyMin() float64 {
	return float64(d.ptr.software_latency_min)
}

func (d *Device) GetSoftwareLatencyMax() float64 {
	return float64(d.ptr.software_latency_max)
}

func (d *Device) GetSoftwareLatencyCurrent() float64 {
	return float64(d.ptr.software_latency_current)
}

func (d *Device) IsRaw() bool {
	return bool(d.ptr.is_raw)
}

func (d *Device) GetRefCount() int {
	return int(d.ptr.ref_count)
}

func (d *Device) GetProbeError() error {
	return convertToError(d.ptr.probe_error)
}

// functions

func (d *Device) AddReference() {
	C.soundio_device_ref(d.ptr)
}

func (d *Device) RemoveReference() {
	C.soundio_device_unref(d.ptr)
}

func (d *Device) Equal(o *Device) bool {
	return bool(C.soundio_device_equal(d.ptr, o.ptr))
}

func (d *Device) SortChannelLayouts() {
	C.soundio_device_sort_channel_layouts(d.ptr)
}

func (d *Device) SupportsFormat(format Format) bool {
	return bool(C.soundio_device_supports_format(d.ptr, uint32(format)))
}

func (d *Device) SupportsLayout(layout ChannelLayout) bool {
	return bool(C.soundio_device_supports_layout(d.ptr, layout.ptr))
}

func (d *Device) SupportsSampleRate(sampleRate int) bool {
	return bool(C.soundio_device_supports_sample_rate(d.ptr, C.int(sampleRate)))
}

func (d *Device) NearestSampleRate(sampleRate int) int {
	return int(C.soundio_device_nearest_sample_rate(d.ptr, C.int(sampleRate)))
}

func (d *Device) InStreamCreate() *InStream {
	ptr := C.soundio_instream_create(d.ptr)
	stream := &InStream{
		ptr: ptr,
	}
	ptr.userdata = unsafe.Pointer(stream)
	C.setInstreamCallback(ptr)
	return stream
}

func (d *Device) OutStreamCreate() *OutStream {
	ptr := C.soundio_outstream_create(d.ptr)
	stream := &OutStream{
		ptr: ptr,
	}
	ptr.userdata = unsafe.Pointer(stream)
	C.setOutstreamCallback(ptr)
	return stream
}
