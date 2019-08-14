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

// GetID returns device id.
func (d *Device) GetID() string {
	return C.GoString(d.ptr.id)
}

// GetName returns device name.
func (d *Device) GetName() string {
	return C.GoString(d.ptr.name)
}

// GetAim returns whether this device is an input device or an output device.
func (d *Device) GetAim() DeviceAim {
	return DeviceAim(uint32(d.ptr.aim))
}

// GetLayouts returns list of channel layout.
// Channel layouts are handled similarly to GetFormats.
// If this information is missing due to a GetProbeError,
// layouts will be nil. It's OK to modify this data, for example calling
// SortChannelLayouts on it.
// Devices are guaranteed to have at least 1 channel layout.
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

// GetLayoutCount returns how many formats are available in GetLayouts.
func (d *Device) GetLayoutCount() int {
	return int(d.ptr.layout_count)
}

// GetCurrentLayout returns current layout.
func (d *Device) GetCurrentLayout() *ChannelLayout {
	return &ChannelLayout{
		ptr: &(d.ptr.current_layout),
	}
}

// GetFormats returns list of formats this device supports.
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

// GetFormatCount returns how many formats are available in GetFormats.
func (d *Device) GetFormatCount() int {
	return int(d.ptr.format_count)
}

// GetCurrentFormat returns current format.
func (d *Device) GetCurrentFormat() Format {
	return Format(uint32(d.ptr.current_format))
}

// GetSampleRates returns list of sample rate this device supports.
func (d *Device) GetSampleRates() *SampleRateRange {
	return &SampleRateRange{
		ptr: d.ptr.sample_rates,
	}
}

// GetSampleRateCount returns how many sample rate are available in GetSampleRates.
func (d *Device) GetSampleRateCount() int {
	return int(d.ptr.sample_rate_count)
}

// GetSampleRateCurrent returns current sample rate.
func (d *Device) GetSampleRateCurrent() int {
	return int(d.ptr.sample_rate_current)
}

// GetSoftwareLatencyMin returns software latency minimum in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) GetSoftwareLatencyMin() float64 {
	return float64(d.ptr.software_latency_min)
}

// GetSoftwareLatencyMax returns software latency maximum in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) GetSoftwareLatencyMax() float64 {
	return float64(d.ptr.software_latency_max)
}

// GetSoftwareLatencyCurrent returns software latency in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) GetSoftwareLatencyCurrent() float64 {
	return float64(d.ptr.software_latency_current)
}

// IsRaw means that you are directly opening the hardware device and not
// going through a proxy such as dmix, PulseAudio, or JACK. When you open a
// raw device, other applications on the computer are not able to
// simultaneously access the device. Raw devices do not perform automatic
// resampling and thus tend to have fewer formats available.
func (d *Device) IsRaw() bool {
	return bool(d.ptr.is_raw)
}

// GetRefCount returns number of devices referenced.
func (d *Device) GetRefCount() int {
	return int(d.ptr.ref_count)
}

// GetProbeError returns error representing the result of the device probe.
// Ideally this will be nil in which case all the fields of
// the device will be populated. If there is an error code here
// then information about formats, sample rates, and channel layouts might be missing.
//
// Possible errors:
// * ErrorOpeningDevice
// * ErrorNoMem
func (d *Device) GetProbeError() error {
	return convertToError(d.ptr.probe_error)
}

// functions

// AddReference is increments the device's reference count.
func (d *Device) AddReference() {
	C.soundio_device_ref(d.ptr)
}

// RemoveReference is decrements the device's reference count.
func (d *Device) RemoveReference() {
	C.soundio_device_unref(d.ptr)
}

// Equal returns true if and only if the devices have the same GetID,
// IsRaw, and GetAim are the same.
func (d *Device) Equal(o *Device) bool {
	return bool(C.soundio_device_equal(d.ptr, o.ptr))
}

// SortChannelLayouts sorts channel layouts by channel count, descending.
func (d *Device) SortChannelLayouts() {
	C.soundio_device_sort_channel_layouts(d.ptr)
}

// SupportsFormat returns whether `format` is included in the device's supported formats.
func (d *Device) SupportsFormat(format Format) bool {
	return bool(C.soundio_device_supports_format(d.ptr, uint32(format)))
}

// SupportsLayout returns whether `layout` is included in the device's supported channel layouts.
func (d *Device) SupportsLayout(layout ChannelLayout) bool {
	return bool(C.soundio_device_supports_layout(d.ptr, layout.ptr))
}

// SupportsSampleRate returns whether `sampleRate` is included in the device's supported sample rates.
func (d *Device) SupportsSampleRate(sampleRate int) bool {
	return bool(C.soundio_device_supports_sample_rate(d.ptr, C.int(sampleRate)))
}

// NearestSampleRate returns the available sample rate nearest to sampleRate, rounding up.
func (d *Device) NearestSampleRate(sampleRate int) int {
	return int(C.soundio_device_nearest_sample_rate(d.ptr, C.int(sampleRate)))
}

// InStreamCreate allocates memory and sets defaults.
// Next you should fill out the struct fields and then call Open function.
// Sets all fields to defaults.
func (d *Device) InStreamCreate() *InStream {
	ptr := C.soundio_instream_create(d.ptr)
	stream := &InStream{
		ptr: ptr,
	}
	ptr.userdata = unsafe.Pointer(stream)
	C.setInstreamCallback(ptr)
	return stream
}

// OutStreamCreate allocates memory and sets defaults.
// Next you should fill out the struct fields and then call Open function.
// Sets all fields to defaults.
func (d *Device) OutStreamCreate() *OutStream {
	ptr := C.soundio_outstream_create(d.ptr)
	stream := &OutStream{
		ptr: ptr,
	}
	ptr.userdata = unsafe.Pointer(stream)
	C.setOutstreamCallback(ptr)
	return stream
}
