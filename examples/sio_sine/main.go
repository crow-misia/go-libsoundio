package main

import (
	"fmt"
	"github.com/crow-misia/go-libsoundio"
	"math"
	"os"
	"unsafe"
)

var exitCode = 0
var secondsOffset = 0.0

const PI = 3.1415926535

func main() {
	realMain()
	os.Exit(exitCode)
}

func realMain() {
	backends := []soundio.Backend{
		soundio.BackendJack, soundio.BackendPulseAudio, soundio.BackendAlsa,
		soundio.BackendCoreAudio, soundio.BackendWasapi, soundio.BackendDummy,
	}

	fmt.Printf("libsound Version = %s\n", soundio.Version())
	fmt.Printf("Front Center Channel Name = %s\n", soundio.ChannelID(soundio.ChannelIDFrontCenter))
	fmt.Printf("Front Left Channel ID = %d\n", soundio.ParseChannelId("Front Left"))
	fmt.Printf("front-right Channel ID = %d\n", soundio.ParseChannelId("front-right"))
	fmt.Printf("Channel Max Count = %d\n", soundio.MaxChannels)
	fmt.Printf("Channel Layout Builtin Count = %d\n", soundio.ChannelLayoutBuiltinCount())

	s := soundio.Create()
	defer s.Destroy()

	backendCount := s.BackendCount()
	fmt.Printf("Backend Count = %d\n", backendCount)
	for i := 0; i < backendCount; i++ {
		backend := s.GetBackend(i)
		fmt.Printf("Backend Index(%d) = %d (%s)\n", i, backend, backend)
	}

	err := s.Connect()
	if err != nil {
		fmt.Printf("error connecting: %s\n", err)
		exitCode = 1
		return
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
	sampleRateRange := device.GetSampleRates()
	fmt.Printf("  SampleRate = %d ... %d\n", sampleRateRange.GetMin(), sampleRateRange.GetMax())
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

			for frame := 0; frame < frameCount; frame++ {
				sample := float32(math.Sin((secondsOffset + float64(frame)*secondsPerFrame) * radiansPerSecond))

				for channel := 0; channel < layout.GetChannelCount(); channel++ {
					area := areas.GetArea(channel)
					ptr := (*float32)(unsafe.Pointer(area.GetBuffer() + uintptr(area.GetStep()*frame)))
					*ptr = sample
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
		fmt.Printf("error opening: %s\n", err)
		exitCode = 1
		return
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
		fmt.Printf("      Channel ID = %d, Name = %s\n", (*channels)[i], (*channels)[i])
	}

	err = outStream.Start()
	if err != nil {
		fmt.Printf("error opening: %s\n", err)
		exitCode = 1
		return
	}

	for {
		s.WaitEvents()
	}
}
