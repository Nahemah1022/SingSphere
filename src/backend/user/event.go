package user

import (
	"encoding/json"
	"fmt"

	"github.com/Nahemah1022/singsphere-backend/stereo"
	"github.com/pion/webrtc/v2"
)

// Event represents web socket user event
type Event struct {
	Type string `json:"type"`

	Offer     *webrtc.SessionDescription `json:"offer,omitempty"`
	Answer    *webrtc.SessionDescription `json:"answer,omitempty"`
	Candidate *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
	User      *UserWrap                  `json:"user,omitempty"`
	Room      *RoomWrap                  `json:"room,omitempty"`
	Desc      string                     `json:"desc,omitempty"`
	Song      *stereo.Song               `json:"song,omitempty"`
}

// SendEvent sends json body to web socket
func (u *User) SendEvent(event Event) error {
	json, err := json.Marshal(event)
	if err != nil {
		return err
	}
	u.send <- json
	return nil
}

// SendEventUser sends user to client to identify himself
func (u *User) SendEventUser() error {
	return u.SendEvent(Event{Type: "user", User: u.Wrap()})
}

// SendEventRoom sends room to client with users except me
func (u *User) SendEventRoom() error {
	return u.SendEvent(Event{Type: "room", Room: u.room.Wrap(u)})
}

// BroadcastEvent sends json body to everyone in the room except this user
func (u *User) BroadcastEvent(event Event) error {
	json, err := json.Marshal(event)
	if err != nil {
		return err
	}
	u.room.Broadcast(json, u)
	return nil
}

// BroadcastEventJoin sends user_join event
func (u *User) BroadcastEventJoin() error {
	return u.BroadcastEvent(Event{Type: "user_join", User: u.Wrap()})
}

// BroadcastEventLeave sends user_leave event
func (u *User) BroadcastEventLeave() error {
	return u.BroadcastEvent(Event{Type: "user_leave", User: u.Wrap()})
}

// BroadcastEventMute sends microphone mute event to everyone
func (u *User) BroadcastEventMute() error {
	return u.BroadcastEvent(Event{Type: "mute", User: u.Wrap()})
}

// BroadcastEventUnmute sends microphone unmute event to everyone
func (u *User) BroadcastEventUnmute() error {
	return u.BroadcastEvent(Event{Type: "unmute", User: u.Wrap()})
}

// SendErr sends error in json format to web socket
func (u *User) SendErr(err error) error {
	return u.SendEvent(Event{Type: "error", Desc: fmt.Sprint(err)})
}
