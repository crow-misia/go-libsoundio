/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */

package soundio

/*
#include <soundio/soundio.h>
*/
import "C"

// Format is audio format.
type Format uint32

// Format enumeration.
const (
	FormatInvalid   = Format(C.SoundIoFormatInvalid)
	FormatS8        = Format(C.SoundIoFormatS8)        // Signed 8 bit
	FormatU8        = Format(C.SoundIoFormatU8)        // Unsigned 8 bit
	FormatS16LE     = Format(C.SoundIoFormatS16LE)     // Signed 16 bit Little Endian
	FormatS16BE     = Format(C.SoundIoFormatS16BE)     // Signed 16 bit Big Endian
	FormatU16LE     = Format(C.SoundIoFormatU16LE)     // Unsigned 16 bit Little Endian
	FormatU16BE     = Format(C.SoundIoFormatU16BE)     // Unsigned 16 bit Big Endian
	FormatS24LE     = Format(C.SoundIoFormatS24LE)     // Signed 24 bit Little Endian using low three bytes in 32-bit word
	FormatS24BE     = Format(C.SoundIoFormatS24BE)     // Signed 24 bit Big Endian using low three bytes in 32-bit word
	FormatU24LE     = Format(C.SoundIoFormatU24LE)     // Unsigned 24 bit Little Endian using low three bytes in 32-bit word
	FormatU24BE     = Format(C.SoundIoFormatU24BE)     // Unsigned 24 bit Big Endian using low three bytes in 32-bit word
	FormatS32LE     = Format(C.SoundIoFormatS32LE)     // Signed 32 bit Little Endian
	FormatS32BE     = Format(C.SoundIoFormatS32BE)     // Signed 32 bit Big Endian
	FormatU32LE     = Format(C.SoundIoFormatU32LE)     // Unsigned 32 bit Little Endian
	FormatU32BE     = Format(C.SoundIoFormatU32BE)     // Unsigned 32 bit Big Endian
	FormatFloat32LE = Format(C.SoundIoFormatFloat32LE) // Float 32 bit Little Endian, Range -1.0 to 1.0
	FormatFloat32BE = Format(C.SoundIoFormatFloat32BE) // Float 32 bit Big Endian, Range -1.0 to 1.0
	FormatFloat64LE = Format(C.SoundIoFormatFloat64LE) // Float 64 bit Little Endian, Range -1.0 to 1.0
	FormatFloat64BE = Format(C.SoundIoFormatFloat64BE) // Float 64 bit Big Endian, Range -1.0 to 1.0

	FormatS16FE     = Format(C.SoundIoFormatS16NE)     // Signed 16 bit Native Endian
	FormatS16NE     = Format(C.SoundIoFormatS16FE)     // Signed 16 bit Foreign Endian
	FormatU16FE     = Format(C.SoundIoFormatU16NE)     // Unsigned 16 bit Native Endian
	FormatU16NE     = Format(C.SoundIoFormatU16FE)     // Unsigned 16 bit Foreign Endian
	FormatS24FE     = Format(C.SoundIoFormatS24NE)     // Signed 24 bit Native Endian using low three bytes in 32-bit word
	FormatS24NE     = Format(C.SoundIoFormatS24FE)     // Signed 24 bit Foreign Endian using low three bytes in 32-bit word
	FormatU24FE     = Format(C.SoundIoFormatU24NE)     // Unsigned 24 bit Native Endian using low three bytes in 32-bit word
	FormatU24NE     = Format(C.SoundIoFormatU24FE)     // Unsigned 24 bit Foreign Endian using low three bytes in 32-bit word
	FormatS32FE     = Format(C.SoundIoFormatS32NE)     // Signed 32 bit Native Endian
	FormatS32NE     = Format(C.SoundIoFormatS32FE)     // Signed 32 bit Foreign Endian
	FormatU32FE     = Format(C.SoundIoFormatU32NE)     // Unsigned 32 bit Native Endian
	FormatU32NE     = Format(C.SoundIoFormatU32FE)     // Unsigned 32 bit Foreign Endian
	FormatFloat32FE = Format(C.SoundIoFormatFloat32NE) // Float 32 bit Native Endian, Range -1.0 to 1.0
	FormatFloat32NE = Format(C.SoundIoFormatFloat32FE) // Float 32 bit Foreign Endian, Range -1.0 to 1.0
	FormatFloat64NE = Format(C.SoundIoFormatFloat64NE) // Float 64 bit Native Endian, Range -1.0 to 1.0
	FormatFloat64FE = Format(C.SoundIoFormatFloat64FE) // Float 64 bit Foreign Endian, Range -1.0 to 1.0
)

func (f Format) String() string {
	return C.GoString(C.soundio_format_string(uint32(f)))
}
