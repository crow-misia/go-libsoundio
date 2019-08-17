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
	"time"
)

const ringBufferDurationSeconds = 30

var overflowCount = 0
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
		deviceId string
		backend  string
		isRaw    bool
		outfile  string
	)
	flag.NewFlagSet("help", flag.ExitOnError)
	flag.StringVar(&deviceId, "device", "", "id")
	flag.StringVar(&backend, "backend", "", "dummy|alsa|pulseaudio|jack|coreaudio|wasapi")
	flag.BoolVar(&isRaw, "raw", false, "raw")
	flag.StringVar(&outfile, "file", "", "filename")
	flag.Parse()

	enumBackend, err := parseBackend(backend)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		exitCode = 1
	} else if len(outfile) == 0 {
		flag.PrintDefaults()
		exitCode = 1
	} else {
		ctx := context.Background()
		parentCtx := signalContext(ctx)
		err := realMain(parentCtx, enumBackend, deviceId, isRaw, outfile)
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

func realMain(ctx context.Context, backend soundio.Backend, deviceId string, isRaw bool, outfile string) error {
	_, cancelParent := context.WithCancel(ctx)
	defer cancelParent()

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

	selectedDevice, err := selectDevice(s, deviceId, isRaw, func(io *soundio.SoundIo) int {
		return io.InputDeviceCount()
	}, func(io *soundio.SoundIo) int {
		return io.DefaultInputDeviceIndex()
	}, func(io *soundio.SoundIo, index int) *soundio.Device {
		return io.GetInputDevice(index)
	})
	if err != nil {
		return err
	}
	defer selectedDevice.RemoveReference()

	_, _ = fmt.Fprintf(os.Stderr, "Device: %s\n", selectedDevice.GetName())

	if selectedDevice.GetProbeError() != nil {
		return fmt.Errorf("unable to probe device: %s", selectedDevice.GetProbeError())
	}

	selectedDevice.SortChannelLayouts()

	sampleRate := 0
	for _, rate := range prioritizedSampleRates {
		if selectedDevice.SupportsSampleRate(rate) {
			sampleRate = rate
			break
		}
	}
	if sampleRate == 0 {
		sampleRate = selectedDevice.GetSampleRates()[0].GetMax()
	}
	_, _ = fmt.Fprintf(os.Stderr, "Sample rate: %d\n", sampleRate)

	format := soundio.FormatInvalid
	for _, f := range prioritizedFormats {
		if selectedDevice.SupportsFormat(f) {
			format = f
			break
		}
	}
	if format == soundio.FormatInvalid {
		format = selectedDevice.GetFormats()[0]
	}
	_, _ = fmt.Fprintf(os.Stderr, "Format: %s\n", format)

	file, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("unable to open %s: %s", outfile, err)
	}
	defer file.Close()

	var ringBuffer *rbuf.FixedSizeRingBuf
	instream := selectedDevice.InStreamCreate()
	defer instream.Destroy()
	instream.SetFormat(format)
	instream.SetSampleRate(sampleRate)
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
				_, _ = ringBuffer.Write(make([]byte, frameCount*channelCount*stream.GetBytesPerFrame()))
			} else {
				for frame := 0; frame < frameCount; frame++ {
					for ch := 0; ch < channelCount; ch++ {
						buffer := areas.GetBuffer(ch, frame)
						_, _ = ringBuffer.Write(buffer)
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

	capacity := ringBufferDurationSeconds * instream.GetSampleRate() * instream.GetBytesPerFrame()
	ringBuffer = rbuf.NewFixedSizeRingBuf(capacity)

	err = instream.Start()
	if err != nil {
		return fmt.Errorf("unable to start input device: %s", err)
	}

	_, _ = fmt.Fprintln(os.Stderr, "Type CTRL+C to quit by killing process...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.FlushEvents()
			time.Sleep(1 * time.Second)
			_, err := ringBuffer.WriteTo(file)

			if err != nil {
				return fmt.Errorf("write error: %s", err)
			}
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
