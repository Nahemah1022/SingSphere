package user

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
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
	joinCh  chan *User
	leaveCh chan *User
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (u *User) PeerConnect() error {
	log.Println("PC connected")
	return nil
}

func New(joinCh chan *User, leaveCh chan *User) *User {
	return &User{
		ID:      strconv.FormatInt(time.Now().UnixNano(), 10), // generate random id based on timestamp
		Mute:    true,
		joinCh:  joinCh,
		leaveCh: leaveCh,
	}
}

func (u *User) Run() {
	defer func() {
		// u.pc.Close()
		u.leaveCh <- u
		u.conn.Close()
	}()
	// infinite loop to read websocket message until connection closed
	u.joinCh <- u
	u.wsRead()
}
