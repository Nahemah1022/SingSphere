package user

import (
	"log"

	"github.com/pion/webrtc/v3"
)

func (u *User) PeerConnect() error {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return err
	}
	u.pc = peerConnection
	log.Println("PC connected")
	return nil
}
