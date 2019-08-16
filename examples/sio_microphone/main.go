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
		_, _ = fmt.Fprintln(os.Stderr, err)
		exitCode = 1
	} else {
		ctx := context.Background()
		parentCtx := signalContext(ctx)
		err := realMain(parentCtx, enumBackend, inputDeviceId, inputIsRaw, outputDeviceId, outputIsRaw, latencySec)
		if err != nil {
			exitCode = 1
			_, _ = fmt.Fprintln(os.Stderr, err)
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
			if device.IsRaw() == isRaw && deviceId == device.GetID() {
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

	s := soundio.Create()
	defer s.Destroy()

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
		return io.GetInputDevice(index)
	})
	if err != nil {
		return err
	}
	defer selectedInputDevice.RemoveReference()
	_, _ = fmt.Fprintf(os.Stderr, "Input device: %s\n", selectedInputDevice.GetName())
	if selectedInputDevice.GetProbeError() != nil {
		return fmt.Errorf("unable to probe device: %s", selectedInputDevice.GetProbeError())
	}

	selectedOutputDevice, err := selectDevice(s, outputDeviceId, outputIsRaw, func(io *soundio.SoundIo) int {
		return io.OutputDeviceCount()
	}, func(io *soundio.SoundIo) int {
		return io.DefaultOutputDeviceIndex()
	}, func(io *soundio.SoundIo, index int) *soundio.Device {
		return io.GetOutputDevice(index)
	})
	if err != nil {
		return err
	}
	defer selectedOutputDevice.RemoveReference()
	_, _ = fmt.Fprintf(os.Stderr, "Output device: %s\n", selectedOutputDevice.GetName())
	if selectedOutputDevice.GetProbeError() != nil {
		return fmt.Errorf("unable to probe device: %s", selectedOutputDevice.GetProbeError())
	}

	selectedOutputDevice.SortChannelLayouts()
	layout := soundio.BestMatchingLayout(selectedOutputDevice, selectedInputDevice)
	if layout == nil {
		return fmt.Errorf("channel layouts not compatible")
	}

	sampleRate := 0
	for _, rate := range prioritizedSampleRates {
		if selectedInputDevice.SupportsSampleRate(rate) && selectedOutputDevice.SupportsSampleRate(rate) {
			sampleRate = rate
			break
		}
	}
	if sampleRate == 0 {
		return fmt.Errorf("incompatible sample rates")
	}
	_, _ = fmt.Fprintf(os.Stderr, "Sample rate: %d\n", sampleRate)

	format := soundio.FormatInvalid
	for _, f := range prioritizedFormats {
		if selectedInputDevice.SupportsFormat(f) && selectedOutputDevice.SupportsFormat(f) {
			format = f
			break
		}
	}
	if format == soundio.FormatInvalid {
		return fmt.Errorf("incompatible sample formats")
	}
	_, _ = fmt.Fprintf(os.Stderr, "Format: %s\n", format)

	var ringBuffer *rbuf.FixedSizeRingBuf

	instream := selectedInputDevice.InStreamCreate()
	defer instream.Destroy()
	instream.SetFormat(format)
	instream.SetLayout(layout)
	instream.SetSampleRate(sampleRate)
	instream.SetSoftwareLatency(latencySec)
	instream.SetReadCallback(func(stream *soundio.InStream, frameCountMin int, frameCountMax int) {
		freeCount := ringBuffer.N - ringBuffer.Readable
		writeFrames := freeCount
		if writeFrames > frameCountMax {
			writeFrames = frameCountMax
		}

		channelCount := stream.GetLayout().GetChannelCount()
		frameLeft := writeFrames

		for {
			frameCount := frameLeft
			areas, err := stream.BeginRead(&frameCount)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "begin read error: %s", err)
				cancelParent()
				return
			}
			if frameCount <= 0 {
				break
			}
			if areas == nil {
				_, _ = ringBuffer.Write(make([]byte, frameCount*instream.GetBytesPerFrame()))
			} else {
				for frame := 0; frame < frameCount; frame++ {
					for ch := 0; ch < channelCount; ch++ {
						area := areas.GetArea(ch)
						step := area.GetStep()
						offset := frame * step
						_, _ = ringBuffer.Write(area.GetBuffer()[offset : offset+step])
					}
				}
			}
			_ = stream.EndRead()

			frameLeft -= frameCount
			if frameLeft <= 0 {
				break
			}
		}
	})
	instream.SetOverflowCallback(func(stream *soundio.InStream) {
		overflowCount++
		_, _ = fmt.Fprintf(os.Stderr, "overflow %d\n", overflowCount)
	})
	err = instream.Open()
	if err != nil {
		return fmt.Errorf("unable to open input device: %s", err)
	}

	outstream := selectedOutputDevice.OutStreamCreate()
	defer outstream.Destroy()
	fmt.Printf("Format: %s\n", format)
	fmt.Printf("layout: %s\n", layout.GetName())
	fmt.Printf("layout: %d\n", layout.GetChannelCount())
	fmt.Printf("sampleRate: %d\n", sampleRate)
	fmt.Printf("latencySec: %f\n", latencySec)
	outstream.SetFormat(format)
	outstream.SetLayout(layout)
	outstream.SetSampleRate(sampleRate)
	outstream.SetSoftwareLatency(latencySec)
	outstream.SetWriteCallback(func(stream *soundio.OutStream, frameCountMin int, frameCountMax int) {
		fillbytes := ringBuffer.Readable
		fillCount := fillbytes / stream.GetBytesPerFrame()

		channelCount := stream.GetLayout().GetChannelCount()

		if frameCountMin > fillCount {
			// Ring buffer does not have enough data, fill with zeroes.
			frameCount := frameCountMin

			areas, err := stream.BeginWrite(&frameCount)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "begin write error: %s", err)
				cancelParent()
				return
			}
			if frameCount <= 0 {
				return
			}
			zeroArray := make([]byte, 16)
			for frame := 0; frame < frameCount; frame++ {
				for ch := 0; ch < channelCount; ch++ {
					area := areas.GetArea(ch)
					step := area.GetStep()
					offset := frame * step
					copy(area.GetBuffer()[offset:offset+step], zeroArray[:step])
				}
			}
			_ = stream.EndWrite()
		}

		readCount := frameCountMax
		if readCount > fillCount {
			readCount = fillCount
		}

		frameLeft := readCount
		for {
			frameCount := frameLeft
			if frameCount <= 0 {
				return
			}

			areas, err := stream.BeginWrite(&frameCount)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "begin write error: %s", err)
				cancelParent()
				return
			}
			if frameCount <= 0 {
				break
			}
			for frame := 0; frame < frameCount; frame++ {
				for ch := 0; ch < channelCount; ch++ {
					area := areas.GetArea(ch)
					step := area.GetStep()
					offset := frame * step
					_, _ = ringBuffer.Read(area.GetBuffer()[offset : offset+step])
				}
			}
			_ = stream.EndWrite()

			frameLeft -= frameCount
			if frameLeft <= 0 {
				break
			}
		}
	})
	outstream.SetUnderflowCallback(func(stream *soundio.OutStream) {
		underflowCount++
		_, _ = fmt.Fprintf(os.Stderr, "underflow %d\n", overflowCount)
	})
	err = outstream.Open()
	if err != nil {
		return fmt.Errorf("unable to open output device: %s", err)
	}

	capacity := int(latencySec * 2 * float64(instream.GetSampleRate()*instream.GetBytesPerFrame()))
	ringBuffer = rbuf.NewFixedSizeRingBuf(capacity)

	err = instream.Start()
	if err != nil {
		return fmt.Errorf("unable to start input device: %s", err)
	}
	err = outstream.Start()
	if err != nil {
		return fmt.Errorf("unable to start output device: %s", err)
	}

	_, _ = fmt.Fprintln(os.Stderr, "Type CTRL+C to quit by killing process...")

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
			fmt.Println("Cancel from parent")
			return
		case s := <-sig:
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				fmt.Println("Stop!")
				return
			}
		}
	}()

	return parent
}
