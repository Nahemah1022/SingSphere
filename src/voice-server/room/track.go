package room

import (
	"log"

	"github.com/Nahemah1022/singsphere-voice-server/user"
)

// acceptRoomTracks adds all user's mic track in this room to the given user
func (r *Room) acceptRoomTracks(u *user.User) error {
	for _, roomUser := range r.users {
		micTrack, err := roomUser.GetMicTrack()
		if err != nil {
			log.Println(err)
			return err
		}
		if err := u.AcceptMicTrack(micTrack); err != nil {
			log.Println("ERROR Add remote track", err)
			return err
		}
	}
	return nil
}

// attachMicTrack adds the given user's mic track to all users in this room
func (r *Room) attachMicTrack(u *user.User) error {
	<-u.MicReadyCtx.Done()
	micTrack, err := u.GetMicTrack()
	if err != nil {
		return err
	}
	for _, roomUser := range r.users {
		if err := roomUser.AcceptMicTrack(micTrack); err != nil {
			log.Println("ERROR Add remote track", err)
			return err
		}
	}
	return nil
}
