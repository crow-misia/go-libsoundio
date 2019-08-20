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
	"log"
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
		log.Println(err)
	} else {
		ctx := context.Background()
		parentCtx := signalContext(ctx)
		err := realMain(parentCtx, enumBackend, watchEvents, shortOutput)
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

func printChannelLayout(layout *soundio.ChannelLayout) {
	name := layout.Name()
	if len(name) == 0 {
		names := make([]string, layout.ChannelCount())
		for i, channel := range layout.Channels() {
			names[i] = fmt.Sprint(channel)
		}
		log.Printf("    %s", strings.Join(names, ", "))
	} else {
		log.Printf("    %s", name)
	}
}

func printDevice(device *soundio.Device, shortOutput bool, isDefault bool) {
	defaultStr := ""
	if isDefault {
		defaultStr = " (default)"
	}

	rawStr := ""
	if device.Raw() {
		rawStr = " (raw)"
	}

	log.Printf("%s%s%s", device.Name(), defaultStr, rawStr)
	if shortOutput {
		return
	}

	log.Printf("  id: %s", device.ID())

	if device.ProbeError() == nil {
		log.Println("  channel layouts:")
		for _, layout := range device.Layouts() {
			printChannelLayout(layout)
		}
		if device.CurrentLayout().ChannelCount() > 0 {
			log.Print("  current layout: ")
			printChannelLayout(device.CurrentLayout())
		}

		log.Println("  sample rates:")
		for _, rate := range device.SampleRates() {
			log.Printf("    %d - %d", rate.Min(), rate.Max())
		}
		if device.SampleRateCurrent() > 0 {
			log.Printf("  current sample rate: %d", device.SampleRateCurrent())
		}

		formats := make([]string, device.FormatCount())
		for i, format := range device.Formats() {
			formats[i] = fmt.Sprint(format)
		}
		log.Printf("  formats: %s", strings.Join(formats, ", "))

		if device.CurrentFormat() != soundio.FormatInvalid {
			log.Printf("  current format: %s", device.CurrentFormat())
		}

		log.Printf("  min software latency: %0.8f sec", device.SoftwareLatencyMin())
		log.Printf("  max software latency: %0.8f sec", device.SoftwareLatencyMax())
		if device.SoftwareLatencyCurrent() != 0.0 {
			log.Printf("  current software latency: %0.8f sec", device.SoftwareLatencyCurrent())
		}
	} else {
		log.Printf("  probe error: %s", device.ProbeError())
	}

	log.Println()
}

func listDevices(s *soundio.SoundIo, shortOutput bool) {
	outputCount := s.OutputDeviceCount()
	inputCount := s.InputDeviceCount()

	defaultOutput := s.DefaultOutputDeviceIndex()
	defaultInput := s.DefaultInputDeviceIndex()

	log.Println("--------Input Devices--------")
	for i := 0; i < inputCount; i++ {
		device, err := s.InputDevice(i)
		if err != nil {
			log.Print(err)
		} else {
			printDevice(device, shortOutput, defaultInput == i)
			device.RemoveReference()
		}
	}

	log.Println("--------Output Devices--------")
	for i := 0; i < outputCount; i++ {
		device, err := s.OutputDevice(i)
		if err != nil {
			log.Print(err)
		} else {
			printDevice(device, shortOutput, defaultOutput == i)
			device.RemoveReference()
		}
	}

	log.Println()
	log.Printf("%d devices found", inputCount+outputCount)
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
			log.Println("devices changed")
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
