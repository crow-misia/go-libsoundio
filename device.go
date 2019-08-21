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
import (
	"unsafe"
)

// Device is input/output device.
type Device struct {
	p *C.struct_SoundIoDevice
}

// fields

// ID returns device id.
func (d *Device) ID() string {
	return C.GoString(d.p.id)
}

// Name returns device name.
func (d *Device) Name() string {
	return C.GoString(d.p.name)
}

// Aim returns whether this device is an input device or an output device.
func (d *Device) Aim() DeviceAim {
	return DeviceAim(uint32(d.p.aim))
}

// Layouts returns list of channel layout.
// Channel layouts are handled similarly to GetFormats.
// If this information is missing due to a GetProbeError,
// layouts will be nil. It's OK to modify this data, for example calling
// SortChannelLayouts on it.
// Devices are guaranteed to have at least 1 channel layout.
func (d *Device) Layouts() []*ChannelLayout {
	count := int(d.p.layout_count)
	layouts := make([]*ChannelLayout, count)
	size := C.sizeof_struct_SoundIoChannelLayout
	base := uintptr(unsafe.Pointer(d.p.layouts))
	for i := 0; i < count; i++ {
		layouts[i] = newChannelLayout((*C.struct_SoundIoChannelLayout)(unsafe.Pointer(base + uintptr(i*size))))
	}
	return layouts
}

// LayoutCount returns how many formats are available in GetLayouts.
func (d *Device) LayoutCount() int {
	return int(d.p.layout_count)
}

// CurrentLayout returns current layout.
func (d *Device) CurrentLayout() *ChannelLayout {
	return newChannelLayout(&d.p.current_layout)
}

// Formats returns list of formats this device supports.
func (d *Device) Formats() []Format {
	count := int(d.p.format_count)
	formats := make([]Format, count)
	size := C.sizeof_int
	base := uintptr(unsafe.Pointer(d.p.formats))
	for i := 0; i < count; i++ {
		f := (*uint32)(unsafe.Pointer(base + uintptr(i*size)))
		formats[i] = Format(*f)
	}
	return formats
}

// FormatCount returns how many formats are available in GetFormats.
func (d *Device) FormatCount() int {
	return int(d.p.format_count)
}

// CurrentFormat returns current format.
func (d *Device) CurrentFormat() Format {
	return Format(uint32(d.p.current_format))
}

// SampleRates returns list of sample rate this device supports.
func (d *Device) SampleRates() []SampleRateRange {
	count := int(d.p.sample_rate_count)
	rates := make([]SampleRateRange, count)
	size := C.sizeof_struct_SoundIoSampleRateRange
	base := uintptr(unsafe.Pointer(d.p.sample_rates))
	for i := 0; i < count; i++ {
		rates[i] = newSampleRateRange(base + uintptr(i*size))
	}
	return rates
}

// SampleRateCount returns how many sample rate are available in GetSampleRates.
func (d *Device) SampleRateCount() int {
	return int(d.p.sample_rate_count)
}

// SampleRateCurrent returns current sample rate.
func (d *Device) SampleRateCurrent() int {
	return int(d.p.sample_rate_current)
}

// SoftwareLatencyMin returns software latency minimum in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) SoftwareLatencyMin() float64 {
	return float64(d.p.software_latency_min)
}

// SoftwareLatencyMax returns software latency maximum in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) SoftwareLatencyMax() float64 {
	return float64(d.p.software_latency_max)
}

// SoftwareLatencyCurrent returns software latency in seconds.
// If this value is unknown or irrelevant, it is set to 0.0.
// For PulseAudio and WASAPI this value is unknown until you open a stream.
func (d *Device) SoftwareLatencyCurrent() float64 {
	return float64(d.p.software_latency_current)
}

// Raw means that you are directly opening the hardware device and not
// going through a proxy such as dmix, PulseAudio, or JACK. When you open a
// raw device, other applications on the computer are not able to
// simultaneously access the device. Raw devices do not perform automatic
// resampling and thus tend to have fewer formats available.
func (d *Device) Raw() bool {
	return bool(d.p.is_raw)
}

// RefCount returns number of devices referenced.
func (d *Device) RefCount() int {
	return int(d.p.ref_count)
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
	return convertToError(d.p.probe_error)
}

// functions

// AddReference is increments the device's reference count.
func (d *Device) AddReference() {
	C.soundio_device_ref(d.p)
}

// RemoveReference is decrements the device's reference count.
func (d *Device) RemoveReference() {
	C.soundio_device_unref(d.p)
}

// Equal returns true if and only if the devices have the same GetID,
// IsRaw, and GetAim are the same.
func (d *Device) Equal(o *Device) bool {
	return bool(C.soundio_device_equal(d.p, o.p))
}

// SortChannelLayouts sorts channel layouts by channel count, descending.
func (d *Device) SortChannelLayouts() {
	C.soundio_device_sort_channel_layouts(d.p)
}

// SupportsFormat returns whether `format` is included in the device's supported formats.
func (d *Device) SupportsFormat(format Format) bool {
	return bool(C.soundio_device_supports_format(d.p, uint32(format)))
}

// SupportsLayout returns whether `layout` is included in the device's supported channel layouts.
func (d *Device) SupportsLayout(layout ChannelLayout) bool {
	return bool(C.soundio_device_supports_layout(d.p, layout.p))
}

// SupportsSampleRate returns whether `sampleRate` is included in the device's supported sample rates.
func (d *Device) SupportsSampleRate(sampleRate int) bool {
	return bool(C.soundio_device_supports_sample_rate(d.p, C.int(sampleRate)))
}

// NearestSampleRate returns the available sample rate nearest to sampleRate, rounding up.
func (d *Device) NearestSampleRate(sampleRate int) int {
	return int(C.soundio_device_nearest_sample_rate(d.p, C.int(sampleRate)))
}

// NewInStream allocates memory and sets defaults.
// Next you should fill out the struct fields and then call Open function.
// Sets all fields to defaults.
func (d *Device) NewInStream() *InStream {
	return newInStream(C.soundio_instream_create(d.p), d)
}

// NewOutStream allocates memory and sets defaults.
// Next you should fill out the struct fields and then call Open function.
// Sets all fields to defaults.
func (d *Device) NewOutStream() *OutStream {
	return newOutStream(C.soundio_outstream_create(d.p), d)
}

func newDevice(p *C.struct_SoundIoDevice) *Device {
	return &Device{
		p: p,
	}
}
