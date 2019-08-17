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

// ChannelLayoutID is channel layout id
type ChannelLayoutID uint32

type ChannelLayout struct {
	ptr uintptr
}

const (
	ChannelLayoutIDMono            = ChannelLayoutID(C.SoundIoChannelLayoutIdMono)
	ChannelLayoutIDStereo          = ChannelLayoutID(C.SoundIoChannelLayoutIdStereo)
	ChannelLayoutID2Point1         = ChannelLayoutID(C.SoundIoChannelLayoutId2Point1)
	ChannelLayoutID3Point0         = ChannelLayoutID(C.SoundIoChannelLayoutId3Point0)
	ChannelLayoutID3Point0Back     = ChannelLayoutID(C.SoundIoChannelLayoutId3Point0Back)
	ChannelLayoutID3Point1         = ChannelLayoutID(C.SoundIoChannelLayoutId3Point1)
	ChannelLayoutID4Point0         = ChannelLayoutID(C.SoundIoChannelLayoutId4Point0)
	ChannelLayoutIDQuad            = ChannelLayoutID(C.SoundIoChannelLayoutIdQuad)
	ChannelLayoutIDQuadSide        = ChannelLayoutID(C.SoundIoChannelLayoutIdQuadSide)
	ChannelLayoutID4Point1         = ChannelLayoutID(C.SoundIoChannelLayoutId4Point1)
	ChannelLayoutID5Point0Back     = ChannelLayoutID(C.SoundIoChannelLayoutId5Point0Back)
	ChannelLayoutID5Point0Side     = ChannelLayoutID(C.SoundIoChannelLayoutId5Point0Side)
	ChannelLayoutID5Point1         = ChannelLayoutID(C.SoundIoChannelLayoutId5Point1)
	ChannelLayoutID5Point1Back     = ChannelLayoutID(C.SoundIoChannelLayoutId5Point1Back)
	ChannelLayoutID6Point0Side     = ChannelLayoutID(C.SoundIoChannelLayoutId6Point0Side)
	ChannelLayoutID6Point0Front    = ChannelLayoutID(C.SoundIoChannelLayoutId6Point0Front)
	ChannelLayoutIDHexagonal       = ChannelLayoutID(C.SoundIoChannelLayoutIdHexagonal)
	ChannelLayoutID6Point1         = ChannelLayoutID(C.SoundIoChannelLayoutId6Point1)
	ChannelLayoutID6Point1Back     = ChannelLayoutID(C.SoundIoChannelLayoutId6Point1Back)
	ChannelLayoutID6Point1Front    = ChannelLayoutID(C.SoundIoChannelLayoutId6Point1Front)
	ChannelLayoutID7Point0         = ChannelLayoutID(C.SoundIoChannelLayoutId7Point0)
	ChannelLayoutID7Point0Front    = ChannelLayoutID(C.SoundIoChannelLayoutId7Point0Front)
	ChannelLayoutID7Point1         = ChannelLayoutID(C.SoundIoChannelLayoutId7Point1)
	ChannelLayoutID7Point1Wide     = ChannelLayoutID(C.SoundIoChannelLayoutId7Point1Wide)
	ChannelLayoutID7Point1WideBack = ChannelLayoutID(C.SoundIoChannelLayoutId7Point1WideBack)
	ChannelLayoutIDOctagonal       = ChannelLayoutID(C.SoundIoChannelLayoutIdOctagonal)
)

// ChannelLayoutBuiltinCount returns the number of builtin channel layouts.
func ChannelLayoutBuiltinCount() int {
	return int(C.soundio_channel_layout_builtin_count())
}

// ChannelLayoutGetBuiltin returns a builtin channel layout.
// 0 <= `index` < ChannelLayoutBuiltinCount
func ChannelLayoutGetBuiltin(index ChannelLayoutID) *ChannelLayout {
	return newChannelLayout(C.soundio_channel_layout_get_builtin(C.int(index)))
}

// ChannelLayoutGetDefault returns the default builtin channel layout for the given number of channels.
func ChannelLayoutGetDefault(channelCount int) *ChannelLayout {
	return newChannelLayout(C.soundio_channel_layout_get_default(C.int(channelCount)))
}

// BestMatchingLayout returns NULL if none matches.
// Iterates over preferredLayouts. Returns the first channel layout in
// preferredLayouts which matches one of the channel layouts in availableLayouts.
func BestMatchingLayout(device1 *Device, device2 *Device) *ChannelLayout {
	device1Ptr := device1.pointer()
	device2Ptr := device1.pointer()

	return newChannelLayout(C.soundio_best_matching_channel_layout(device1Ptr.layouts, device1Ptr.layout_count, device2Ptr.layouts, device2Ptr.layout_count))
}

// fields

// Name returns channel layout name.
func (l *ChannelLayout) Name() string {
	p := l.pointer()
	return C.GoString(p.name)
}

// ChannelCount returns channel count.
func (l *ChannelLayout) ChannelCount() int {
	p := l.pointer()
	return int(p.channel_count)
}

// Channels returns list of channelID.
func (l *ChannelLayout) Channels() []ChannelID {
	p := l.pointer()
	channels := make([]ChannelID, MaxChannels)
	for i := range channels {
		channels[i] = ChannelID(uint32(p.channels[i]))
	}
	return channels
}

// functions

// FindChannel returns the index of `channel` in `layout`, or `-1` if not found.
func (l *ChannelLayout) FindChannel(channel ChannelID) int {
	return int(C.soundio_channel_layout_find_channel(l.pointer(), uint32(channel)))
}

// DetectBuiltin returns whether it found a match.
// Populates the name field of layout if it matches a builtin one.
func (l *ChannelLayout) DetectBuiltin() bool {
	return bool(C.soundio_channel_layout_detect_builtin(l.pointer()))
}

// Equal returns whether the channel count field and each channel id matches in
// the supplied channel layouts.
func (l *ChannelLayout) Equal(o *ChannelLayout) bool {
	return bool(C.soundio_channel_layout_equal(l.pointer(), o.pointer()))
}

// SortChannelLayouts sorts by channel count, descending.
func (l *ChannelLayout) SortChannelLayouts(layoutCount int) {
	C.soundio_sort_channel_layouts(l.pointer(), C.int(layoutCount))
}

func (l *ChannelLayout) pointer() *C.struct_SoundIoChannelLayout {
	return (*C.struct_SoundIoChannelLayout)(unsafe.Pointer(l.ptr))
}

func newChannelLayout(l *C.struct_SoundIoChannelLayout) *ChannelLayout {
	return &ChannelLayout{
		ptr: uintptr(unsafe.Pointer(l)),
	}
}
