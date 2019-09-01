package soundio

// Option is SoundIo option function.
type Option func(*SoundIo)

// WithBackend is backend setter.
func WithBackend(backend Backend) Option {
	return func(io *SoundIo) {
		io.backend = backend
	}
}

// WithAppName is application name setter.
func WithAppName(appName string) Option {
	return func(io *SoundIo) {
		io.appName = appName
	}
}

// WithOnDevicesChange is onDeviceChange callback setter.
func WithOnDevicesChange(callback func(io *SoundIo)) Option {
	return func(io *SoundIo) {
		io.onDevicesChange = callback
	}
}

// WithOnBackendDisconnect is onBackendDisconnect callback setter.
func WithOnBackendDisconnect(callback func(io *SoundIo, err error)) Option {
	return func(io *SoundIo) {
		io.onBackendDisconnect = callback
	}
}

// WithOnEventsSignal is onEventsSignal callback setter.
func WithOnEventsSignal(callback func(io *SoundIo)) Option {
	return func(io *SoundIo) {
		io.onEventsSignal = callback
	}
}
