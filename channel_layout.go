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

type ChannelLayoutID uint32

type ChannelLayout struct {
	ptr uintptr
}

const (
	ChannelLayoutIDMono            ChannelLayoutID = C.SoundIoChannelLayoutIdMono
	ChannelLayoutIDStereo                          = C.SoundIoChannelLayoutIdStereo
	ChannelLayoutID2Point1                         = C.SoundIoChannelLayoutId2Point1
	ChannelLayoutID3Point0                         = C.SoundIoChannelLayoutId3Point0
	ChannelLayoutID3Point0Back                     = C.SoundIoChannelLayoutId3Point0Back
	ChannelLayoutID3Point1                         = C.SoundIoChannelLayoutId3Point1
	ChannelLayoutID4Point0                         = C.SoundIoChannelLayoutId4Point0
	ChannelLayoutIDQuad                            = C.SoundIoChannelLayoutIdQuad
	ChannelLayoutIDQuadSide                        = C.SoundIoChannelLayoutIdQuadSide
	ChannelLayoutID4Point1                         = C.SoundIoChannelLayoutId4Point1
	ChannelLayoutID5Point0Back                     = C.SoundIoChannelLayoutId5Point0Back
	ChannelLayoutID5Point0Side                     = C.SoundIoChannelLayoutId5Point0Side
	ChannelLayoutID5Point1                         = C.SoundIoChannelLayoutId5Point1
	ChannelLayoutID5Point1Back                     = C.SoundIoChannelLayoutId5Point1Back
	ChannelLayoutID6Point0Side                     = C.SoundIoChannelLayoutId6Point0Side
	ChannelLayoutID6Point0Front                    = C.SoundIoChannelLayoutId6Point0Front
	ChannelLayoutIDHexagonal                       = C.SoundIoChannelLayoutIdHexagonal
	ChannelLayoutID6Point1                         = C.SoundIoChannelLayoutId6Point1
	ChannelLayoutID6Point1Back                     = C.SoundIoChannelLayoutId6Point1Back
	ChannelLayoutID6Point1Front                    = C.SoundIoChannelLayoutId6Point1Front
	ChannelLayoutID7Point0                         = C.SoundIoChannelLayoutId7Point0
	ChannelLayoutID7Point0Front                    = C.SoundIoChannelLayoutId7Point0Front
	ChannelLayoutID7Point1                         = C.SoundIoChannelLayoutId7Point1
	ChannelLayoutID7Point1Wide                     = C.SoundIoChannelLayoutId7Point1Wide
	ChannelLayoutID7Point1WideBack                 = C.SoundIoChannelLayoutId7Point1WideBack
	ChannelLayoutIDOctagonal                       = C.SoundIoChannelLayoutIdOctagonal
)

// ChannelLayoutBuiltinCount returns the number of builtin channel layouts.
func ChannelLayoutBuiltinCount() int {
	return int(C.soundio_channel_layout_builtin_count())
}

// ChannelLayoutGetBuiltin returns a builtin channel layout.
// 0 <= `index` < ChannelLayoutBuiltinCount
func ChannelLayoutGetBuiltin(index ChannelLayoutID) *ChannelLayout {
	return createChannelLayout(C.soundio_channel_layout_get_builtin(C.int(index)))
}

// ChannelLayoutGetDefault returns the default builtin channel layout for the given number of channels.
func ChannelLayoutGetDefault(channelCount int) *ChannelLayout {
	return createChannelLayout(C.soundio_channel_layout_get_default(C.int(channelCount)))
}

// BestMatchingLayout returns NULL if none matches.
// Iterates over preferredLayouts. Returns the first channel layout in
// preferredLayouts which matches one of the channel layouts in availableLayouts.
func BestMatchingLayout(preferredLayouts []ChannelLayout, availableLayouts []ChannelLayout) *ChannelLayout {
	size := C.sizeof_struct_SoundIoChannelLayout
	preferredLayoutCount := len(preferredLayouts)
	preferredBuffer := C.malloc(C.size_t(size * preferredLayoutCount))
	defer C.free(preferredBuffer)
	for i := 0; i < preferredLayoutCount; i++ {
		C.memcpy(unsafe.Pointer(uintptr(preferredBuffer)+uintptr(i*size)), unsafe.Pointer(preferredLayouts[i].ptr), C.size_t(size))
	}

	availableLayoutCount := len(availableLayouts)
	availableBuffer := C.malloc(C.size_t(size * preferredLayoutCount))
	defer C.free(availableBuffer)
	for i := 0; i < availableLayoutCount; i++ {
		C.memcpy(unsafe.Pointer(uintptr(availableBuffer)+uintptr(i*size)), unsafe.Pointer(availableLayouts[i].ptr), C.size_t(size))
	}

	return createChannelLayout(C.soundio_best_matching_channel_layout(
		(*C.struct_SoundIoChannelLayout)(preferredBuffer), C.int(preferredLayoutCount),
		(*C.struct_SoundIoChannelLayout)(availableBuffer), C.int(availableLayoutCount)))
}

// fields

// GetName returns channel layout name.
func (l *ChannelLayout) GetName() string {
	p := l.getPointer()
	return C.GoString(p.name)
}

// GetChannelCount returns channel count.
func (l *ChannelLayout) GetChannelCount() int {
	p := l.getPointer()
	return int(p.channel_count)
}

// GetChannels returns list of channelID.
func (l *ChannelLayout) GetChannels() []ChannelID {
	p := l.getPointer()
	channels := make([]ChannelID, MaxChannels)
	for i := range channels {
		channels[i] = ChannelID(uint32(p.channels[i]))
	}
	return channels
}

// functions

// FindChannel returns the index of `channel` in `layout`, or `-1` if not found.
func (l *ChannelLayout) FindChannel(channel ChannelID) int {
	return int(C.soundio_channel_layout_find_channel(l.getPointer(), uint32(channel)))
}

// DetectBuiltin returns whether it found a match.
// Populates the name field of layout if it matches a builtin one.
func (l *ChannelLayout) DetectBuiltin() bool {
	return bool(C.soundio_channel_layout_detect_builtin(l.getPointer()))
}

// Equal returns whether the channel count field and each channel id matches in
// the supplied channel layouts.
func (l *ChannelLayout) Equal(o *ChannelLayout) bool {
	return bool(C.soundio_channel_layout_equal(l.getPointer(), o.getPointer()))
}

// SortChannelLayouts sorts by channel count, descending.
func (l *ChannelLayout) SortChannelLayouts(layoutCount int) {
	C.soundio_sort_channel_layouts(l.getPointer(), C.int(layoutCount))
}

func (l *ChannelLayout) getPointer() *C.struct_SoundIoChannelLayout {
	return (*C.struct_SoundIoChannelLayout)(unsafe.Pointer(l.ptr))
}

func createChannelLayout(l *C.struct_SoundIoChannelLayout) *ChannelLayout {
	return &ChannelLayout{
		ptr: uintptr(unsafe.Pointer(l)),
	}
}
