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
import (
	"sync/atomic"
	"unsafe"
)

// Device is input/output device.
type Device struct {
	ptr uintptr
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

// ID returns device id.
func (d *Device) ID() string {
	p := d.pointer()
	if p == nil {
		return ""
	}
	return C.GoString(p.id)
}

// Name returns device name.
func (d *Device) Name() string {
	p := d.pointer()
	if p == nil {
		return ""
	}
	return C.GoString(p.name)
}

// Aim returns whether this device is an input device or an output device.
func (d *Device) Aim() DeviceAim {
	p := d.pointer()
	if p == nil {
		return deviceAimUnknown
	}
	return DeviceAim(uint32(p.aim))
}

// Layouts returns list of channel layout.
// Channel layouts are handled similarly to GetFormats.
// If this information is missing due to a GetProbeError,
// layouts will be nil. It's OK to modify this data, for example calling
// SortChannelLayouts on it.
// Devices are guaranteed to have at least 1 channel layout.
func (d *Device) Layouts() []*ChannelLayout {
	p := d.pointer()
	if p == nil {
		return make([]*ChannelLayout, 0)
	}
	count := int(p.layout_count)
	layouts := make([]*ChannelLayout, count)
	size := C.sizeof_struct_SoundIoChannelLayout
	base := uintptr(unsafe.Pointer(p.layouts))
	for i := 0; i < count; i++ {
		layouts[i] = &ChannelLayout{
			ptr: base + uintptr(i*size),
		}
	}
	return layouts
}

// LayoutCount returns how many formats are available in GetLayouts.
func (d *Device) LayoutCount() int {
	p := d.pointer()
	if p == nil {
		return 0
	}
	return int(p.layout_count)
}

// CurrentLayout returns current layout.
func (d *Device) CurrentLayout() *ChannelLayout {
	p := d.pointer()
	if p == nil {
		return nil
	}
	return newChannelLayout(&p.current_layout)
}

// Formats returns list of formats this device supports.
func (d *Device) Formats() []Format {
	p := d.pointer()
	if p == nil {
		return make([]Format, 0)
	}
	count := int(p.format_count)
	formats := make([]Format, count)
	size := C.sizeof_int
	base := uintptr(unsafe.Pointer(p.formats))
	for i := 0; i < count; i++ {
		f := (*uint32)(unsafe.Pointer(base + uintptr(i*size)))
		formats[i] = Format(*f)
	}
	return formats
}

// FormatCount returns how many formats are available in GetFormats.
func (d *Device) FormatCount() int {
	p := d.pointer()
	if p == nil {
		return 0
	}
	return int(p.format_count)
}

// CurrentFormat returns current format.
func (d *Device) CurrentFormat() Format {
	p := d.pointer()
	if p == nil {
		return FormatInvalid
	}
	return Format(uint32(p.current_format))
}

// SampleRates returns list of sample rate this device supports.
func (d *Device) SampleRates() []SampleRateRange {
	p := d.pointer()
	if p == nil {
		return make([]SampleRateRange, 0)
	}
	count := int(p.sample_rate_count)
	rates := make([]SampleRateRange, count)
	size := C.sizeof_struct_SoundIoSampleRateRange
	base := uintptr(unsafe.Pointer(p.sample_rates))
	for i := 0; i < count; i++ {
		rates[i] = SampleRateRange{
			ptr: base + uintptr(i*size),
		}
	}
	return rates
}

// SampleRateCount returns how many sample rate are available in GetSampleRates.
func (d *Device) SampleRateCount() int {
	p := d.pointer()
	if p == nil {
		return 0
	}
	return int(p.sample_rate_count)
}

// SampleRateCurrent returns current sample rate.
func (d *Device) SampleRateCurrent() int {
	p := d.pointer()
	if p == nil {
		return 0
	}
	return int(p.sample_rate_current)
}

// SoftwareLatencyMin returns software latency minimum in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) SoftwareLatencyMin() float64 {
	p := d.pointer()
	if p == nil {
		return 0.0
	}
	return float64(p.software_latency_min)
}

// SoftwareLatencyMax returns software latency maximum in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) SoftwareLatencyMax() float64 {
	p := d.pointer()
	if p == nil {
		return 0.0
	}
	return float64(p.software_latency_max)
}

// SoftwareLatencyCurrent returns software latency in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) SoftwareLatencyCurrent() float64 {
	p := d.pointer()
	if p == nil {
		return 0.0
	}
	return float64(p.software_latency_current)
}

// Raw means that you are directly opening the hardware device and not
// going through a proxy such as dmix, PulseAudio, or JACK. When you open a
// raw device, other applications on the computer are not able to
// simultaneously access the device. Raw devices do not perform automatic
// resampling and thus tend to have fewer formats available.
func (d *Device) Raw() bool {
	p := d.pointer()
	if p == nil {
		return false
	}
	return bool(p.is_raw)
}

// RefCount returns number of devices referenced.
func (d *Device) RefCount() int {
	p := d.pointer()
	if p == nil {
		return 0
	}
	return int(p.ref_count)
}

// ProbeError returns error representing the result of the device probe.
// Ideally this will be nil in which case all the fields of
// the device will be populated. If there is an error code here
// then information about formats, sample rates, and channel layouts might be missing.
//
// Possible errors:
// * ErrorOpeningDevice
// * ErrorNoMem
func (d *Device) ProbeError() error {
	p := d.pointer()
	if p == nil {
		return errorUninitialized
	}
	return convertToError(p.probe_error)
}

// functions

// AddReference is increments the device's reference count.
func (d *Device) AddReference() {
	p := d.pointer()
	if p == nil {
		return
	}
	C.soundio_device_ref(p)
}

// RemoveReference is decrements the device's reference count.
func (d *Device) RemoveReference() {
	p := d.pointer()
	if p == nil {
		return
	}
	C.soundio_device_unref(p)
}

// Equal returns true if and only if the devices have the same GetID,
// IsRaw, and GetAim are the same.
func (d *Device) Equal(o *Device) bool {
	p := d.pointer()
	op := o.pointer()
	if p == nil || op == nil {
		return false
	}
	return bool(C.soundio_device_equal(p, op))
}

// SortChannelLayouts sorts channel layouts by channel count, descending.
func (d *Device) SortChannelLayouts() {
	p := d.pointer()
	if p == nil {
		return
	}
	C.soundio_device_sort_channel_layouts(p)
}

// SupportsFormat returns whether `format` is included in the device's supported formats.
func (d *Device) SupportsFormat(format Format) bool {
	p := d.pointer()
	if p == nil {
		return false
	}
	return bool(C.soundio_device_supports_format(p, uint32(format)))
}

// SupportsLayout returns whether `layout` is included in the device's supported channel layouts.
func (d *Device) SupportsLayout(layout ChannelLayout) bool {
	p := d.pointer()
	if p == nil {
		return false
	}
	return bool(C.soundio_device_supports_layout(p, layout.pointer()))
}

// SupportsSampleRate returns whether `sampleRate` is included in the device's supported sample rates.
func (d *Device) SupportsSampleRate(sampleRate int) bool {
	p := d.pointer()
	if p == nil {
		return false
	}
	return bool(C.soundio_device_supports_sample_rate(p, C.int(sampleRate)))
}

// NearestSampleRate returns the available sample rate nearest to sampleRate, rounding up.
func (d *Device) NearestSampleRate(sampleRate int) int {
	p := d.pointer()
	if p == nil {
		return 0
	}
	return int(C.soundio_device_nearest_sample_rate(p, C.int(sampleRate)))
}

// NewInStream allocates memory and sets defaults.
// Next you should fill out the struct fields and then call Open function.
// Sets all fields to defaults.
func (d *Device) NewInStream() *InStream {
	p := d.pointer()
	if p == nil {
		return nil
	}

	ptr := C.soundio_instream_create(p)
	stream := &InStream{
		ptr:    uintptr(unsafe.Pointer(ptr)),
		device: d,
	}
	ptr.userdata = unsafe.Pointer(stream)
	C.setInstreamCallback(ptr)
	return stream
}

// NewOutStream allocates memory and sets defaults.
// Next you should fill out the struct fields and then call Open function.
// Sets all fields to defaults.
func (d *Device) NewOutStream() *OutStream {
	p := d.pointer()
	if p == nil {
		return nil
	}

	ptr := C.soundio_outstream_create(p)
	stream := &OutStream{
		ptr:    uintptr(unsafe.Pointer(ptr)),
		device: d,
	}
	ptr.userdata = unsafe.Pointer(stream)
	C.setOutstreamCallback(ptr)
	return stream
}

func (d *Device) pointer() *C.struct_SoundIoDevice {
	if d == nil {
		return nil
	}
	p := atomic.LoadUintptr(&d.ptr)
	if p == 0 {
		return nil
	}
	return (*C.struct_SoundIoDevice)(unsafe.Pointer(p))
}

func newDevice(p *C.struct_SoundIoDevice) *Device {
	return &Device{
		ptr: uintptr(unsafe.Pointer(p)),
	}
}
