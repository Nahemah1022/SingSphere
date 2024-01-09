/*
This manager.go implements functions for the room package to create rooms
or provide statstical metadata of all rooms across the application.
*/
package room

import (
	"errors"

	"github.com/Nahemah1022/singsphere-voice-server/pkg/socket"
	"github.com/Nahemah1022/singsphere-voice-server/user"
)

type RoomsStats struct {
	Online int                `json:"online"`
	Rooms  []*socket.RoomWrap `json:"rooms"`
}

type RoomManager struct {
	rooms map[string]*Room
}

var ErrNotFound = errors.New("not found")

// Get a room or create one if it does not exist
func (rm *RoomManager) GetOrCreate(name string) *Room {
	if room, exist := rm.rooms[name]; exist {
		return room
	}
	newRoom := &Room{
		Name:        name,
		users:       make(map[string]*user.User),
		UserJoinCh:  make(chan *user.User),
		UserLeaveCh: make(chan *user.User),
	}
	rm.rooms[name] = newRoom
	go newRoom.run()
	return newRoom
}

// Get a room if exists, return error instead
func (rm *RoomManager) Get(name string) (*Room, error) {
	if room, exist := rm.rooms[name]; exist {
		return room, nil
	}
	return nil, ErrNotFound
}

// Get statistical metadata of all rooms
func (rm *RoomManager) GetStats() RoomsStats {
	stats := RoomsStats{
		Rooms: []*socket.RoomWrap{},
	}
	for _, r := range rm.rooms {
		if len(r.users) == 0 {
			continue
		}
		stats.Online += len(r.users)
		stats.Rooms = append(stats.Rooms, r.Wrap())
	}
	return stats
}

func (r *Room) Wrap() *socket.RoomWrap {
	usersWrap := []*socket.UserWrap{}
	for _, user := range r.users {
		usersWrap = append(usersWrap, user.Wrap())
	}
	return &socket.RoomWrap{
		Users:   usersWrap,
		Name:    r.Name,
		Online:  len(r.users),
		Playing: nil,
	}
}

// Instanciate a room manager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room, 100),
	}
}
