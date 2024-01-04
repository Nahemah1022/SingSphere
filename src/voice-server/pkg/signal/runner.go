package signal

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Websocket struct {
	conn         *websocket.Conn
	closeHandler func()
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 51200
)

var (
	newline  = []byte{'\n'}
	space    = []byte{' '}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func New(w http.ResponseWriter, r *http.Request, closeHandler func()) (*Websocket, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	ws := &Websocket{
		conn:         conn,
		closeHandler: closeHandler,
	}
	log.Println("ws connected")
	return ws, nil
}

func (ws *Websocket) ReadLoop() {
	defer ws.conn.Close()
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				go ws.closeHandler()
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		log.Println("recieved message: ", message)
		// go func() {
		// 	err := u.HandleEvent(message)
		// 	if err != nil {
		// 		log.Println(err)
		// 		u.SendErr(err)
		// 	}
		// }()
	}
}
