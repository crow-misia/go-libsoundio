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
	"math"
	"os"
	"os/signal"
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
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	parentCtx.Done()

	os.Exit(exitCode)
}

func realMain(ctx context.Context) error {
	backends := []soundio.Backend{
		soundio.BackendJack, soundio.BackendPulseAudio, soundio.BackendAlsa,
		soundio.BackendCoreAudio, soundio.BackendWasapi, soundio.BackendDummy,
	}

	fmt.Printf("libsound Version = %s\n", soundio.Version())
	fmt.Printf("Front Center Channel Name = %s\n", soundio.ChannelID(soundio.ChannelIDFrontCenter))
	fmt.Printf("Front Left Channel ID = %d\n", soundio.ParseChannelID("Front Left"))
	fmt.Printf("front-right Channel ID = %d\n", soundio.ParseChannelID("front-right"))
	fmt.Printf("Channel Max Count = %d\n", soundio.MaxChannels)
	fmt.Printf("Channel Layout Builtin Count = %d\n", soundio.ChannelLayoutBuiltinCount())

	s := soundio.Create()
	defer s.Destroy()

	fmt.Printf("App Name = %s\n", s.GetAppName())
	s.SetAppName("FugaHoge")
	fmt.Printf("Changed App Name = %s\n", s.GetAppName())

	backendCount := s.BackendCount()
	fmt.Printf("Backend Count = %d\n", backendCount)
	for i := 0; i < backendCount; i++ {
		backend := s.GetBackend(i)
		fmt.Printf("Backend Index(%d) = %d (%s)\n", i, backend, backend)
	}

	err := s.Connect()
	if err != nil {
		return fmt.Errorf("error connecting: %s", err)
	}

	fmt.Printf("Current Backend Name = %s\n", s.GetCurrentBackend())
	for _, b := range backends {
		fmt.Printf("Have %s = %t\n", b, b.Have())
	}

	s.FlushEvents()

	defaultInputDeviceIndex := s.DefaultInputDeviceIndex()
	fmt.Printf("Default Input Device Index = %d\n", defaultInputDeviceIndex)

	defaultOutputDeviceIndex := s.DefaultOutputDeviceIndex()
	fmt.Printf("Default Output Device Index = %d\n", defaultOutputDeviceIndex)

	device := s.GetOutputDevice(defaultOutputDeviceIndex)
	defer device.RemoveReference()
	fmt.Println("Device")
	fmt.Printf("  ID = %s\n", device.GetID())
	fmt.Printf("  Aim = %s\n", device.GetAim())
	fmt.Printf("  Name = %s\n", device.GetName())
	fmt.Printf("  Layouts Count = %d\n", device.GetLayoutCount())
	for _, l := range device.GetLayouts() {
		fmt.Printf("    %s\n", l.GetName())
	}
	fmt.Printf("  Formats Count = %d\n", device.GetFormatCount())
	for _, f := range device.GetFormats() {
		fmt.Printf("    %s\n", f)
	}
	fmt.Println("  SampleRates")
	for _, sampleRate := range device.GetSampleRates() {
		fmt.Printf("    %d ... %d\n", sampleRate.GetMin(), sampleRate.GetMax())
	}
	fmt.Printf("  SampleRate Count = %d\n", device.GetSampleRateCount())
	fmt.Printf("  SampleRate Current = %d\n", device.GetSampleRateCurrent())
	fmt.Printf("  Software Latency Min = %f\n", device.GetSoftwareLatencyMin())
	fmt.Printf("  Software Latency Min = %f\n", device.GetSoftwareLatencyMax())
	fmt.Printf("  Software Latency Current = %f\n", device.GetSoftwareLatencyCurrent())
	fmt.Printf("  Is Raw = %t\n", device.IsRaw())
	fmt.Printf("  Ref Count = %d\n", device.GetRefCount())
	fmt.Printf("  Probe Error = %s\n", device.GetProbeError())

	outStream := device.OutStreamCreate()
	outStream.SetFormat(soundio.FormatFloat32LE)
	fmt.Println("  OutStream")
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
				panic(fmt.Sprintf("%s\n", err))
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
					area := areas.GetArea(channel)
					buffer := area.GetBuffer()
					bites := math.Float32bits(sample)
					step := area.GetStep()
					binary.LittleEndian.PutUint32(buffer[step*frame:], bites)
				}
			}

			secondsOffset = math.Mod(secondsOffset+secondsPerFrame*float64(frameCount), 1.0)

			err = outStream.EndWrite()
			if err != nil {
				panic(fmt.Sprintf("%s\n", err))
			}
			framesLeft -= frameCount
		}
	})

	layout := device.GetCurrentLayout()
	outStream.SetLayout(layout)
	fmt.Printf("    Layout Error = %s\n", outStream.GetLayoutError())
	fmt.Printf("    Layout Name = %s\n", layout.GetName())
	fmt.Printf("    Layout Detect Builtin = %t\n", layout.DetectBuiltin())

	err = outStream.Open()
	if err != nil {
		return fmt.Errorf("error opening: %s", err)
	}
	defer outStream.Destroy()

	fmt.Printf("    Name = %s\n", outStream.GetName())
	fmt.Printf("    BytePerFrame = %d\n", outStream.GetBytesPerSample())
	fmt.Printf("    BytePerSample = %d\n", outStream.GetBytesPerSample())
	fmt.Printf("    SoftwareLatency = %f\n", outStream.GetSoftwareLatency())
	fmt.Printf("    SampleRate = %d\n", outStream.GetSampleRate())
	fmt.Printf("    Format = %s\n", outStream.GetFormat())
	fmt.Printf("    Volume = %f\n", outStream.GetVolume())
	fmt.Printf("    NonTerminalHint = %t\n", outStream.GetNonTerminalHint())
	fmt.Printf("    Channel Count = %d\n", layout.GetChannelCount())
	fmt.Println("    Channels")
	channels := layout.GetChannels()
	channelCount := layout.GetChannelCount()
	for i := 0; i < channelCount; i++ {
		fmt.Printf("      Channel ID = %d, Name = %s\n", channels[i], channels[i])
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
			_, _ = fmt.Fprintln(os.Stderr, "Cancel from parent")
			return
		case s := <-sig:
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				_, _ = fmt.Fprintln(os.Stderr, "Stop!")
				return
			}
		}
	}()

	return parent
}
