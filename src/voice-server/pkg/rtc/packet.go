package rtc

import (
	"errors"
	"io"
	"log"

	"github.com/pion/webrtc/v3"
)

var (
	errChanClosed    = errors.New("channel closed")
	errInvalidTrack  = errors.New("track is nil")
	errInvalidPacket = errors.New("packet is nil")
)

// receiveTrackRTP receive mic track's rtp and sent to one channel
func (node *RtcNode) receiveTrackRTP(track *webrtc.TrackRemote) {
	for {
		rtp, _, err := track.ReadRTP()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Fatalf("rtp err => %v", err)
		}
		node.RtpCh <- rtp
	}
}

// ReadRTP read rtp packet
// func (node *RtcNode) ReadRTP() (*rtp.Packet, error) {
// 	rtp, ok := <-node.RtpCh
// 	if !ok {
// 		return nil, errChanClosed
// 	}
// 	return rtp, nil
// }

// WriteRTP send rtp packet to user outgoing tracks
// func (node *RtcNode) WriteRTP(pkt *rtp.Packet, track *webrtc.TrackLocalStaticRTP) error {
// 	if pkt == nil {
// 		return errInvalidPacket
// 	}

// 	if track == nil {
// 		// log.Printf("WebRTCTransport.WriteRTP track==nil pkt.SSRC=%d", pkt.SSRC)
// 		return errInvalidTrack
// 	}

// 	// log.Debugf("WebRTCTransport.WriteRTP pkt=%v", pkt)
// 	err := track.WriteRTP(pkt)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
