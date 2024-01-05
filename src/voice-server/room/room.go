package room

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/socket"
	"github.com/Nahemah1022/singsphere-voice-server/user"
)

type Room struct {
	ID          string
	users       map[string]*user.User
	UserJoinCh  chan *user.User
	UserLeaveCh chan *user.User
	userLock    sync.RWMutex
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
	r.userLock.Lock()
	r.users[u.ID] = u
	r.userLock.Unlock()
	r.broadcast(&socket.OutboundEvent{
		EventBase: socket.EventBase{Type: "join", Desc: fmt.Sprintf("user %s joined this room", u.ID)},
	}, nil)
	log.Println("New user joined room:", r.ID)
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
		EventBase: socket.EventBase{Type: "leave", Desc: fmt.Sprintf("user %s left this room", u.ID)},
	}, nil)
	log.Println("user leave room:", r.ID)
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
		}
	}
}
