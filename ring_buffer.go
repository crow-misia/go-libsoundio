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
import "unsafe"

type RingBuffer struct {
	ptr uintptr
}

// functions

// Destroy releases resources.
func (r *RingBuffer) Destroy() {
	if r.ptr != 0 {
		C.soundio_ring_buffer_destroy(r.getPointer())
		r.ptr = 0
	}
}

// Capacity returns the actual capacity of ring buffer.
// When you create a ring buffer, capacity might be more than the requested capacity for alignment purposes.
func (r *RingBuffer) Capacity() int {
	return int(C.soundio_ring_buffer_capacity(r.getPointer()))
}

// WritePtr returns writable pointer.
// Do not write more than capacity.
func (r *RingBuffer) WritePtr() uintptr {
	return uintptr(unsafe.Pointer(C.soundio_ring_buffer_write_ptr(r.getPointer())))
}

// AdvanceWritePtr advance `count` in bytes.
func (r *RingBuffer) AdvanceWritePtr(count int) {
	C.soundio_ring_buffer_advance_write_ptr(r.getPointer(), C.int(count))
}

// ReadPtr returns readable pointer.
// Do not read more than capacity.
func (r *RingBuffer) ReadPtr() uintptr {
	return uintptr(unsafe.Pointer(C.soundio_ring_buffer_read_ptr(r.getPointer())))
}

// AdvanceReadPtr advance `count` in bytes.
func (r *RingBuffer) AdvanceReadPtr(count int) {
	C.soundio_ring_buffer_advance_read_ptr(r.getPointer(), C.int(count))
}

// FillCount returns how many bytes of the buffer is used, ready for reading.
func (r *RingBuffer) FillCount(count int) int {
	return int(C.soundio_ring_buffer_fill_count(r.getPointer()))
}

// FreeCount returns how many bytes of the buffer is free, ready for writing.
func (r *RingBuffer) FreeCount(count int) int {
	return int(C.soundio_ring_buffer_free_count(r.getPointer()))
}

// Clear ring buffer.
// Must be called by the writer.
func (r *RingBuffer) Clear() {
	C.soundio_ring_buffer_clear(r.getPointer())
}

func (r *RingBuffer) getPointer() *C.struct_SoundIoRingBuffer {
	return (*C.struct_SoundIoRingBuffer)(unsafe.Pointer(r.ptr))
}
