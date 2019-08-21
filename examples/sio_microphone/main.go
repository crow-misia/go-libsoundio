/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	soundio "github.com/crow-misia/go-libsoundio"
	"github.com/glycerine/rbuf"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var overflowCount = 0
var underflowCount = 0
var exitCode = 0

var prioritizedFormats = []soundio.Format{
	soundio.FormatFloat32NE,
	soundio.FormatFloat32FE,
	soundio.FormatS32NE,
	soundio.FormatS32FE,
	soundio.FormatS24NE,
	soundio.FormatS24FE,
	soundio.FormatS16NE,
	soundio.FormatS16FE,
	soundio.FormatFloat64NE,
	soundio.FormatFloat64FE,
	soundio.FormatU32NE,
	soundio.FormatU32FE,
	soundio.FormatU24NE,
	soundio.FormatU24FE,
	soundio.FormatU16NE,
	soundio.FormatU16FE,
	soundio.FormatS8,
	soundio.FormatU8,
}

var prioritizedSampleRates = []int{
	48000,
	44100,
	96000,
	24000,
}

func main() {
	var (
		backend        string
		inputDeviceId  string
		inputIsRaw     bool
		outputDeviceId string
		outputIsRaw    bool
		latencySec     float64
	)
	flag.NewFlagSet("help", flag.ExitOnError)
	flag.StringVar(&backend, "backend", "", "dummy|alsa|pulseaudio|jack|coreaudio|wasapi")
	flag.StringVar(&inputDeviceId, "in-device", "", "id")
	flag.BoolVar(&inputIsRaw, "in-raw", false, "raw")
	flag.StringVar(&outputDeviceId, "out-device", "", "id")
	flag.BoolVar(&outputIsRaw, "out-raw", false, "raw")
	flag.Float64Var(&latencySec, "latency-sec", 0.2, "latency seconds")
	flag.Parse()

	enumBackend, err := parseBackend(backend)
	if err != nil {
		log.Println(err)
		exitCode = 1
	} else {
		ctx := context.Background()
		parentCtx := signalContext(ctx)
		err := realMain(parentCtx, enumBackend, inputDeviceId, inputIsRaw, outputDeviceId, outputIsRaw, latencySec)
		if err != nil {
			exitCode = 1
			log.Println(err)
		}
		parentCtx.Done()
	}

	os.Exit(exitCode)
}

func parseBackend(str string) (soundio.Backend, error) {
	switch strings.ToLower(str) {
	case "":
		return soundio.BackendNone, nil
	case "dummy":
		return soundio.BackendDummy, nil
	case "alsa":
		return soundio.BackendAlsa, nil
	case "pulseaudio":
		return soundio.BackendPulseAudio, nil
	case "jack":
		return soundio.BackendJack, nil
	case "coreaudio":
		return soundio.BackendCoreAudio, nil
	case "wasapi":
		return soundio.BackendWasapi, nil
	default:
		return soundio.BackendNone, fmt.Errorf("invalid backend: %s", str)
	}
}

func selectDevice(s *soundio.SoundIo, deviceId string, isRaw bool, getDeviceCount func(io *soundio.SoundIo) int, getDefaultIndex func(io *soundio.SoundIo) int, getDevice func(io *soundio.SoundIo, index int) *soundio.Device) (*soundio.Device, error) {
	var selectedDevice *soundio.Device
	if len(deviceId) > 0 {
		count := getDeviceCount(s)
		for i := 0; i < count; i++ {
			device := getDevice(s, i)
			if device.Raw() == isRaw && deviceId == device.ID() {
				selectedDevice = device
				break
			}
			device.RemoveReference()
		}
		if selectedDevice == nil {
			return nil, fmt.Errorf("invalid device id: %s", deviceId)
		}
	} else {
		deviceIndex := getDefaultIndex(s)
		selectedDevice = getDevice(s, deviceIndex)
		if selectedDevice == nil {
			return nil, errors.New("no input devices available")
		}
	}
	return selectedDevice, nil
}

func realMain(ctx context.Context, backend soundio.Backend, inputDeviceId string, inputIsRaw bool, outputDeviceId string, outputIsRaw bool, latencySec float64) error {
	_, cancelParent := context.WithCancel(ctx)
	defer cancelParent()

	s := soundio.Create()

	var err error
	if backend == soundio.BackendNone {
		err = s.Connect()
	} else {
		err = s.ConnectBackend(backend)
	}
	if err != nil {
		return err
	}

	s.FlushEvents()

	selectedInputDevice, err := selectDevice(s, inputDeviceId, inputIsRaw, func(io *soundio.SoundIo) int {
		return io.InputDeviceCount()
	}, func(io *soundio.SoundIo) int {
		return io.DefaultInputDeviceIndex()
	}, func(io *soundio.SoundIo, index int) *soundio.Device {
		return io.InputDevice(index)
	})
	if err != nil {
		return err
	}
	defer selectedInputDevice.RemoveReference()
	log.Printf("Input device: %s", selectedInputDevice.Name())
	if selectedInputDevice.ProbeError() != nil {
		return fmt.Errorf("unable to probe device: %s", selectedInputDevice.ProbeError())
	}

	selectedOutputDevice, err := selectDevice(s, outputDeviceId, outputIsRaw, func(io *soundio.SoundIo) int {
		return io.OutputDeviceCount()
	}, func(io *soundio.SoundIo) int {
		return io.DefaultOutputDeviceIndex()
	}, func(io *soundio.SoundIo, index int) *soundio.Device {
		return io.OutputDevice(index)
	})
	if err != nil {
		return err
	}
	defer selectedOutputDevice.RemoveReference()
	log.Printf("Output device: %s", selectedOutputDevice.Name())
	if selectedOutputDevice.ProbeError() != nil {
		return fmt.Errorf("unable to probe device: %s", selectedOutputDevice.ProbeError())
	}

	selectedInputDevice.SortChannelLayouts()
	selectedOutputDevice.SortChannelLayouts()
	layout := soundio.BestMatchingLayout(selectedOutputDevice, selectedInputDevice)
	if layout == nil {
		return errors.New("channel layouts not compatible")
	}
	channels := layout.ChannelCount()

	sampleRate := 0
	for _, rate := range prioritizedSampleRates {
		if selectedInputDevice.SupportsSampleRate(rate) && selectedOutputDevice.SupportsSampleRate(rate) {
			sampleRate = rate
			break
		}
	}
	if sampleRate == 0 {
		return errors.New("incompatible sample rates")
	}
	log.Printf("Sample rate: %d", sampleRate)

	format := soundio.FormatInvalid
	for _, f := range prioritizedFormats {
		if selectedInputDevice.SupportsFormat(f) && selectedOutputDevice.SupportsFormat(f) {
			format = f
			break
		}
	}
	if format == soundio.FormatInvalid {
		return errors.New("incompatible sample formats")
	}
	log.Printf("Format: %s", format)

	var ringBuffer *rbuf.FixedSizeRingBuf

	instream := selectedInputDevice.NewInStream()
	defer instream.Destroy()
	instream.SetFormat(format)
	instream.SetLayout(layout)
	instream.SetSampleRate(sampleRate)
	instream.SetSoftwareLatency(latencySec)
	instream.SetReadCallback(func(stream *soundio.InStream, frameCountMin int, frameCountMax int) {
		frameBytes := layout.ChannelCount() * instream.BytesPerFrame()
		freeBytes := ringBuffer.N - ringBuffer.Readable
		freeCount := freeBytes / frameBytes
		if frameCountMin > freeCount {
			log.Println("ring buffer overflow")
			return
		}
		writeFrames := freeCount
		if writeFrames > frameCountMax {
			writeFrames = frameCountMax
		}

		frameLeft := writeFrames

		for {
			frameCount := frameLeft
			if frameLeft <= 0 {
				break
			}

			areas, err := stream.BeginRead(&frameCount)
			if err != nil {
				log.Printf("begin read error: %s", err)
				cancelParent()
				return
			}
			if frameCount <= 0 {
				break
			}
			if areas != nil {
				for frame := 0; frame < frameCount; frame++ {
					for ch := 0; ch < channels; ch++ {
						buffer := areas.Buffer(ch, frame)
						_, err = ringBuffer.Write(buffer)
						if err != nil {
							log.Printf("ringbuffer write error: %s %d", err, len(buffer))
						}
					}
				}
			}
			err = stream.EndRead()
			if err != nil {
				log.Printf("end read error: %s", err)
				cancelParent()
				return
			}

			frameLeft -= frameCount
		}
	})
	instream.SetOverflowCallback(func(stream *soundio.InStream) {
		overflowCount++
		log.Printf("overflow %d", overflowCount)
	})
	err = instream.Open()
	if err != nil {
		return fmt.Errorf("unable to open input device: %s", err)
	}

	outstream := selectedOutputDevice.NewOutStream()
	defer outstream.Destroy()
	log.Printf("format: %s", format)
	log.Printf("layout name: %s", layout.Name())
	log.Printf("layout channel count: %d", channels)
	log.Printf("sample rate: %d", sampleRate)
	log.Printf("latency seconds: %f sec", latencySec)
	outstream.SetFormat(format)
	outstream.SetLayout(layout)
	outstream.SetSampleRate(sampleRate)
	outstream.SetSoftwareLatency(latencySec)
	outstream.SetWriteCallback(func(stream *soundio.OutStream, frameCountMin int, frameCountMax int) {
		channelCount := stream.Layout().ChannelCount()

		fillBytes := ringBuffer.Readable
		fillCount := fillBytes / (stream.BytesPerFrame() * channels)

		if frameCountMin > fillCount {
			// Ring buffer does not have enough data, fill with zeroes.
			frameCount := frameCountMin
			if frameCount <= 0 {
				return
			}

			areas, err := stream.BeginWrite(&frameCount)
			if err != nil {
				log.Printf("begin write error: %s", err)
				cancelParent()
				return
			}
			if frameCount <= 0 {
				return
			}
			zeroArray := make([]byte, 64)
			for frame := 0; frame < frameCount; frame++ {
				for ch := 0; ch < channelCount; ch++ {
					buffer := areas.Buffer(ch, frame)
					copy(buffer, zeroArray)
				}
			}
			_ = stream.EndWrite()
		}

		readCount := fillCount
		if readCount > frameCountMax {
			readCount = frameCountMax
		}

		frameLeft := readCount
		for {
			frameCount := frameLeft
			if frameCount <= 0 {
				break
			}

			areas, err := stream.BeginWrite(&frameCount)
			if err != nil {
				log.Printf("begin write error: %s", err)
				cancelParent()
				return
			}
			if frameCount <= 0 {
				break
			}
			for frame := 0; frame < frameCount; frame++ {
				for ch := 0; ch < channelCount; ch++ {
					buffer := areas.Buffer(ch, frame)
					_, err = ringBuffer.Read(buffer)
					if err != nil {
						//	log.Printf("ringbuffer read error: %s", err)
					}
				}
			}
			err = stream.EndWrite()
			if err != nil {
				log.Printf("end write error: %s", err)
				cancelParent()
				return
			}

			frameLeft -= frameCount
		}
	})
	outstream.SetUnderflowCallback(func(stream *soundio.OutStream) {
		underflowCount++
		log.Printf("underflow %d", underflowCount)
	})
	err = outstream.Open()
	if err != nil {
		return fmt.Errorf("unable to open output device: %s", err)
	}

	capacity := channels * int(0.2*float64(instream.SampleRate()*instream.BytesPerFrame()))
	log.Printf("capacity %d", capacity)
	ringBuffer = rbuf.NewFixedSizeRingBuf(capacity)
	_, _ = ringBuffer.Write(make([]byte, capacity))

	err = instream.Start()
	if err != nil {
		return fmt.Errorf("unable to start input device: %s", err)
	}
	err = outstream.Start()
	if err != nil {
		return fmt.Errorf("unable to start output device: %s", err)
	}

	log.Println("Type CTRL+C to quit by killing process...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				s.WaitEvents()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.Wakeup()
			return ctx.Err()
		}
	}
}

func signalContext(ctx context.Context) context.Context {
	parent, cancelParent := context.WithCancel(ctx)
	go func() {
		defer cancelParent()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)
		defer signal.Stop(sig)

		select {
		case <-parent.Done():
			log.Println("Cancel from parent")
			return
		case s := <-sig:
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("Stop!")
				return
			}
		}
	}()

	return parent
}
