/*
This manager.go implements functions for the room package to create rooms
or provide statstical metadata of all rooms across the application.
*/
package room

import (
	"errors"

	"github.com/Nahemah1022/singsphere-voice-server/streaming"
	"github.com/Nahemah1022/singsphere-voice-server/user"
)

type RoomsStats struct {
	Online int         `json:"online"`
	Rooms  []*RoomWrap `json:"rooms"`
}

// Public representation of a room
type RoomWrap struct {
	ID      string `json:"id"`
	Online  int    `json:"online"`
	Playing *streaming.MusicWrap
}

type RoomManager struct {
	rooms map[string]*Room
}

var ErrNotFound = errors.New("not found")

// Get a room or create one if it does not exist
func (rm *RoomManager) GetOrCreate(id string) *Room {
	if room, exist := rm.rooms[id]; exist {
		return room
	}
	newRoom := &Room{
		ID:          id,
		users:       make(map[string]*user.User),
		UserJoinCh:  make(chan *user.User),
		UserLeaveCh: make(chan *user.User),
	}
	rm.rooms[id] = newRoom
	go newRoom.run()
	return newRoom
}

// Get a room if exists, return error instead
func (rm *RoomManager) Get(id string) (*Room, error) {
	if room, exist := rm.rooms[id]; exist {
		return room, nil
	}
	return nil, ErrNotFound
}

// Get statistical metadata of all rooms
func (rm *RoomManager) GetStats() RoomsStats {
	stats := RoomsStats{
		Rooms: []*RoomWrap{},
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

func (r *Room) Wrap() *RoomWrap {
	return &RoomWrap{
		ID:      r.ID,
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
