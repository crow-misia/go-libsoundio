/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/crow-misia/go-libsoundio"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var exitCode = 0
var secondsOffset = 0.0

const PI = 3.1415926535

func main() {
	ctx := context.Background()
	parentCtx := signalContext(ctx)
	err := realMain(parentCtx)
	if err != nil {
		exitCode = 1
		log.Println(err)
	}
	parentCtx.Done()

	os.Exit(exitCode)
}

func realMain(ctx context.Context) error {
	_, cancelParent := context.WithCancel(ctx)
	defer cancelParent()

	backends := []soundio.Backend{
		soundio.BackendJack, soundio.BackendPulseAudio, soundio.BackendAlsa,
		soundio.BackendCoreAudio, soundio.BackendWasapi, soundio.BackendDummy,
	}

	log.Printf("libsound Version = %s", soundio.Version())
	log.Printf("Front Center Channel Name = %s", soundio.ChannelID(soundio.ChannelIDFrontCenter))
	log.Printf("Front Left Channel ID = %d", soundio.ParseChannelID("Front Left"))
	log.Printf("front-right Channel ID = %d", soundio.ParseChannelID("front-right"))
	log.Printf("Channel Max Count = %d", soundio.MaxChannels)
	log.Printf("Channel Layout Builtin Count = %d", soundio.ChannelLayoutBuiltinCount())

	s := soundio.Create()
	defer s.Destroy()

	log.Printf("App Name = %s", s.GetAppName())
	s.SetAppName("FugaHoge")
	log.Printf("Changed App Name = %s", s.GetAppName())

	backendCount := s.BackendCount()
	log.Printf("Backend Count = %d", backendCount)
	for i := 0; i < backendCount; i++ {
		backend := s.GetBackend(i)
		log.Printf("Backend Index(%d) = %d (%s)", i, backend, backend)
	}

	err := s.Connect()
	if err != nil {
		return fmt.Errorf("error connecting: %s", err)
	}

	log.Printf("Current Backend Name = %s", s.GetCurrentBackend())
	for _, b := range backends {
		log.Printf("Have %s = %t", b, b.Have())
	}

	s.FlushEvents()

	defaultInputDeviceIndex := s.DefaultInputDeviceIndex()
	log.Printf("Default Input Device Index = %d", defaultInputDeviceIndex)

	defaultOutputDeviceIndex := s.DefaultOutputDeviceIndex()
	log.Printf("Default Output Device Index = %d", defaultOutputDeviceIndex)

	device := s.GetOutputDevice(defaultOutputDeviceIndex)
	defer device.RemoveReference()
	log.Println("Device")
	log.Printf("  ID = %s", device.GetID())
	log.Printf("  Aim = %s", device.GetAim())
	log.Printf("  Name = %s", device.GetName())
	log.Printf("  Layouts Count = %d", device.GetLayoutCount())
	for _, l := range device.GetLayouts() {
		log.Printf("    %s", l.GetName())
	}
	log.Printf("  Formats Count = %d", device.GetFormatCount())
	formats := make([]string, device.GetFormatCount())
	for i, format := range device.GetFormats() {
		formats[i] = fmt.Sprint(format)
	}
	log.Printf("  formats: %s", strings.Join(formats, ", "))
	log.Println("  SampleRates")
	for _, sampleRate := range device.GetSampleRates() {
		log.Printf("    %d - %d", sampleRate.GetMin(), sampleRate.GetMax())
	}
	log.Printf("  SampleRate Count = %d", device.GetSampleRateCount())
	log.Printf("  SampleRate Current = %d", device.GetSampleRateCurrent())
	log.Printf("  Software Latency Min = %f", device.GetSoftwareLatencyMin())
	log.Printf("  Software Latency Min = %f", device.GetSoftwareLatencyMax())
	log.Printf("  Software Latency Current = %f", device.GetSoftwareLatencyCurrent())
	log.Printf("  Is Raw = %t", device.IsRaw())
	log.Printf("  Ref Count = %d", device.GetRefCount())
	log.Printf("  Probe Error = %s", device.GetProbeError())

	outStream := device.OutStreamCreate()
	outStream.SetFormat(soundio.FormatFloat32LE)
	log.Println("  OutStream")
	outStream.SetWriteCallback(func(stream *soundio.OutStream, frameCountMix int, frameCountMax int) {
		layout := stream.GetLayout()
		sampleRate := float64(stream.GetSampleRate())
		secondsPerFrame := 1.0 / sampleRate
		var areas *soundio.ChannelAreas

		framesLeft := frameCountMax
		for framesLeft > 0 {
			frameCount := framesLeft
			areas, err = stream.BeginWrite(&frameCount)
			if err != nil {
				log.Println(err)
				cancelParent()
			}

			if frameCount <= 0 {
				break
			}

			pitch := 440.0
			radiansPerSecond := pitch * 2.0 * PI
			channelCount := layout.GetChannelCount()

			for frame := 0; frame < frameCount; frame++ {
				sample := float32(math.Sin((secondsOffset + float64(frame)*secondsPerFrame) * radiansPerSecond))

				for channel := 0; channel < channelCount; channel++ {
					buffer := areas.GetBuffer(channel, frame)
					bites := math.Float32bits(sample)
					binary.LittleEndian.PutUint32(buffer, bites)
				}
			}

			secondsOffset = math.Mod(secondsOffset+secondsPerFrame*float64(frameCount), 1.0)

			err = outStream.EndWrite()
			if err != nil {
				log.Println(err)
				cancelParent()
			}
			framesLeft -= frameCount
		}
	})

	layout := device.GetCurrentLayout()
	outStream.SetLayout(layout)
	log.Printf("    Layout Error = %s", outStream.GetLayoutError())
	log.Printf("    Layout Name = %s", layout.GetName())
	log.Printf("    Layout Detect Builtin = %t", layout.DetectBuiltin())

	err = outStream.Open()
	if err != nil {
		return fmt.Errorf("error opening: %s", err)
	}
	defer outStream.Destroy()

	log.Printf("    Name = %s", outStream.GetName())
	log.Printf("    BytePerFrame = %d", outStream.GetBytesPerSample())
	log.Printf("    BytePerSample = %d", outStream.GetBytesPerSample())
	log.Printf("    SoftwareLatency = %f", outStream.GetSoftwareLatency())
	log.Printf("    SampleRate = %d", outStream.GetSampleRate())
	log.Printf("    Format = %s", outStream.GetFormat())
	log.Printf("    Volume = %f", outStream.GetVolume())
	log.Printf("    NonTerminalHint = %t", outStream.GetNonTerminalHint())
	log.Printf("    Channel Count = %d", layout.GetChannelCount())
	log.Println("    Channels")
	channels := layout.GetChannels()
	channelCount := layout.GetChannelCount()
	for i := 0; i < channelCount; i++ {
		log.Printf("      Channel ID = %d, Name = %s", channels[i], channels[i])
	}

	err = outStream.Start()
	if err != nil {
		return fmt.Errorf("error opening: %s", err)
	}

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
