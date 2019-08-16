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

type Format uint32

const (
	FormatInvalid   Format = C.SoundIoFormatInvalid
	FormatS8               = C.SoundIoFormatS8        // Signed 8 bit
	FormatU8               = C.SoundIoFormatU8        // Unsigned 8 bit
	FormatS16LE            = C.SoundIoFormatS16LE     // Signed 16 bit Little Endian
	FormatS16BE            = C.SoundIoFormatS16BE     // Signed 16 bit Big Endian
	FormatU16LE            = C.SoundIoFormatU16LE     // Unsigned 16 bit Little Endian
	FormatU16BE            = C.SoundIoFormatU16BE     // Unsigned 16 bit Big Endian
	FormatS24LE            = C.SoundIoFormatS24LE     // Signed 24 bit Little Endian using low three bytes in 32-bit word
	FormatS24BE            = C.SoundIoFormatS24BE     // Signed 24 bit Big Endian using low three bytes in 32-bit word
	FormatU24LE            = C.SoundIoFormatU24LE     // Unsigned 24 bit Little Endian using low three bytes in 32-bit word
	FormatU24BE            = C.SoundIoFormatU24BE     // Unsigned 24 bit Big Endian using low three bytes in 32-bit word
	FormatS32LE            = C.SoundIoFormatS32LE     // Signed 32 bit Little Endian
	FormatS32BE            = C.SoundIoFormatS32BE     // Signed 32 bit Big Endian
	FormatU32LE            = C.SoundIoFormatU32LE     // Unsigned 32 bit Little Endian
	FormatU32BE            = C.SoundIoFormatU32BE     // Unsigned 32 bit Big Endian
	FormatFloat32LE        = C.SoundIoFormatFloat32LE // Float 32 bit Little Endian, Range -1.0 to 1.0
	FormatFloat32BE        = C.SoundIoFormatFloat32BE // Float 32 bit Big Endian, Range -1.0 to 1.0
	FormatFloat64LE        = C.SoundIoFormatFloat64LE // Float 64 bit Little Endian, Range -1.0 to 1.0
	FormatFloat64BE        = C.SoundIoFormatFloat64BE // Float 64 bit Big Endian, Range -1.0 to 1.0

	FormatS16FE     = C.SoundIoFormatS16NE     // Signed 16 bit Native Endian
	FormatS16NE     = C.SoundIoFormatS16FE     // Signed 16 bit Foreign Endian
	FormatU16FE     = C.SoundIoFormatU16NE     // Unsigned 16 bit Native Endian
	FormatU16NE     = C.SoundIoFormatU16FE     // Unsigned 16 bit Foreign Endian
	FormatS24FE     = C.SoundIoFormatS24NE     // Signed 24 bit Native Endian using low three bytes in 32-bit word
	FormatS24NE     = C.SoundIoFormatS24FE     // Signed 24 bit Foreign Endian using low three bytes in 32-bit word
	FormatU24FE     = C.SoundIoFormatU24NE     // Unsigned 24 bit Native Endian using low three bytes in 32-bit word
	FormatU24NE     = C.SoundIoFormatU24FE     // Unsigned 24 bit Foreign Endian using low three bytes in 32-bit word
	FormatS32FE     = C.SoundIoFormatS32NE     // Signed 32 bit Native Endian
	FormatS32NE     = C.SoundIoFormatS32FE     // Signed 32 bit Foreign Endian
	FormatU32FE     = C.SoundIoFormatU32NE     // Unsigned 32 bit Native Endian
	FormatU32NE     = C.SoundIoFormatU32FE     // Unsigned 32 bit Foreign Endian
	FormatFloat32FE = C.SoundIoFormatFloat32NE // Float 32 bit Native Endian, Range -1.0 to 1.0
	FormatFloat32NE = C.SoundIoFormatFloat32FE // Float 32 bit Foreign Endian, Range -1.0 to 1.0
	FormatFloat64NE = C.SoundIoFormatFloat64NE // Float 64 bit Native Endian, Range -1.0 to 1.0
	FormatFloat64FE = C.SoundIoFormatFloat64FE // Float 64 bit Foreign Endian, Range -1.0 to 1.0
)

func (f Format) String() string {
	return C.GoString(C.soundio_format_string(uint32(f)))
}
