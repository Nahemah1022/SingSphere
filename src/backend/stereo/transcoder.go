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
	"github.com/tcolgate/mp3"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Song struct {
	Name     string
	Path     string
	Duration uint
}

var (
	input_dir  = os.Getenv("TRANSCODE_INPUT_PATH")
	output_dir = os.Getenv("TRANSCODE_OUTPUT_PATH")
)

func Trans(filename string, playlist chan<- *Song) {
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
		Output(outpath, ffmpeg.KwArgs{"c:a": "libopus", "page_duration": 10000, "loglevel": "debug"}).
		OverWriteOutput().ErrorToStdOut().Run()
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(outpath); errors.Is(err, os.ErrNotExist) {
		log.Printf("transcoded file '%s' not found\n", outpath)
		return
	}

	log.Printf("convert to opus completed, output filepath: %s\n", outpath)

	t, err := getMP3Length(inpath)
	if err != nil {
		panic(err)
	}

	song := &Song{
		Name:     filename,
		Path:     outpath,
		Duration: t,
	}
	log.Printf("song is enqueued: %v\n", song)
	playlist <- song
}

func getMP3Length(file string) (uint, error) {
	t := 0.0

	r, err := os.Open(file)
	if err != nil {
		return 0, err
	}

	d := mp3.NewDecoder(r)
	var f mp3.Frame
	skipped := 0

	for {
		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return 0, err
		}

		t = t + f.Duration().Seconds()
	}

	return uint(t), nil
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
