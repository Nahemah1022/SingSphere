package socket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pion/webrtc/v3"
)

// Event represents web socket user event
type EventBase struct {
	Type string `json:"type"`
	Desc string `json:"description,omitempty"`
}

type InboundEvent struct {
	EventBase
	Offer     *webrtc.SessionDescription `json:"offer,omitempty"`
	Answer    *webrtc.SessionDescription `json:"answer,omitempty"`
	Candidate *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
}

type OutboundEvent struct {
	EventBase
	// Sender string `json:"sender,omitempty"` // Should be speficied if it is a broadcast event
	Offer     *webrtc.SessionDescription `json:"offer,omitempty"`
	Answer    *webrtc.SessionDescription `json:"answer,omitempty"`
	Candidate *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
	User      *UserWrap                  `json:"user,omitempty"`
}

type UserWrap struct {
	ID    string `json:"id"`
	Emoji string `json:"emoji"`
	Mute  bool   `json:"mute"`
}

// SendEvent enocde event json body an sends it to write loop
func (ws *Websocket) Send(event *OutboundEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	ws.sendCh <- bytes
	return nil
}

// SendErr sends error in json format to web socket
func (ws *Websocket) SendError(err error) error {
	return ws.Send(&OutboundEvent{EventBase: EventBase{Type: "error", Desc: fmt.Sprint(err)}})
}

// recieveEvent decode inbound event raw bytes and push it to public channel for user access
func (ws *Websocket) recieveEvent(eventRaw []byte) {
	var event *InboundEvent
	if err := json.Unmarshal(eventRaw, &event); err != nil {
		log.Println(err)
		return
	}
	ws.InboundEventCh <- event
}
