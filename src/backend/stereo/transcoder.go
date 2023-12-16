package stereo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var (
	input_dir  = os.Getenv("TRANSCODE_INPUT_PATH")
	output_dir = os.Getenv("TRANSCODE_OUTPUT_PATH")
)

func Trans(filename string, playlist chan<- string) {
	if input_dir == "" || output_dir == "" {
		input_dir = "./s3"
		output_dir = "./media"
		// panic("env var not found: TRANSCODE_INPUT_PATH or TRANSCODE_OUTPUT_PATH")
	}
	outpath := fmt.Sprintf("%s/%s.ogg", output_dir, filename)
	inpath := fmt.Sprintf("%s/%s", input_dir, filename)
	if _, err := os.Stat(inpath); errors.Is(err, os.ErrNotExist) {
		log.Printf("requested file '%s' not found\n", inpath)
		return
	}

	err := ffmpeg.Input(inpath).
		Output(outpath, ffmpeg.KwArgs{"c:a": "libopus", "page_duration": 20000, "loglevel": "debug"}).
		OverWriteOutput().ErrorToStdOut().Run()
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(outpath); errors.Is(err, os.ErrNotExist) {
		log.Printf("converted file '%s' not found\n", outpath)
		return
	}

	log.Printf("convert to opus completed, output filepath: %s\n", outpath)
	playlist <- outpath
}

func Play(filepath string, targetTrack *webrtc.Track, cancel context.CancelFunc) {
	// Open a IVF file and start reading using our IVFReader
	file, oggErr := os.Open(filepath)
	if oggErr != nil {
		panic(oggErr)
	}

	// Open on oggfile in non-checksum mode.
	ogg, _, oggErr := NewWith(file)
	if oggErr != nil {
		panic(oggErr)
	}

	// Wait for connection established
	// <-iceConnectedCtx.Done()

	// Keep track of last granule, the difference is the amount of samples in the buffer
	var lastGranule uint64
	for {
		pageData, pageHeader, oggErr := ogg.ParseNextPage()
		// log.Println(pageData)
		if oggErr == io.EOF {
			fmt.Println("All audio pages parsed and sent")
			break
			// os.Exit(0)
		}

		if oggErr != nil {
			panic(oggErr)
		}

		// The amount of samples is the difference between the last and current timestamp
		sampleCount := float64((pageHeader.GranulePosition - lastGranule))
		lastGranule = pageHeader.GranulePosition

		if oggErr = targetTrack.WriteSample(media.Sample{Data: pageData, Samples: uint32(sampleCount)}); oggErr != nil {
			panic(oggErr)
		}

		// Convert seconds to Milliseconds, Sleep doesn't accept floats
		time.Sleep(time.Duration((sampleCount/48000)*1000) * time.Millisecond)
	}
	cancel()
}
