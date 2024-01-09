package rtc

import (
	"errors"

	"github.com/pion/webrtc/v3"
)

// AddTrack adds the given track to this RTC node's peer connection instance
func (node *RtcNode) AddTrack(track *webrtc.TrackLocalStaticRTP) (*webrtc.RTPSender, error) {
	sender, err := node.pc.AddTrack(track)
	if err != nil {
		return nil, err
	}
	return sender, nil
}

// RemoveTrack remove the given track from this RTC node's peer connection instance
func (node *RtcNode) RemoveTrack(ssrc webrtc.SSRC) error {
	node.SendersLock.Lock()
	defer node.SendersLock.Unlock()
	sender, ok := node.Senders[ssrc]
	if !ok {
		return errors.New("sender doesn't exist")
	}
	if err := node.pc.RemoveTrack(sender); err != nil {
		return err
	}
	return nil
}
