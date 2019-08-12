package soundio

/*
#include <soundio/soundio.h>
*/
import "C"
import "unsafe"

type RingBuffer struct {
	ptr *C.struct_SoundIoRingBuffer
}

// functions

func (r *RingBuffer) Destroy() {
	C.soundio_ring_buffer_destroy(r.ptr)
	r.ptr = nil
}

func (r *RingBuffer) Capacity() int {
	return int(C.soundio_ring_buffer_capacity(r.ptr))
}

func (r *RingBuffer) WritePtr() uintptr {
	return uintptr(unsafe.Pointer(C.soundio_ring_buffer_write_ptr(r.ptr)))
}

func (r *RingBuffer) AdvanceWritePtr(count int) {
	C.soundio_ring_buffer_advance_write_ptr(r.ptr, C.int(count))
}

func (r *RingBuffer) ReadPtr() uintptr {
	return uintptr(unsafe.Pointer(C.soundio_ring_buffer_read_ptr(r.ptr)))
}

func (r *RingBuffer) AdvanceReadPtr(count int) {
	C.soundio_ring_buffer_advance_read_ptr(r.ptr, C.int(count))
}

func (r *RingBuffer) FillCount(count int) int {
	return int(C.soundio_ring_buffer_fill_count(r.ptr))
}

func (r *RingBuffer) FreeCount(count int) int {
	return int(C.soundio_ring_buffer_free_count(r.ptr))
}

func (r *RingBuffer) Clear() {
	C.soundio_ring_buffer_clear(r.ptr)
}
