package user

import (
	"fmt"
	"io"
	"log"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

// receiveInTrackRTP receive all incoming tracks' rtp and sent to one channel
func (u *User) receiveInTrackRTP(remoteTrack *webrtc.Track) {
	for {
		if u.stop {
			return
		}
		rtp, err := remoteTrack.ReadRTP()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Fatalf("rtp err => %v", err)
		}
		u.rtpCh <- rtp
	}
}

// ReadRTP read rtp packet
func (u *User) ReadRTP() (*rtp.Packet, error) {
	rtp, ok := <-u.rtpCh
	if !ok {
		return nil, errChanClosed
	}
	return rtp, nil
}

// WriteRTP send rtp packet to user outgoing tracks
func (u *User) WriteRTP(pkt *rtp.Packet) error {
	if pkt == nil {
		return errInvalidPacket
	}
	u.outTracksLock.RLock()
	track := u.outTracks[pkt.SSRC]
	u.outTracksLock.RUnlock()

	if track == nil {
		log.Printf("WebRTCTransport.WriteRTP track==nil pkt.SSRC=%d", pkt.SSRC)
		return errInvalidTrack
	}

	// log.Debugf("WebRTCTransport.WriteRTP pkt=%v", pkt)
	err := track.WriteRTP(pkt)
	if err != nil {
		// log.Errorf(err.Error())
		// u.writeErrCnt++
		return err
	}
	return nil
}

func (u *User) broadcastIncomingRTP() {
	for {
		rtp, err := u.ReadRTP()
		if err != nil {
			panic(err)
		}
		for _, user := range u.room.GetOtherUsers(u) {
			err := user.WriteRTP(rtp)
			if err != nil {
				// panic(err)
				fmt.Println(err)
			}
		}
	}
}
