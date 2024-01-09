package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/rtc"
	"github.com/Nahemah1022/singsphere-voice-server/pkg/socket"
	"github.com/pion/webrtc/v3"
)

type User struct {
	ID                string
	Emoji             string
	Mute              bool
	ws                *socket.Websocket
	rtc               *rtc.RtcNode
	joinCh            chan *User
	leaveCh           chan *User
	MicReadyCtx       context.Context
	micReadyCtxCancel context.CancelFunc
}

var emojis = []string{
	"😎", "🧐", "🤡", "👻", "😷", "🤗", "😏",
	"👽", "👨‍🚀", "🐺", "🐯", "🦁", "🐶", "🐼", "🙈",
}

var errNotImplemented = errors.New("not implemented")

func (u *User) Wrap() *socket.UserWrap {
	return &socket.UserWrap{
		ID:    u.ID,
		Emoji: u.Emoji,
		Mute:  u.Mute,
	}
}

func New(joinCh chan *User, leaveCh chan *User, ws *socket.Websocket, rtcNode *rtc.RtcNode) *User {
	ctx, ctxCancel := context.WithCancel(context.TODO())
	return &User{
		ID:                strconv.FormatInt(time.Now().UnixNano(), 10), // generate random id based on timestamp
		Mute:              true,
		Emoji:             emojis[rand.Intn(len(emojis))],
		joinCh:            joinCh,
		leaveCh:           leaveCh,
		ws:                ws,
		rtc:               rtcNode,
		MicReadyCtx:       ctx,
		micReadyCtxCancel: ctxCancel,
	}
}

func (u *User) Run() {
	defer func() {
		u.rtc.Ternimate()
		u.leaveCh <- u
	}()
	go u.ws.Run()
	go func() {
		// When client's ICE succesfully connected, notify its room through channel
		<-u.rtc.ICEConnectedCtx.Done()
		u.joinCh <- u
	}()
	go func() {
		// When client's mic successfully attach, cancel the context to notify user's room
		<-u.rtc.MicReadyCtx.Done()
		u.micReadyCtxCancel()
	}()
	for {
		select {
		case inboundEvent := <-u.ws.InboundEventCh:
			go u.handleInboundEvent(inboundEvent)
		case answer := <-u.rtc.SignalChs.AnswerCh:
			if err := u.SendAnswer(answer); err != nil {
				u.ws.SendError(errors.New("fail to send answer"))
			}
		case ICEcandidate := <-u.rtc.SignalChs.CandidateCh:
			if err := u.SendCandidate(ICEcandidate); err != nil {
				u.ws.SendError(errors.New("fail to send ICE candidate"))
			}
		case <-u.rtc.ICEDisconnectedCtx.Done():
			return
		}
	}
}

// handleInboundEvent handle the given inbound event based on its type
func (u *User) handleInboundEvent(event *socket.InboundEvent) error {
	if event == nil {
		return errors.New("empty event")
	}
	u.log("handle event: ", event.Type)
	if event.Type == "offer" {
		if event.Offer == nil {
			return u.ws.SendError(errors.New("empty offer"))
		}
		if err := u.rtc.HandleOffer(*event.Offer); err != nil {
			return err
		}
		return nil
	} else if event.Type == "answer" {
		if event.Answer == nil {
			return u.ws.SendError(errors.New("empty answer"))
		}
		if err := u.rtc.AcceptAnswer(*event.Answer); err != nil {
			return u.ws.SendError(errors.New("fail to accept answer"))
		}
		return nil
	} else if event.Type == "candidate" {
		if event.Candidate == nil {
			return u.ws.SendError(errors.New("empty candidate"))
		}
		if err := u.rtc.AddICECandidate(*event.Candidate); err != nil {
			return u.ws.SendError(errors.New("fail to add candidate"))
		}
		return nil
	} else if event.Type == "mute" {
		return nil
	} else if event.Type == "unmute" {
		return nil
	}
	return u.ws.SendError(errNotImplemented)
}

// GetMicTrack return user's mic track or error if haven't attached yet
func (u *User) GetMicTrack() (*webrtc.TrackRemote, error) {
	if u.rtc.MicTrack == nil {
		return nil, errors.New("client mic haven't attached yet")
	}
	return u.rtc.MicTrack, nil
}

// AcceptMicTrack attaches the given track to user's peer connection instance
func (u *User) AcceptMicTrack(tr *webrtc.TrackRemote) error {
	localTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(tr.Codec().RTPCodecCapability, "audio", "mic")

	// keep track of the correspondence for sending remote packets
	u.rtc.ListeningTracksLock.Lock()
	u.rtc.ListeningTracks[tr.SSRC()] = localTrack
	u.rtc.ListeningTracksLock.Unlock()

	if newTrackErr != nil {
		panic(newTrackErr)
	}
	_, err := u.rtc.AddTrack(localTrack)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) log(msg ...interface{}) {
	log.Println(
		fmt.Sprintf("user %s:", u.ID),
		fmt.Sprint(msg...),
	)
}
