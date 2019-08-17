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
import "unsafe"

// ChannelID is channel id.
type ChannelID uint32

const (
	ChannelIDInvalid          = ChannelID(C.SoundIoChannelIdInvalid)
	ChannelIDFrontLeft        = ChannelID(C.SoundIoChannelIdFrontLeft)
	ChannelIDFrontRight       = ChannelID(C.SoundIoChannelIdFrontRight)
	ChannelIDFrontCenter      = ChannelID(C.SoundIoChannelIdFrontCenter)
	ChannelIDLfe              = ChannelID(C.SoundIoChannelIdLfe)
	ChannelIDBackLeft         = ChannelID(C.SoundIoChannelIdBackLeft)
	ChannelIDBackRight        = ChannelID(C.SoundIoChannelIdBackRight)
	ChannelIDFrontLeftCenter  = ChannelID(C.SoundIoChannelIdFrontLeftCenter)
	ChannelIDFrontRightCenter = ChannelID(C.SoundIoChannelIdFrontRightCenter)
	ChannelIDBackCenter       = ChannelID(C.SoundIoChannelIdBackCenter)
	ChannelIDSideLeft         = ChannelID(C.SoundIoChannelIdSideLeft)
	ChannelIDSideRight        = ChannelID(C.SoundIoChannelIdSideRight)
	ChannelIDTopCenter        = ChannelID(C.SoundIoChannelIdTopCenter)
	ChannelIDTopFrontLeft     = ChannelID(C.SoundIoChannelIdTopFrontLeft)
	ChannelIDTopFrontCenter   = ChannelID(C.SoundIoChannelIdTopFrontCenter)
	ChannelIDTopFrontRight    = ChannelID(C.SoundIoChannelIdTopFrontRight)
	ChannelIDTopBackLeft      = ChannelID(C.SoundIoChannelIdTopBackLeft)
	ChannelIDTopBackCenter    = ChannelID(C.SoundIoChannelIdTopBackCenter)
	ChannelIDTopBackRight     = ChannelID(C.SoundIoChannelIdTopBackRight)

	ChannelIDBackLeftCenter      = ChannelID(C.SoundIoChannelIdBackLeftCenter)
	ChannelIDBackRightCenter     = ChannelID(C.SoundIoChannelIdBackRightCenter)
	ChannelIDFrontLeftWide       = ChannelID(C.SoundIoChannelIdFrontLeftWide)
	ChannelIDFrontRightWide      = ChannelID(C.SoundIoChannelIdFrontRightWide)
	ChannelIDFrontLeftHigh       = ChannelID(C.SoundIoChannelIdFrontLeftHigh)
	ChannelIDFrontCenterHigh     = ChannelID(C.SoundIoChannelIdFrontCenterHigh)
	ChannelIDFrontRightHigh      = ChannelID(C.SoundIoChannelIdFrontRightHigh)
	ChannelIDTopFrontLeftCenter  = ChannelID(C.SoundIoChannelIdTopFrontLeftCenter)
	ChannelIDTopFrontRightCenter = ChannelID(C.SoundIoChannelIdTopFrontRightCenter)
	ChannelIDTopSideLeft         = ChannelID(C.SoundIoChannelIdTopSideLeft)
	ChannelIDTopSideRight        = ChannelID(C.SoundIoChannelIdTopSideRight)
	ChannelIDLeftLfe             = ChannelID(C.SoundIoChannelIdLeftLfe)
	ChannelIDRightLfe            = ChannelID(C.SoundIoChannelIdRightLfe)
	ChannelIDLfe2                = ChannelID(C.SoundIoChannelIdLfe2)
	ChannelIDBottomCenter        = ChannelID(C.SoundIoChannelIdBottomCenter)
	ChannelIDBottomLeftCenter    = ChannelID(C.SoundIoChannelIdBottomLeftCenter)
	ChannelIDBottomRightCenter   = ChannelID(C.SoundIoChannelIdBottomRightCenter)

	ChannelIDMsMid  = ChannelID(C.SoundIoChannelIdMsMid)  // Mid recording
	ChannelIDMsSide = ChannelID(C.SoundIoChannelIdMsSide) // Side recording

	ChannelIDAmbisonicW = ChannelID(C.SoundIoChannelIdAmbisonicW)
	ChannelIDAmbisonicX = ChannelID(C.SoundIoChannelIdAmbisonicX)
	ChannelIDAmbisonicY = ChannelID(C.SoundIoChannelIdAmbisonicY)
	ChannelIDAmbisonicZ = ChannelID(C.SoundIoChannelIdAmbisonicZ)

	// ChannelIDXyX is X of X-Y Recording
	ChannelIDXyX = ChannelID(C.SoundIoChannelIdXyX)
	// ChannelIDXyY is Y of X-Y Recording
	ChannelIDXyY = ChannelID(C.SoundIoChannelIdXyY)

	ChannelIDHeadphonesLeft   = ChannelID(C.SoundIoChannelIdHeadphonesLeft)
	ChannelIDHeadphonesRight  = ChannelID(C.SoundIoChannelIdHeadphonesRight)
	ChannelIDClickTrack       = ChannelID(C.SoundIoChannelIdClickTrack)
	ChannelIDForeignLanguage  = ChannelID(C.SoundIoChannelIdForeignLanguage)
	ChannelIDHearingImpaired  = ChannelID(C.SoundIoChannelIdHearingImpaired)
	ChannelIDNarration        = ChannelID(C.SoundIoChannelIdNarration)
	ChannelIDHaptic           = ChannelID(C.SoundIoChannelIdHaptic)
	ChannelIDDialogCentricMix = ChannelID(C.SoundIoChannelIdDialogCentricMix)

	ChannelIDAux   = ChannelID(C.SoundIoChannelIdAux)
	ChannelIDAux0  = ChannelID(C.SoundIoChannelIdAux0)
	ChannelIDAux1  = ChannelID(C.SoundIoChannelIdAux1)
	ChannelIDAux2  = ChannelID(C.SoundIoChannelIdAux2)
	ChannelIDAux3  = ChannelID(C.SoundIoChannelIdAux3)
	ChannelIDAux4  = ChannelID(C.SoundIoChannelIdAux4)
	ChannelIDAux5  = ChannelID(C.SoundIoChannelIdAux5)
	ChannelIDAux6  = ChannelID(C.SoundIoChannelIdAux6)
	ChannelIDAux7  = ChannelID(C.SoundIoChannelIdAux7)
	ChannelIDAux8  = ChannelID(C.SoundIoChannelIdAux8)
	ChannelIDAux9  = ChannelID(C.SoundIoChannelIdAux9)
	ChannelIDAux10 = ChannelID(C.SoundIoChannelIdAux10)
	ChannelIDAux11 = ChannelID(C.SoundIoChannelIdAux11)
	ChannelIDAux12 = ChannelID(C.SoundIoChannelIdAux12)
	ChannelIDAux13 = ChannelID(C.SoundIoChannelIdAux13)
	ChannelIDAux14 = ChannelID(C.SoundIoChannelIdAux14)
	ChannelIDAux15 = ChannelID(C.SoundIoChannelIdAux15)
)

func (c ChannelID) String() string {
	return C.GoString(C.soundio_get_channel_name(uint32(c)))
}

// functions

// ParseChannelID returns ChannelID from string.
func ParseChannelID(str string) ChannelID {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	return ChannelID(uint32(C.soundio_parse_channel_id(cstr, C.int(len(str)))))
}
