/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */
package main

import (
	"context"
	"flag"
	"fmt"
	soundio "github.com/crow-misia/go-libsoundio"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var exitCode = 0

func main() {
	var (
		watchEvents bool
		backend     string
		shortOutput bool
	)
	flag.NewFlagSet("help", flag.ExitOnError)
	flag.BoolVar(&watchEvents, "watch", false, "watch")
	flag.StringVar(&backend, "backend", "", "dummy|alsa|pulseaudio|jack|coreaudio|wasapi")
	flag.BoolVar(&shortOutput, "short", false, "short")
	flag.Parse()

	enumBackend, err := parseBackend(backend)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	} else {
		ctx := context.Background()
		parentCtx := signalContext(ctx)
		err := realMain(parentCtx, enumBackend, watchEvents, shortOutput)
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

func printChannelLayout(layout *soundio.ChannelLayout) {
	name := layout.GetName()
	if len(name) == 0 {
		for i, channel := range layout.GetChannels() {
			if i == 0 {
				_, _ = fmt.Fprintf(os.Stderr, "%s", channel)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, ", %s", channel)
			}
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, name)
	}
}

func printDevice(device *soundio.Device, shortOutput bool, isDefault bool) {
	defaultStr := ""
	if isDefault {
		defaultStr = " (default)"
	}

	rawStr := ""
	if device.IsRaw() {
		rawStr = " (raw)"
	}

	_, _ = fmt.Fprintf(os.Stderr, "%s%s%s\n", device.GetName(), defaultStr, rawStr)
	if shortOutput {
		return
	}

	_, _ = fmt.Fprintf(os.Stderr, "  id: %s\n", device.GetID())

	if device.GetProbeError() == nil {
		_, _ = fmt.Fprintln(os.Stderr, "  channel layouts:")
		for _, layout := range device.GetLayouts() {
			_, _ = fmt.Fprint(os.Stderr, "    ")
			printChannelLayout(layout)
			_, _ = fmt.Fprintln(os.Stderr)
		}
		if device.GetCurrentLayout().GetChannelCount() > 0 {
			_, _ = fmt.Fprint(os.Stderr, "  current layout: ")
			printChannelLayout(device.GetCurrentLayout())
			_, _ = fmt.Fprintln(os.Stderr)
		}

		_, _ = fmt.Fprintln(os.Stderr, "  sample rates:")
		for _, rate := range device.GetSampleRates() {
			_, _ = fmt.Fprintf(os.Stderr, "    %d - %d\n", rate.GetMin(), rate.GetMax())
		}
		if device.GetSampleRateCurrent() > 0 {
			_, _ = fmt.Fprintf(os.Stderr, "  current sample rate: %d\n", device.GetSampleRateCurrent())
		}

		_, _ = fmt.Fprint(os.Stderr, "  formats: ")
		for i, format := range device.GetFormats() {
			if i == 0 {
				_, _ = fmt.Fprint(os.Stderr, format)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, ", %s", format)
			}
		}
		_, _ = fmt.Fprintln(os.Stderr)

		if device.GetCurrentFormat() != soundio.FormatInvalid {
			_, _ = fmt.Fprintf(os.Stderr, "  current format: %s\n", device.GetCurrentFormat())
		}

		_, _ = fmt.Fprintf(os.Stderr, "  min software latency: %0.8f sec\n", device.GetSoftwareLatencyMin())
		_, _ = fmt.Fprintf(os.Stderr, "  max software latency: %0.8f sec\n", device.GetSoftwareLatencyMax())
		if device.GetSoftwareLatencyCurrent() != 0.0 {
			_, _ = fmt.Fprintf(os.Stderr, "  current software latency: %0.8f sec\n", device.GetSoftwareLatencyCurrent())
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "  probe error: %s\n", device.GetProbeError())
	}

	_, _ = fmt.Fprintln(os.Stderr)
}

func listDevices(s *soundio.SoundIo, shortOutput bool) {
	outputCount := s.OutputDeviceCount()
	inputCount := s.InputDeviceCount()

	defaultOutput := s.DefaultOutputDeviceIndex()
	defaultInput := s.DefaultInputDeviceIndex()

	_, _ = fmt.Fprintf(os.Stderr, "--------Input Devices--------\n\n")
	for i := 0; i < inputCount; i++ {
		device := s.GetInputDevice(i)
		printDevice(device, shortOutput, defaultInput == i)
		device.RemoveReference()
	}

	_, _ = fmt.Fprintf(os.Stderr, "--------Output Devices--------\n\n")
	for i := 0; i < outputCount; i++ {
		device := s.GetOutputDevice(i)
		printDevice(device, shortOutput, defaultOutput == i)
		device.RemoveReference()
	}

	_, _ = fmt.Fprintf(os.Stderr, "\n%d devices found\n", inputCount+outputCount)
}

func realMain(ctx context.Context, backend soundio.Backend, watchEvents bool, shortOutput bool) error {
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

	if watchEvents {
		s.SetOnDevicesChange(func(s *soundio.SoundIo) {
			_, _ = fmt.Fprintln(os.Stderr, "devices changed")
			listDevices(s, shortOutput)
		})

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
	} else {
		s.FlushEvents()
		listDevices(s, shortOutput)
	}

	return nil
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
