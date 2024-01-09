package user

import (
	"errors"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// ReadRTP receive packets from user's rtp channel
func (u *User) ReadRTP() (*rtp.Packet, error) {
	rtp, ok := <-u.rtc.RtpCh
	if !ok {
		return nil, errors.New("channel closed")
	}
	return rtp, nil
}

// WriteRTP select the target track from map, and send the packet
func (u *User) WriteRTP(pkt *rtp.Packet, ssrc webrtc.SSRC) error {
	u.rtc.ListeningTracksLock.RLock()
	targetTrack := u.rtc.ListeningTracks[ssrc]
	u.rtc.ListeningTracksLock.RUnlock()

	if targetTrack == nil {
		return errors.New("track is nil")
	}

	if err := targetTrack.WriteRTP(pkt); err != nil {
		return err
	}
	return nil
}
