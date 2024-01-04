package user

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/signal"
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
	ws      *signal.Websocket
	pc      *webrtc.PeerConnection
	joinCh  chan *User
	leaveCh chan *User
}

var emojis = []string{
	"😎", "🧐", "🤡", "👻", "😷", "🤗", "😏",
	"👽", "👨‍🚀", "🐺", "🐯", "🦁", "🐶", "🐼", "🙈",
}

func New(joinCh chan *User, leaveCh chan *User, ws *signal.Websocket) (*User, error) {
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

func (u *User) Run() {
	defer func() {
		u.pc.Close()
		u.leaveCh <- u
	}()
	// infinite loop to read websocket message until connection closed
	u.joinCh <- u
	u.ws.ReadLoop()
}
