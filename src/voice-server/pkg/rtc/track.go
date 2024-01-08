package rtc

import "github.com/pion/webrtc/v3"

// AddTrack adds the given track to this RTC node's peer connection instance
func (node *RtcNode) AddTrack(track *webrtc.TrackLocalStaticRTP) (*webrtc.RTPSender, error) {
	sender, err := node.pc.AddTrack(track)
	if err != nil {
		return nil, err
	}
	return sender, nil
}

// RemoveTrack remove the given track from this RTC node's peer connection instance
func (node *RtcNode) RemoveTrack(track *webrtc.TrackLocalStaticRTP) error {
	// node.pc.RemoveTrack()
	return nil
}
