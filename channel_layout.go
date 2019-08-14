package soundio

/*
#include <soundio/soundio.h>
*/
import "C"

type ChannelLayoutID uint32

type ChannelLayout struct {
	ptr *C.struct_SoundIoChannelLayout
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

func ChannelLayoutGetBuiltin(index ChannelLayoutID) *ChannelLayout {
	return &ChannelLayout{
		ptr: C.soundio_channel_layout_get_builtin(C.int(index)),
	}
}

func ChannelLayoutGetDefault(channelCount int) *ChannelLayout {
	return &ChannelLayout{
		ptr: C.soundio_channel_layout_get_default(C.int(channelCount)),
	}
}

func ChannelLayoutBuiltinCount() int {
	return int(C.soundio_channel_layout_builtin_count())
}

func BestMatchingLayout(preferredLayouts []ChannelLayout, availableLayouts []ChannelLayout) *ChannelLayout {
	preferredLayoutCount := len(preferredLayouts)
	preferredLayoutsPtr := make([]C.struct_SoundIoChannelLayout, preferredLayoutCount)
	for i := 0; i < preferredLayoutCount; i++ {
		preferredLayoutsPtr[i] = *(preferredLayouts[i].ptr)
	}

	availableLayoutCount := len(availableLayouts)
	availableLayoutsPtr := make([]C.struct_SoundIoChannelLayout, availableLayoutCount)
	for i := 0; i < availableLayoutCount; i++ {
		availableLayoutsPtr[i] = *(availableLayouts[i].ptr)
	}

	return &ChannelLayout{
		ptr: C.soundio_best_matching_channel_layout(&preferredLayoutsPtr[0], C.int(preferredLayoutCount), &availableLayoutsPtr[0], C.int(availableLayoutCount)),
	}
}

// fields

func (l *ChannelLayout) GetName() string {
	return C.GoString(l.ptr.name)
}

func (l *ChannelLayout) GetChannelCount() int {
	return int(l.ptr.channel_count)
}

func (l *ChannelLayout) GetChannels() *[]ChannelID {
	channels := make([]ChannelID, MaxChannels)
	for i := range channels {
		channels[i] = ChannelID(uint32(l.ptr.channels[i]))
	}
	return &channels
}

// functions

func (l *ChannelLayout) FindChannel(channel ChannelID) int {
	return int(C.soundio_channel_layout_find_channel(l.ptr, uint32(channel)))
}

func (l *ChannelLayout) DetectBuiltin() bool {
	return bool(C.soundio_channel_layout_detect_builtin(l.ptr))
}

func (l *ChannelLayout) Equal(o *ChannelLayout) bool {
	return bool(C.soundio_channel_layout_equal(l.ptr, o.ptr))
}

func (l *ChannelLayout) SortChannelLayouts(layoutCount int) {
	C.soundio_sort_channel_layouts(l.ptr, C.int(layoutCount))
}
