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
		log.Println(err)
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

func realMain(ctx context.Context, backend soundio.Backend, deviceId string, isRaw bool, outfile string) error {
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

	selectedDevice, err := selectDevice(s, deviceId, isRaw, func(io *soundio.SoundIo) int {
		return io.InputDeviceCount()
	}, func(io *soundio.SoundIo) int {
		return io.DefaultInputDeviceIndex()
	}, func(io *soundio.SoundIo, index int) *soundio.Device {
		return io.InputDevice(index)
	})
	if err != nil {
		return err
	}
	defer selectedDevice.RemoveReference()

	log.Printf("Device: %s", selectedDevice.Name())

	if selectedDevice.ProbeError() != nil {
		return fmt.Errorf("unable to probe device: %s", selectedDevice.ProbeError())
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
		sampleRate = selectedDevice.SampleRates()[0].Max()
	}
	log.Printf("Sample rate: %d", sampleRate)

	format := soundio.FormatInvalid
	for _, f := range prioritizedFormats {
		if selectedDevice.SupportsFormat(f) {
			format = f
			break
		}
	}
	if format == soundio.FormatInvalid {
		format = selectedDevice.Formats()[0]
	}
	log.Printf("Format: %s", format)

	file, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("unable to open %s: %s", outfile, err)
	}
	defer file.Close()

	var ringBuffer *rbuf.FixedSizeRingBuf
	instream := selectedDevice.NewInStream()
	defer instream.Destroy()
	instream.SetFormat(format)
	instream.SetSampleRate(sampleRate)
	frameBytes := instream.Layout().ChannelCount() * instream.BytesPerFrame()
	instream.SetReadCallback(func(stream *soundio.InStream, frameCountMin int, frameCountMax int) {
		freeBytes := ringBuffer.N - ringBuffer.Readable
		freeCount := freeBytes / frameBytes
		writeFrames := freeCount
		if writeFrames > frameCountMax {
			writeFrames = frameCountMax
		}

		channelCount := stream.Layout().ChannelCount()
		frameLeft := writeFrames

		for {
			frameCount := frameLeft
			if frameCount <= 0 {
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
			if areas == nil {
				_, _ = ringBuffer.Write(make([]byte, frameCount*channelCount*stream.BytesPerFrame()))
			} else {
				for frame := 0; frame < frameCount; frame++ {
					for ch := 0; ch < channelCount; ch++ {
						buffer := areas.Buffer(ch, frame)
						_, err = ringBuffer.Write(buffer)
						if err != nil {
							log.Printf("ringbuffer write error: %s, len %d", err, len(buffer))
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

	capacity := instream.Layout().ChannelCount() * ringBufferDurationSeconds * 5 * instream.SampleRate() * instream.BytesPerFrame()
	ringBuffer = rbuf.NewFixedSizeRingBuf(capacity)

	err = instream.Start()
	if err != nil {
		return fmt.Errorf("unable to start input device: %s", err)
	}

	log.Println("Type CTRL+C to quit by killing process...")

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
