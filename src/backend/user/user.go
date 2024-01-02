package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

var (
	// only support unified plan
	// cfg = webrtc.Configuration{
	// 	SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	// }

	// prepare the configuration
	peerConnectionConfig = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// setting webrtc.SettingEngine

	errChanClosed    = errors.New("channel closed")
	errInvalidTrack  = errors.New("track is nil")
	errInvalidPacket = errors.New("packet is nil")
	// errInvalidPC      = errors.New("pc is nil")
	// errInvalidOptions = errors.New("invalid options")
	errNotImplemented = errors.New("not implemented")
)

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
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// User is a middleman between the websocket connection and the hub.
type User struct {
	ID            string
	room          *Room
	conn          *websocket.Conn          // The websocket connection.
	send          chan []byte              // Buffered channel of outbound messages.
	pc            *webrtc.PeerConnection   // WebRTC Peer Connection
	inTracks      map[uint32]*webrtc.Track // Microphone
	inTracksLock  sync.RWMutex
	outTracks     map[uint32]*webrtc.Track // Rest of the room's tracks
	outTracksLock sync.RWMutex

	rtpCh chan *rtp.Packet

	stop bool

	info UserInfo
}

// UserInfo contains some user data
type UserInfo struct {
	Emoji string `json:"emoji"` // emoji-face like on clients (for test)
	Mute  bool   `json:"mute"`
}

// UserWrap represents user object sent to client
type UserWrap struct {
	ID string `json:"id"`
	UserInfo
}

// Wrap wraps user
func (u *User) Wrap() *UserWrap {
	return &UserWrap{
		ID:       u.ID,
		UserInfo: u.info,
	}
}

func (u *User) log(msg ...interface{}) {
	log.Println(
		fmt.Sprintf("user %s:", u.ID),
		fmt.Sprint(msg...),
	)
}

// HandleEvent handles user event from web socket
func (u *User) HandleEvent(eventRaw []byte) error {
	var event *Event
	err := json.Unmarshal(eventRaw, &event)
	if err != nil {
		return err
	}
	u.log("handle event: ", event.Type)
	if event.Type == "offer" {
		if event.Offer == nil {
			return u.SendErr(errors.New("empty offer"))
		}
		err := u.HandleOffer(*event.Offer)
		if err != nil {
			return err
		}
		return nil
	} else if event.Type == "answer" {
		if event.Answer == nil {
			return u.SendErr(errors.New("empty answer"))
		}
		time.Sleep(3000 * time.Millisecond)
		u.pc.SetRemoteDescription(*event.Answer)
		return nil
	} else if event.Type == "candidate" {
		if event.Candidate == nil {
			return u.SendErr(errors.New("empty candidate"))
		}
		u.log("adding candidate")
		u.pc.AddICECandidate(*event.Candidate)
		return nil
	} else if event.Type == "mute" {
		u.info.Mute = true
		u.BroadcastEventMute()
		return nil
	} else if event.Type == "unmute" {
		u.info.Mute = false
		u.BroadcastEventUnmute()
		return nil
	}

	return u.SendErr(errNotImplemented)
}

func (u *User) supportOpus(offer webrtc.SessionDescription) bool {
	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.PopulateFromSDP(offer)
	var payloadType uint8
	// Search for Payload type. If the offer doesn't support codec exit since
	// since they won't be able to decode anything we send them
	for _, audioCodec := range mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
		if audioCodec.Name == "OPUS" {
			payloadType = audioCodec.PayloadType
			break
		}
	}
	return payloadType != 0
}

// HandleOffer handles webrtc offer
func (u *User) HandleOffer(offer webrtc.SessionDescription) error {
	if ok := u.supportOpus(offer); !ok {
		return errors.New("remote peer does not support opus codec")
	}

	if len(u.pc.GetTransceivers()) == 0 {
		// add receive only transciever to get user microphone audio
		_, err := u.pc.AddTransceiver(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		})
		if err != nil {
			return err
		}
	}

	// Set the remote SessionDescription
	if err := u.pc.SetRemoteDescription(offer); err != nil {
		return err
	}

	err := u.SendAnswer()
	if err != nil {
		return err
	}

	return nil
}

// Offer return a offer
func (u *User) Offer() (webrtc.SessionDescription, error) {
	offer, err := u.pc.CreateOffer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	err = u.pc.SetLocalDescription(offer)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	return offer, nil
}

// SendOffer creates webrtc offer
func (u *User) SendOffer() error {
	offer, err := u.Offer()
	if err != nil {
		panic(err)
	}
	err = u.SendEvent(Event{Type: "offer", Offer: &offer})
	if err != nil {
		panic(err)
	}
	return nil
}

// SendCandidate sends ice candidate to peer
func (u *User) SendCandidate(iceCandidate *webrtc.ICECandidate) error {
	if iceCandidate == nil {
		return errors.New("nil ice candidate")
	}
	iceCandidateInit := iceCandidate.ToJSON()
	err := u.SendEvent(Event{Type: "candidate", Candidate: &iceCandidateInit})
	if err != nil {
		return err
	}
	return nil
}

// Answer creates webrtc answer
func (u *User) Answer() (webrtc.SessionDescription, error) {
	answer, err := u.pc.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	// Sets the LocalDescription, and starts our UDP listeners
	if err = u.pc.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}
	return answer, nil
}

// SendAnswer creates answer and send it via websocket
func (u *User) SendAnswer() error {
	answer, err := u.Answer()
	if err != nil {
		return err
	}
	err = u.SendEvent(Event{Type: "answer", Answer: &answer})
	if err != nil {
		panic(err)
	}
	return nil
}

// Watch for debug
func (u *User) Watch() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		if u.stop {
			ticker.Stop()
			return
		}
		fmt.Println("ID:", u.ID, "out: ", u.GetOutTracks())
		fmt.Println("ID:", u.ID, "in: ", u.GetInTracks())
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(rooms *Rooms, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	peerConnection, err := api.NewPeerConnection(peerConnectionConfig)

	roomID := strings.ReplaceAll(r.URL.Path, "/", "")
	room := rooms.GetOrCreate(roomID)

	log.Println("ws connection to room:", roomID, len(room.GetUsers()), "users")

	emojis := []string{
		"😎", "🧐", "🤡", "👻", "😷", "🤗", "😏",
		"👽", "👨‍🚀", "🐺", "🐯", "🦁", "🐶", "🐼", "🙈",
	}

	user := &User{
		ID:        strconv.FormatInt(time.Now().UnixNano(), 10), // generate random id based on timestamp
		room:      room,
		conn:      conn,
		send:      make(chan []byte, 256),
		pc:        peerConnection,
		inTracks:  make(map[uint32]*webrtc.Track),
		outTracks: make(map[uint32]*webrtc.Track),
		rtpCh:     make(chan *rtp.Packet, 100),

		info: UserInfo{
			Emoji: emojis[rand.Intn(len(emojis))],
			Mute:  true, // user is muted by default
		},
	}

	user.pc.OnICECandidate(func(iceCandidate *webrtc.ICECandidate) {
		if iceCandidate != nil {
			err := user.SendCandidate(iceCandidate)
			if err != nil {
				log.Println("fail send candidate", err)
			}
		}
	})

	user.pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("user joined")
			tracks := user.GetRoomTracks()
			fmt.Println("attach ", len(tracks), "tracks to new user")
			user.log("new user add tracks", len(tracks))
			for _, track := range tracks {
				if err := user.AddTrack(track.SSRC()); err != nil {
					log.Println("ERROR Add remote track as peerConnection local track", err)
					panic(err)
				}
			}

			if _, err := user.pc.AddTrack(user.room.stereoTrack); err != nil {
				// if err := user.AddStereoTrack(); err != nil {
				log.Println("ERROR Add stereo track as peerConnection local track", err)
				panic(err)
			}

			err = user.SendOffer()
			if err != nil {
				panic(err)
			}
			user.room.Join(user)
			go user.room.StereoPlay()
			// user.room.requests <- "peach.mp3"
		} else if connectionState == webrtc.ICEConnectionStateDisconnected ||
			connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateClosed {

			user.stop = true

			for _, roomUser := range user.room.GetOtherUsers(user) {
				user.log("removing tracks from user")
				// for _, sender := range senders {
				for _, receiver := range user.pc.GetReceivers() {
					if receiver.Track() == nil {
						continue
					}
					ssrc := receiver.Track().SSRC()
					
					roomUserSenders := roomUser.pc.GetSenders()
					for _, roomUserSender := range roomUserSenders {
						// fmt.Printf("%v vs %v\n", ssrc, roomUserSender.Track().SSRC())
						if roomUserSender.Track().SSRC() == ssrc {
							err := roomUser.pc.RemoveTrack(roomUserSender)
							if err != nil {
								panic(err)
							}
							roomUser.outTracksLock.Lock()
							delete(roomUser.outTracks, ssrc)
							roomUser.outTracksLock.Unlock()
						}
					}
				}
			}

			// user.pc.RemoveTrack()

		}
	})

	user.pc.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		user.log(
			"peerConnection.OnTrack",
			fmt.Sprintf("track has started, of type %d: %s, ssrc: %d \n", remoteTrack.PayloadType(), remoteTrack.Codec().Name, remoteTrack.SSRC()),
		)
		if _, alreadyAdded := user.inTracks[remoteTrack.SSRC()]; alreadyAdded {
			// user.log("user.inTrack != nil", "already handled")
			return
		}

		user.inTracks[remoteTrack.SSRC()] = remoteTrack
		for _, roomUser := range user.room.GetOtherUsers(user) {
			log.Println("add remote track", fmt.Sprintf("(user: %s)", user.ID), "track to user ", roomUser.ID)
			if err := roomUser.AddTrack(remoteTrack.SSRC()); err != nil {
				log.Println(err)
				continue
			}
			err := roomUser.SendOffer()
			if err != nil {
				panic(err)
			}
		}
		go user.receiveInTrackRTP(remoteTrack)
		go user.broadcastIncomingRTP()
	})

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go user.writePump()
	go user.readPump()
	// go user.Watch()

	user.SendEventUser()
	user.SendEventRoom()
}
