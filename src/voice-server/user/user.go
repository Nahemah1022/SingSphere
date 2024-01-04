package user

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
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
	conn    *websocket.Conn
	pc      *webrtc.PeerConnection
	joinCh  chan *User
	leaveCh chan *User
}

func New(joinCh chan *User, leaveCh chan *User, w http.ResponseWriter, req *http.Request) (*User, error) {
	newUser := &User{
		ID:      strconv.FormatInt(time.Now().UnixNano(), 10), // generate random id based on timestamp
		Mute:    true,
		joinCh:  joinCh,
		leaveCh: leaveCh,
	}

	// Establish websocket connection
	if err := newUser.WsConnect(w, req); err != nil {
		log.Println(err)
		return nil, err
	}

	// Establish webrtc peer connection
	if err := newUser.PeerConnect(); err != nil {
		log.Println(err)
		return nil, err
	}
	return newUser, nil
}

func (u *User) Run() {
	defer func() {
		u.pc.Close()
		u.conn.Close()
		u.leaveCh <- u
	}()
	// infinite loop to read websocket message until connection closed
	u.joinCh <- u
	u.wsRead()
}
