package user

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/Nahemah1022/singsphere-backend/stereo"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

// GetInTracks return incoming tracks
func (u *User) GetInTracks() map[uint32]*webrtc.Track {
	u.inTracksLock.RLock()
	defer u.inTracksLock.RUnlock()
	return u.inTracks

}

// GetOutTracks return outgoing tracks
func (u *User) GetOutTracks() map[uint32]*webrtc.Track {
	u.outTracksLock.RLock()
	defer u.outTracksLock.RUnlock()
	return u.outTracks
}

// GetRoomTracks returns list of room incoming tracks, and the room's stereo track
func (u *User) GetRoomTracks() []*webrtc.Track {
	tracks := []*webrtc.Track{}
	for _, user := range u.room.GetUsers() {
		for _, track := range user.inTracks {
			tracks = append(tracks, track)
		}
	}
	return tracks
}

// AddTrack adds track to peer connection
func (u *User) AddTrack(ssrc uint32) error {
	track, err := u.pc.NewTrack(webrtc.DefaultPayloadTypeOpus, ssrc, string(ssrc), string(ssrc))
	if err != nil {
		return err
	}
	if _, err := u.pc.AddTrack(track); err != nil {
		log.Println("ERROR Add remote track as peerConnection local track", err)
		return err
	}

	u.outTracksLock.Lock()
	u.outTracks[track.SSRC()] = track
	u.outTracksLock.Unlock()
	return nil
}

func (u *User) AddStereoTrack() error {
	// iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())
	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	audioTrack, addTrackErr := u.pc.NewTrack(getPayloadType(mediaEngine, webrtc.RTPCodecTypeAudio, "OPUS"), rand.Uint32(), "audio", "pion")
	if addTrackErr != nil {
		panic(addTrackErr)
	}
	if _, addTrackErr = u.pc.AddTrack(audioTrack); addTrackErr != nil {
		panic(addTrackErr)
	}
	u.stereoTreck = audioTrack

	go func() {
		// Open a IVF file and start reading using our IVFReader
		file, oggErr := os.Open("./media/decoded/sample.mp3.ogg")
		if oggErr != nil {
			panic(oggErr)
		}

		// Open on oggfile in non-checksum mode.
		ogg, _, oggErr := stereo.NewWith(file)
		if oggErr != nil {
			panic(oggErr)
		}

		// Wait for connection established
		// <-iceConnectedCtx.Done()

		// Keep track of last granule, the difference is the amount of samples in the buffer
		var lastGranule uint64
		for {
			pageData, pageHeader, oggErr := ogg.ParseNextPage()
			if oggErr == io.EOF {
				fmt.Printf("All audio pages parsed and sent")
				os.Exit(0)
			}

			if oggErr != nil {
				panic(oggErr)
			}

			// The amount of samples is the difference between the last and current timestamp
			sampleCount := float64((pageHeader.GranulePosition - lastGranule))
			lastGranule = pageHeader.GranulePosition

			if oggErr = audioTrack.WriteSample(media.Sample{Data: pageData, Samples: uint32(sampleCount)}); oggErr != nil {
				panic(oggErr)
			}

			// Convert seconds to Milliseconds, Sleep doesn't accept floats
			time.Sleep(time.Duration((sampleCount/48000)*1000) * time.Millisecond)
		}
	}()

	return nil
}

func getPayloadType(m webrtc.MediaEngine, codecType webrtc.RTPCodecType, codecName string) uint8 {
	for _, codec := range m.GetCodecsByKind(codecType) {
		if codec.Name == codecName {
			return codec.PayloadType
		}
	}
	panic(fmt.Sprintf("Remote peer does not support %s", codecName))
}
