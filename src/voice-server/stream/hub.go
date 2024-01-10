package stream

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
)

type AudioHub struct {
	audioTrack *webrtc.TrackLocalStaticSample
	roomName   string
}

// New creates a new audio hub for the given room
func New(roomName string) (*AudioHub, error) {
	audioTrack, audioTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "music", "hub")
	if audioTrackErr != nil {
		return nil, audioTrackErr
	}
	hub := &AudioHub{
		audioTrack: audioTrack,
		roomName:   roomName,
	}
	return hub, nil
}

// StreamAudioFile starts to stream the given audio file in this hub
func (hub *AudioHub) StreamAudioFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Println(err)
		return
	}
	ogg, _, oggErr := oggreader.NewWith(file)
	if oggErr != nil {
		panic(oggErr)
	}
	// Keep track of last granule, the difference is the amount of samples in the buffer
	var lastGranule uint64

	// It is important to use a time.Ticker instead of time.Sleep because
	// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
	// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
	ticker := time.NewTicker(time.Millisecond * 20)
	for ; true; <-ticker.C {
		pageData, pageHeader, oggErr := ogg.ParseNextPage()
		if errors.Is(oggErr, io.EOF) {
			log.Println("All audio pages parsed and sent")
			return
		}

		if oggErr != nil {
			log.Println(oggErr)
			return
		}

		// The amount of samples is the difference between the last and current timestamp
		sampleCount := float64(pageHeader.GranulePosition - lastGranule)
		lastGranule = pageHeader.GranulePosition
		sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

		if oggErr = hub.audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); oggErr != nil {
			log.Println(oggErr)
			return
		}
	}
}
