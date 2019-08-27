#include "_cgo_export.h"

#include "soundio.h"

void setSoundIoCallback(struct SoundIo *io) {
	io->on_devices_change = soundioOnDevicesChange;
	io->on_backend_disconnect = soundioOnBackendDisconnect;
	io->on_events_signal = soundioOnEventsSignal;
}

void setInStreamCallback(struct SoundIoInStream *stream) {
	stream->read_callback = instreamReadCallbackDelegate;
	stream->overflow_callback = instreamOverflowCallbackDelegate;
	stream->error_callback = instreamErrorCallbackDelegate;
}

void setOutStreamCallback(struct SoundIoOutStream *stream) {
	stream->write_callback = outstreamWriteCallbackDelegate;
	stream->underflow_callback = outstreamUnderflowCallbackDelegate;
	stream->error_callback = outstreamErrorCallbackDelegate;
}
