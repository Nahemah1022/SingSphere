package socket

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Websocket struct {
	conn           *websocket.Conn
	stop           bool
	wg             sync.WaitGroup
	InboundEventCh chan *InboundEvent // Public interface for parsed inbound messages
	sendCh         chan []byte        // Buffered channel of outbound messages.
	closeHandler   func()
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = time.Second
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

// New construct and return a new websocket session as a standalone dependency
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
		conn:           conn,
		closeHandler:   closeHandler,
		stop:           false,
		InboundEventCh: make(chan *InboundEvent, 16),
		sendCh:         make(chan []byte, 256),
	}
	log.Println("ws connected")
	return ws, nil
}

// Run starts the readLoop and writeLoop of the websocket, and terminate until one of them ends
func (ws *Websocket) Run() {
	defer func() {
		ws.conn.Close()
		close(ws.sendCh)
		close(ws.InboundEventCh)
	}()
	ws.wg.Add(2)
	go ws.readLoop()
	go ws.writeLoop()
	ws.wg.Wait()
}

// ReadLoop is a goroutine that should be fired once the owner is ready to receive message from websocket
func (ws *Websocket) readLoop() {
	defer func() {
		ws.stop = true
		ws.wg.Done()
	}()
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ws.closeHandler()
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		go ws.receiveEvent(message)
	}
}

// ReadLoop is a goroutine that should be fired once the owner is ready to send message to websocket
func (ws *Websocket) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		ws.stop = true
		ws.wg.Done()
	}()
	for {
		select {
		case message, ok := <-ws.sendCh:
			ws.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				ws.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := ws.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			ws.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
