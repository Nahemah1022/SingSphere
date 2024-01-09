package user

import (
	"errors"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/socket"
	"github.com/pion/webrtc/v3"
)

func (u *User) SendCandidate(iceCandidate *webrtc.ICECandidate) error {
	if iceCandidate == nil {
		return errors.New("nil ice candidate")
	}
	iceCandidateInit := iceCandidate.ToJSON()
	if err := u.SendEvent(&socket.OutboundEvent{
		EventBase: socket.EventBase{Type: "candidate"},
		Candidate: &iceCandidateInit,
	}); err != nil {
		return err
	}
	return nil

}

// SendAnswer creates answer and send it via websocket
func (u *User) SendAnswer(answer webrtc.SessionDescription) error {
	if err := u.SendEvent(&socket.OutboundEvent{
		EventBase: socket.EventBase{Type: "answer"},
		Answer:    &answer,
	}); err != nil {
		return err
	}
	return nil
}

// SendOffer creates webrtc offer
func (u *User) SendOffer() error {
	offer, err := u.rtc.Offer()
	if err != nil {
		return err
	}
	if err := u.SendEvent(&socket.OutboundEvent{
		EventBase: socket.EventBase{Type: "offer"},
		Offer:     &offer,
	}); err != nil {
		return err
	}
	return nil
}

// SendEvent sends an outbound event to this user
func (u *User) SendEvent(event *socket.OutboundEvent) error {
	if err := u.ws.Send(event); err != nil {
		return err
	}
	return nil
}
