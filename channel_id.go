package soundio

/*
#include <soundio/soundio.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

// ChannelID ...
type ChannelID uint32

const (
	ChannelIDInvalid          ChannelID = C.SoundIoChannelIdInvalid
	ChannelIDFrontLeft                  = C.SoundIoChannelIdFrontLeft
	ChannelIDFrontRight                 = C.SoundIoChannelIdFrontRight
	ChannelIDFrontCenter                = C.SoundIoChannelIdFrontCenter
	ChannelIDLfe                        = C.SoundIoChannelIdLfe
	ChannelIDBackLeft                   = C.SoundIoChannelIdBackLeft
	ChannelIDBackRight                  = C.SoundIoChannelIdBackRight
	ChannelIDFrontLeftCenter            = C.SoundIoChannelIdFrontLeftCenter
	ChannelIDFrontRightCenter           = C.SoundIoChannelIdFrontRightCenter
	ChannelIDBackCenter                 = C.SoundIoChannelIdBackCenter
	ChannelIDSideLeft                   = C.SoundIoChannelIdSideLeft
	ChannelIDSideRight                  = C.SoundIoChannelIdSideRight
	ChannelIDTopCenter                  = C.SoundIoChannelIdTopCenter
	ChannelIDTopFrontLeft               = C.SoundIoChannelIdTopFrontLeft
	ChannelIDTopFrontCenter             = C.SoundIoChannelIdTopFrontCenter
	ChannelIDTopFrontRight              = C.SoundIoChannelIdTopFrontRight
	ChannelIDTopBackLeft                = C.SoundIoChannelIdTopBackLeft
	ChannelIDTopBackCenter              = C.SoundIoChannelIdTopBackCenter
	ChannelIDTopBackRight               = C.SoundIoChannelIdTopBackRight

	ChannelIDBackLeftCenter      = C.SoundIoChannelIdBackLeftCenter
	ChannelIDBackRightCenter     = C.SoundIoChannelIdBackRightCenter
	ChannelIDFrontLeftWide       = C.SoundIoChannelIdFrontLeftWide
	ChannelIDFrontRightWide      = C.SoundIoChannelIdFrontRightWide
	ChannelIDFrontLeftHigh       = C.SoundIoChannelIdFrontLeftHigh
	ChannelIDFrontCenterHigh     = C.SoundIoChannelIdFrontCenterHigh
	ChannelIDFrontRightHigh      = C.SoundIoChannelIdFrontRightHigh
	ChannelIDTopFrontLeftCenter  = C.SoundIoChannelIdTopFrontLeftCenter
	ChannelIDTopFrontRightCenter = C.SoundIoChannelIdTopFrontRightCenter
	ChannelIDTopSideLeft         = C.SoundIoChannelIdTopSideLeft
	ChannelIDTopSideRight        = C.SoundIoChannelIdTopSideRight
	ChannelIDLeftLfe             = C.SoundIoChannelIdLeftLfe
	ChannelIDRightLfe            = C.SoundIoChannelIdRightLfe
	ChannelIDLfe2                = C.SoundIoChannelIdLfe2
	ChannelIDBottomCenter        = C.SoundIoChannelIdBottomCenter
	ChannelIDBottomLeftCenter    = C.SoundIoChannelIdBottomLeftCenter
	ChannelIDBottomRightCenter   = C.SoundIoChannelIdBottomRightCenter

	ChannelIDMsMid  = C.SoundIoChannelIdMsMid  // Mid recording
	ChannelIDMsSide = C.SoundIoChannelIdMsSide // Side recording

	ChannelIDAmbisonicW = C.SoundIoChannelIdAmbisonicW
	ChannelIDAmbisonicX = C.SoundIoChannelIdAmbisonicX
	ChannelIDAmbisonicY = C.SoundIoChannelIdAmbisonicY
	ChannelIDAmbisonicZ = C.SoundIoChannelIdAmbisonicZ

	// ChannelIDXyX ... X of X-Y Recording
	ChannelIDXyX = C.SoundIoChannelIdXyX
	// ChannelIDXyY ... Y of X-Y Recording
	ChannelIDXyY = C.SoundIoChannelIdXyY

	ChannelIDHeadphonesLeft   = C.SoundIoChannelIdHeadphonesLeft
	ChannelIDHeadphonesRight  = C.SoundIoChannelIdHeadphonesRight
	ChannelIDClickTrack       = C.SoundIoChannelIdClickTrack
	ChannelIDForeignLanguage  = C.SoundIoChannelIdForeignLanguage
	ChannelIDHearingImpaired  = C.SoundIoChannelIdHearingImpaired
	ChannelIDNarration        = C.SoundIoChannelIdNarration
	ChannelIDHaptic           = C.SoundIoChannelIdHaptic
	ChannelIDDialogCentricMix = C.SoundIoChannelIdDialogCentricMix

	ChannelIDAux   = C.SoundIoChannelIdAux
	ChannelIDAux0  = C.SoundIoChannelIdAux0
	ChannelIDAux1  = C.SoundIoChannelIdAux1
	ChannelIDAux2  = C.SoundIoChannelIdAux2
	ChannelIDAux3  = C.SoundIoChannelIdAux3
	ChannelIDAux4  = C.SoundIoChannelIdAux4
	ChannelIDAux5  = C.SoundIoChannelIdAux5
	ChannelIDAux6  = C.SoundIoChannelIdAux6
	ChannelIDAux7  = C.SoundIoChannelIdAux7
	ChannelIDAux8  = C.SoundIoChannelIdAux8
	ChannelIDAux9  = C.SoundIoChannelIdAux9
	ChannelIDAux10 = C.SoundIoChannelIdAux10
	ChannelIDAux11 = C.SoundIoChannelIdAux11
	ChannelIDAux12 = C.SoundIoChannelIdAux12
	ChannelIDAux13 = C.SoundIoChannelIdAux13
	ChannelIDAux14 = C.SoundIoChannelIdAux14
	ChannelIDAux15 = C.SoundIoChannelIdAux15
)

func (c ChannelID) String() string {
	return C.GoString(C.soundio_get_channel_name(uint32(c)))
}

// functions

func ParseChannelId(str string) ChannelID {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	return ChannelID(uint32(C.soundio_parse_channel_id(cstr, C.int(len(str)))))
}
