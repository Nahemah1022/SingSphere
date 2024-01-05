package user

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/socket"
	"github.com/pion/webrtc/v3"
)

type UserWrap struct {
	ID    string `json:"id"`
	Emoji string `json:"emoji"`
	Mute  bool   `json:"mute"`
}

type User struct {
	ID      string
	Emoji   string
	Mute    bool
	ws      *socket.Websocket
	pc      *webrtc.PeerConnection
	joinCh  chan *User
	leaveCh chan *User
}

var emojis = []string{
	"😎", "🧐", "🤡", "👻", "😷", "🤗", "😏",
	"👽", "👨‍🚀", "🐺", "🐯", "🦁", "🐶", "🐼", "🙈",
}

func New(joinCh chan *User, leaveCh chan *User, ws *socket.Websocket) (*User, error) {
	newUser := &User{
		ID:      strconv.FormatInt(time.Now().UnixNano(), 10), // generate random id based on timestamp
		Mute:    true,
		Emoji:   emojis[rand.Intn(len(emojis))],
		joinCh:  joinCh,
		leaveCh: leaveCh,
		ws:      ws,
	}

	// Establish webrtc peer connection
	if err := newUser.peerConnect(); err != nil {
		log.Println(err)
		return nil, err
	}
	return newUser, nil
}

func (u *User) handleInboundEvent(event *socket.InboundEvent) error {
	u.log("handle event: ", event.Type)
	return nil
}

func (u *User) SendEvent(event *socket.OutboundEvent) error {
	if err := u.ws.Send(event); err != nil {
		return err
	}
	return nil
}

func (u *User) Run() {
	defer func() {
		u.pc.Close()
		u.leaveCh <- u
	}()
	u.joinCh <- u
	go u.ws.Run()
	for inboundEvent := range u.ws.InboundEventCh {
		if err := u.handleInboundEvent(inboundEvent); err != nil {
			u.ws.SendError(errors.New("fail to handle inbound event"))
		}
	}
}

func (u *User) log(msg ...interface{}) {
	log.Println(
		fmt.Sprintf("user %s:", u.ID),
		fmt.Sprint(msg...),
	)
}
