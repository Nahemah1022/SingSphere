package room

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/mq"
	"github.com/Nahemah1022/singsphere-voice-server/pkg/socket"
	"github.com/Nahemah1022/singsphere-voice-server/stream"
	"github.com/Nahemah1022/singsphere-voice-server/user"
)

type Room struct {
	Name          string
	users         map[string]*user.User
	userLock      sync.RWMutex
	UserJoinCh    chan *user.User
	UserLeaveCh   chan *user.User
	SongRequestCh chan string
	mqConsumer    *mq.Consumer
	audioHub      *stream.AudioHub
}

var (
	ErrUserAlreadyJoined = errors.New("user already joined this room")
	ErrUserNotExist      = errors.New("user not in this room")
)

// broadcast broadcasts event to all users in this room except the given user
func (r *Room) broadcast(event *socket.OutboundEvent, u *user.User) error {
	r.userLock.Lock()
	for _, roomUser := range r.users {
		if u == roomUser {
			continue
		}
		roomUser.SendEvent(event)
	}
	r.userLock.Unlock()
	return nil
}

// join joins the given user to this room
func (r *Room) join(u *user.User) error {
	if _, exist := r.users[u.ID]; exist {
		return ErrUserAlreadyJoined
	}
	u.SendEvent(&socket.OutboundEvent{
		EventBase: socket.EventBase{Type: "room"},
		Room:      r.Wrap(),
	})
	r.broadcast(&socket.OutboundEvent{
		EventBase: socket.EventBase{Type: "user_join", Desc: fmt.Sprintf("user %s joined this room", u.ID)},
		User:      u.Wrap(),
	}, nil)
	r.userLock.Lock()
	if err := r.acceptRoomTracks(u); err != nil {
		log.Println(err)
	}
	r.users[u.ID] = u
	r.userLock.Unlock()
	go r.attachMicTrack(u)
	log.Println("New user joined room:", r.Name)
	return nil
}

// leave removes the given user from this room
func (r *Room) leave(u *user.User) error {
	if _, exist := r.users[u.ID]; !exist {
		return ErrUserNotExist
	}
	r.userLock.Lock()
	delete(r.users, u.ID)
	r.userLock.Unlock()
	r.broadcast(&socket.OutboundEvent{
		EventBase: socket.EventBase{Type: "user_leave", Desc: fmt.Sprintf("user %s left this room", u.ID)},
		User:      u.Wrap(),
	}, nil)
	go r.removeMicTrack(u)
	log.Println("user leave room:", r.Name)
	return nil
}

func (r *Room) run() {
	for {
		select {
		case u := <-r.UserJoinCh:
			if err := r.join(u); err != nil {
				log.Println(err)
			}
		case u := <-r.UserLeaveCh:
			if err := r.leave(u); err != nil {
				log.Println(err)
			}
		case song := <-r.SongRequestCh:
			log.Println("consume song ", song)
		}
	}
}
