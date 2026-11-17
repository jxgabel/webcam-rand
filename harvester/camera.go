package main

import (
	"context"
	"errors"
	"log"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func OpenCameras(ctx context.Context, sources []string) ([]*device.Device, error) {

	var cams []*device.Device

	for _, src := range sources {

		dev, err := device.Open(src, device.WithPixFormat(v4l2.PixFormat{
			Width:       320,
			Height:      240,
			PixelFormat: v4l2.PixelFmtJPEG, // or v4l2.PixelFmtYUYV if MJPEG isn't supported
		}))

		if err != nil {
			log.Printf("error: %s: %v", src, err)
			continue
		}

		err = dev.Start(ctx)
		if err != nil {
			dev.Close()
			log.Printf("error: %s: %v", src, err)
			continue
		}

		cams = append(cams, dev)
	}

	if len(cams) == 0 {
		return nil, errors.New("No Cameras were found.")
	}

	return cams, nil
}

func CaptureFrames(cams []*device.Device) ([][]byte, error) {

	frames := make([][]byte, 0, len(cams))

	for _, c := range cams {

		frame := <-c.GetFrames()

		if len(frame.Data) != 0 {
			frData := make([]byte, len(frame.Data))
			copy(frData, frame.Data)
			frames = append(frames, frData)
		}
		frame.Release()

	}

	if len(frames) == 0 {
		return nil, errors.New("No frames could be aquired.")
	}

	return frames, nil
}

func CloseCameras(cams []*device.Device) {
	for _, c := range cams {
		c.Close()
	}
}
