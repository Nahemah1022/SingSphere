package rtc

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// RtcNode contains a peer connection instance and a set of communication channel associated with it
type RtcNode struct {
	pc                 *webrtc.PeerConnection
	ICEConnectedCtx    context.Context
	ICEDisconnectedCtx context.Context
	MicTrack           *webrtc.TrackRemote
	MicReadyCtx        context.Context
	SignalChs          *SignalChannels
	RtpCh              chan *rtp.Packet // Collect all rtp packets from senders' track
	SendersLock        sync.RWMutex
	Senders            map[webrtc.SSRC]*webrtc.RTPSender // all incoming tracks' senders
}

// SignalChannels are a set of channels used for establishing WebRTC client-server connection
type SignalChannels struct {
	AnswerCh    chan webrtc.SessionDescription
	CandidateCh chan *webrtc.ICECandidate
}

// New creates RtcNode instance
func New() (*RtcNode, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	ICEConnectedCtx, ICEConnectedCancel := context.WithCancel(context.TODO())
	ICEDisconnectedCtx, ICEDisonnectedCancel := context.WithCancel(context.TODO())
	MicReadyCtx, MicReadyCtxCancel := context.WithCancel(context.TODO())
	node := &RtcNode{
		pc:                 peerConnection,
		ICEConnectedCtx:    ICEConnectedCtx,
		ICEDisconnectedCtx: ICEDisconnectedCtx,
		MicTrack:           nil,
		MicReadyCtx:        MicReadyCtx,
		SignalChs: &SignalChannels{
			AnswerCh:    make(chan webrtc.SessionDescription),
			CandidateCh: make(chan *webrtc.ICECandidate),
		},
		RtpCh:   make(chan *rtp.Packet, 100),
		Senders: make(map[webrtc.SSRC]*webrtc.RTPSender),
	}

	peerConnection.OnICECandidate(func(iceCandidate *webrtc.ICECandidate) {
		if iceCandidate != nil {
			node.SignalChs.CandidateCh <- iceCandidate
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			ICEConnectedCancel()
		} else if connectionState == webrtc.ICEConnectionStateDisconnected ||
			connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateClosed {
			ICEDisonnectedCancel()
		}
	})

	// Invoked when the remote client attach its mic track to peer connection
	peerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		log.Println(
			"peerConnection.OnTrack",
			fmt.Sprintf("client mic track has started, of type %d, ssrc: %d \n", tr.PayloadType(), tr.SSRC()),
		)
		node.MicTrack = tr
		MicReadyCtxCancel()
		go node.receiveTrackRTP(tr)
	})
	log.Println("PC connected")
	return node, nil
}

// HandleOffer sends an answer back abd create a transceiver for the connection
func (node *RtcNode) HandleOffer(offer webrtc.SessionDescription) error {
	if len(node.pc.GetTransceivers()) == 0 {
		// add receive only transciever to get user microphone audio
		_, err := node.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		})
		if err != nil {
			return err
		}
	}

	// Set the remote SessionDescription
	if err := node.pc.SetRemoteDescription(offer); err != nil {
		return err
	}

	answer, err := node.Answer()
	if err != nil {
		return err
	}
	node.SignalChs.AnswerCh <- answer

	return nil
}

// Offer return a offer
func (node *RtcNode) Offer() (webrtc.SessionDescription, error) {
	offer, err := node.pc.CreateOffer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	err = node.pc.SetLocalDescription(offer)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	return offer, nil
}

// Answer creates webrtc answer
func (node *RtcNode) Answer() (webrtc.SessionDescription, error) {
	answer, err := node.pc.CreateAnswer(nil)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	if err = node.pc.SetLocalDescription(answer); err != nil {
		return webrtc.SessionDescription{}, err
	}
	return answer, nil
}

func (node *RtcNode) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	if err := node.pc.AddICECandidate(candidate); err != nil {
		return err
	}
	return nil
}

func (node *RtcNode) AcceptAnswer(answer webrtc.SessionDescription) error {
	// time.Sleep(3000 * time.Millisecond)
	if err := node.pc.SetRemoteDescription(answer); err != nil {
		return err
	}
	return nil
}

func (node *RtcNode) Ternimate() error {
	if err := node.pc.Close(); err != nil {
		return err
	}
	return nil
}
