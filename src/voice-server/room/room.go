package room

import (
	"errors"
	"log"
	"sync"

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

// Join the given user to this room
func (r *Room) join(u *user.User) error {
	if _, exist := r.users[u.ID]; exist {
		return ErrUserAlreadyJoined
	}
	r.userLock.Lock()
	r.users[u.ID] = u
	r.userLock.Unlock()
	log.Println("New user joined room:", r.ID)
	return nil
}

// Remove the given user from this room
func (r *Room) leave(u *user.User) error {
	if _, exist := r.users[u.ID]; !exist {
		return ErrUserNotExist
	}
	r.userLock.Lock()
	delete(r.users, u.ID)
	r.userLock.Unlock()
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
