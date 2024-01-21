package room

import (
	"fmt"
	"log"

	"github.com/Nahemah1022/singsphere-voice-server/user"
	"github.com/pion/webrtc/v3"
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
	// To enable the newly attached mic track, we re-send the offer again
	if err := u.SendOffer(); err != nil {
		panic(err)
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
		// skip the user himself
		if u.ID == roomUser.ID {
			continue
		}
		if err := roomUser.AcceptMicTrack(micTrack); err != nil {
			log.Println("ERROR Add remote track", err)
			return err
		}
		// To enable the newly attached mic track, we re-send the offer again
		if err := roomUser.SendOffer(); err != nil {
			panic(err)
		}
	}
	go r.broadcastMicTrack(u, micTrack.SSRC())
	return nil
}

// removeMicTrack removes the given user's mic track from all user's sender in this room
func (r *Room) removeMicTrack(u *user.User) error {
	<-u.MicReadyCtx.Done()
	micTrack, err := u.GetMicTrack()
	if err != nil {
		return err
	}
	for _, roomUser := range r.users {
		// skip the user himself
		if u.ID == roomUser.ID {
			continue
		}
		if err := roomUser.RemoveSender(micTrack.SSRC()); err != nil {
			log.Println("ERROR Remove sender track", err)
			return err
		}
		if err := roomUser.SendOffer(); err != nil {
			panic(err)
		}
	}
	return nil
}

// broadcastMicTrack broadcasts incoming RTP packets from the given user's mic to all room users
func (r *Room) broadcastMicTrack(u *user.User, micTrackSSRC webrtc.SSRC) {
	log.Println("Start Broadcasting")
	for {
		rtp, err := u.ReadRTP()
		if err != nil {
			panic(err)
		}
		for _, roomUser := range r.users {
			// skip the user himself
			if u.ID == roomUser.ID {
				continue
			}
			err := roomUser.WriteRTP(rtp, micTrackSSRC)
			if err != nil {
				// panic(err)
				fmt.Println(err)
			}
		}
	}
}
