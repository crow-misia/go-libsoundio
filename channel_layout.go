package soundio

/*
#include <soundio/soundio.h>
*/
import "C"

type ChannelLayoutId uint32

type ChannelLayout struct {
	ptr *C.struct_SoundIoChannelLayout
}

const (
	ChannelLayoutIdMono            ChannelLayoutId = C.SoundIoChannelLayoutIdMono
	ChannelLayoutIdStereo                          = C.SoundIoChannelLayoutIdStereo
	ChannelLayoutId2Point1                         = C.SoundIoChannelLayoutId2Point1
	ChannelLayoutId3Point0                         = C.SoundIoChannelLayoutId3Point0
	ChannelLayoutId3Point0Back                     = C.SoundIoChannelLayoutId3Point0Back
	ChannelLayoutId3Point1                         = C.SoundIoChannelLayoutId3Point1
	ChannelLayoutId4Point0                         = C.SoundIoChannelLayoutId4Point0
	ChannelLayoutIdQuad                            = C.SoundIoChannelLayoutIdQuad
	ChannelLayoutIdQuadSide                        = C.SoundIoChannelLayoutIdQuadSide
	ChannelLayoutId4Point1                         = C.SoundIoChannelLayoutId4Point1
	ChannelLayoutId5Point0Back                     = C.SoundIoChannelLayoutId5Point0Back
	ChannelLayoutId5Point0Side                     = C.SoundIoChannelLayoutId5Point0Side
	ChannelLayoutId5Point1                         = C.SoundIoChannelLayoutId5Point1
	ChannelLayoutId5Point1Back                     = C.SoundIoChannelLayoutId5Point1Back
	ChannelLayoutId6Point0Side                     = C.SoundIoChannelLayoutId6Point0Side
	ChannelLayoutId6Point0Front                    = C.SoundIoChannelLayoutId6Point0Front
	ChannelLayoutIdHexagonal                       = C.SoundIoChannelLayoutIdHexagonal
	ChannelLayoutId6Point1                         = C.SoundIoChannelLayoutId6Point1
	ChannelLayoutId6Point1Back                     = C.SoundIoChannelLayoutId6Point1Back
	ChannelLayoutId6Point1Front                    = C.SoundIoChannelLayoutId6Point1Front
	ChannelLayoutId7Point0                         = C.SoundIoChannelLayoutId7Point0
	ChannelLayoutId7Point0Front                    = C.SoundIoChannelLayoutId7Point0Front
	ChannelLayoutId7Point1                         = C.SoundIoChannelLayoutId7Point1
	ChannelLayoutId7Point1Wide                     = C.SoundIoChannelLayoutId7Point1Wide
	ChannelLayoutId7Point1WideBack                 = C.SoundIoChannelLayoutId7Point1WideBack
	ChannelLayoutIdOctagonal                       = C.SoundIoChannelLayoutIdOctagonal
)

func ChannelLayoutGetBuiltin(index int) *ChannelLayout {
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
