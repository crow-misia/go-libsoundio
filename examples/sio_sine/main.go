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
	log.Printf("Front Center Channel Name = %s", soundio.ChannelIDFrontCenter)
	log.Printf("Front Left Channel ID = %d", soundio.ParseChannelID("Front Left"))
	log.Printf("front-right Channel ID = %d", soundio.ParseChannelID("front-right"))
	log.Printf("Channel Max Count = %d", soundio.MaxChannels)
	log.Printf("Channel Layout Builtin Count = %d", soundio.ChannelLayoutBuiltinCount())

	s := soundio.Create()
	defer s.Destroy()

	log.Printf("App Name = %s", s.AppName())
	s.SetAppName("FugaHoge")
	log.Printf("Changed App Name = %s", s.AppName())

	backendCount := s.BackendCount()
	log.Printf("Backend Count = %d", backendCount)
	for i := 0; i < backendCount; i++ {
		backend := s.Backend(i)
		log.Printf("Backend Index(%d) = %d (%s)", i, backend, backend)
	}

	err := s.Connect()
	if err != nil {
		return fmt.Errorf("error connecting: %s", err)
	}

	log.Printf("Current Backend Name = %s", s.CurrentBackend())
	for _, b := range backends {
		log.Printf("Have %s = %t", b, b.Have())
	}

	s.FlushEvents()

	defaultInputDeviceIndex := s.DefaultInputDeviceIndex()
	log.Printf("Default Input Device Index = %d", defaultInputDeviceIndex)

	defaultOutputDeviceIndex := s.DefaultOutputDeviceIndex()
	log.Printf("Default Output Device Index = %d", defaultOutputDeviceIndex)

	device := s.OutputDevice(defaultOutputDeviceIndex)
	defer device.RemoveReference()
	log.Println("Device")
	log.Printf("  ID = %s", device.ID())
	log.Printf("  Aim = %s", device.Aim())
	log.Printf("  Name = %s", device.Name())
	log.Printf("  Layouts Count = %d", device.LayoutCount())
	for _, l := range device.Layouts() {
		log.Printf("    %s", l.Name())
	}
	log.Printf("  Formats Count = %d", device.FormatCount())
	formats := make([]string, device.FormatCount())
	for i, format := range device.Formats() {
		formats[i] = fmt.Sprint(format)
	}
	log.Printf("  formats: %s", strings.Join(formats, ", "))
	log.Println("  SampleRates")
	for _, sampleRate := range device.SampleRates() {
		log.Printf("    %d - %d", sampleRate.Min(), sampleRate.Max())
	}
	log.Printf("  SampleRate Count = %d", device.SampleRateCount())
	log.Printf("  SampleRate Current = %d", device.SampleRateCurrent())
	log.Printf("  Software Latency Min = %f", device.SoftwareLatencyMin())
	log.Printf("  Software Latency Min = %f", device.SoftwareLatencyMax())
	log.Printf("  Software Latency Current = %f", device.SoftwareLatencyCurrent())
	log.Printf("  Raw = %t", device.Raw())
	log.Printf("  Ref Count = %d", device.RefCount())
	log.Printf("  Probe Error = %s", device.ProbeError())

	outStream := device.NewOutStream()
	outStream.SetFormat(soundio.FormatFloat32LE)
	log.Println("  OutStream")
	outStream.SetWriteCallback(func(stream *soundio.OutStream, frameCountMix int, frameCountMax int) {
		layout := stream.Layout()
		sampleRate := float64(stream.SampleRate())
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
			channelCount := layout.ChannelCount()

			for frame := 0; frame < frameCount; frame++ {
				sample := float32(math.Sin((secondsOffset + float64(frame)*secondsPerFrame) * radiansPerSecond))

				for channel := 0; channel < channelCount; channel++ {
					buffer := areas.Buffer(channel, frame)
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

	layout := device.CurrentLayout()
	outStream.SetLayout(layout)
	log.Printf("    Layout Error = %s", outStream.LayoutError())
	log.Printf("    Layout Name = %s", layout.Name())
	log.Printf("    Layout Detect Builtin = %t", layout.DetectBuiltin())

	err = outStream.Open()
	if err != nil {
		return fmt.Errorf("error opening: %s", err)
	}
	defer outStream.Destroy()

	log.Printf("    Name = %s", outStream.Name())
	log.Printf("    BytePerFrame = %d", outStream.BytesPerSample())
	log.Printf("    BytePerSample = %d", outStream.BytesPerSample())
	log.Printf("    SoftwareLatency = %f", outStream.SoftwareLatency())
	log.Printf("    SampleRate = %d", outStream.SampleRate())
	log.Printf("    Format = %s", outStream.Format())
	log.Printf("    Volume = %f", outStream.Volume())
	log.Printf("    NonTerminalHint = %t", outStream.NonTerminalHint())
	log.Printf("    Channel Count = %d", layout.ChannelCount())
	log.Println("    Channels")
	channels := layout.Channels()
	channelCount := layout.ChannelCount()
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
