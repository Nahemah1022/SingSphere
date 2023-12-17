import * as React from "react";
import { useRef, useState, useEffect } from "react";
import css from "./VoiceChat.module.css";
import { UserMe, UsersRemoteList, EmptyRoom} from "./Components";
import { useStore, User } from "../api/api";
import { useAudioContext } from './context/audio';
import { useMediaStreamManager } from './context/mediastream';
import WebSocketTransport from './transport';

interface ConferenceProps {
  roomId: string;
}

const Conference = ({ roomId }: ConferenceProps) => {
    const audioContext = useAudioContext();
    const mediaStreamManager = useMediaStreamManager();

    const refAudioEl = useRef<HTMLAudioElement | null>(null);
    const store = useStore();
    const { state, update } = store;

    const [user, setUser] = useState<User>();
    const [volume, setVolume] = useState<number>(5);
    const refTransport = useRef<WebSocketTransport>();
	const WS_URL = `wss://sinsphere-api.nahemah.com/${roomId}`;
    console.log(WS_URL)
    if (!refTransport.current) {
        refTransport.current = new WebSocketTransport(WS_URL);
    }
    const transport = refTransport.current;
    const refPeerConnection = useRef<RTCPeerConnection>(
        new RTCPeerConnection({
            iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
        })
    );
    const peerConnection = refPeerConnection.current;

    const subscribe = async () => {
        peerConnection.ontrack = async (event: RTCTrackEvent) => {
            console.log(`peerConnection::ontrack ${event.track.kind}`);
            const stream = event.streams[0];
            try {
                const audio = document.createElement("audio");
                if (event.track.label === "stereo") {
                    audio.volume = volume / 10;
                    audio.classList.add("stereo_audio");
                }
                audio.srcObject = stream;
                audio.autoplay = true;
                audio.play();
                document.body.appendChild(audio);
            } catch (error) {
                alert(error);
                console.error(error);
            }
        };
        peerConnection.onconnectionstatechange = () => {
            console.log(`peerConnection::onIceConnectionStateChange ${peerConnection.iceConnectionState}`);
        };
        peerConnection.onicecandidate = (event: RTCPeerConnectionIceEvent) => {
            if (event.candidate) {
                transport.sendCandidate(event.candidate.toJSON());
            }
        };
        peerConnection.onnegotiationneeded = async (event: Event) => {
            console.log("peerConnection::negotiationneeded", event);
            await peerConnection.setLocalDescription(
                await peerConnection.createOffer()
            );
            if (!peerConnection.localDescription) {
                throw new Error("no local description");
            }
            transport.sendOffer(peerConnection.localDescription);
        };

        const mediaStream = mediaStreamManager.getInputStream();
        const audioTracks = mediaStream.getAudioTracks();
        console.log("[subscribe]: audioTracks", audioTracks);
        for (const track of audioTracks) {
            peerConnection.addTrack(track, mediaStream);
        }
    };

    useEffect(() => {
        console.log("store", store);
        transport.onOpen(() => { console.log("web socket connection is open"); });
        transport.onOffer(async (offer) => {
            await peerConnection.setRemoteDescription(offer);
            const answer = await peerConnection.createAnswer();
            await peerConnection.setLocalDescription(answer);
            transport.sendAnswer(answer);
        });
        transport.onAnswer(async (answer) => {
            await peerConnection.setRemoteDescription(answer);
        });
        transport.onCandidate(async (candidate) => {
            console.log("[local]: adding ice candidate");
            await peerConnection.addIceCandidate(candidate);
        });
        transport.onEvent(async (event) => {
            console.log("EVENT", event);
            if (event.type === "user_join") {
                if (!event.user) {
                    throw new Error("no user");
                }
                store.api.roomUserAdd(event.user);
            } else if (event.type === "user_leave") {
                if (!event.user) {
                    throw new Error("no user");
                }
                store.api.roomUserRemove(event.user);
            } else if (event.type === "user") {
                setUser(event.user);
            } else if (event.type === "room") {
                store.update({ room: event.room });
            } else if (event.type === "mute") {
                if (!event.user) {
                    throw new Error("no user");
                }
                store.api.roomUserUpdate(event.user);
            } else if (event.type === "unmute") {
                if (!event.user) {
                    throw new Error("no user");
                }
                store.api.roomUserUpdate(event.user);
            } else {
                throw new Error(`type ${event.type} not implemented`);
            }
        });
        return () => {

        }
    }, [store, peerConnection, transport]);

    const renderUsers = () => {
        if (state.room.users.length === 0) {
            return <EmptyRoom />;
        }
        return <UsersRemoteList users={store.state.room.users} />;
    };

    const [showConference, setShowConference] = useState<boolean>(false);

    const renderContent = () => {
        if (!showConference) {
            return (
                <div
                    style={{
                        height: "100%",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                    }}
                >
                    <button
                        onClick={async () => {
                            const playOutputTrack = async () => {
                                console.log("playOutputTrack");
                                try {
                                    const outputStream = mediaStreamManager.getOutputStream();
                                    console.log("refAudioEl.current", refAudioEl.current);
                                    if (!refAudioEl.current) {
                                        throw new Error("no audio node");
                                    }
                                    refAudioEl.current.srcObject = outputStream;
                                    refAudioEl.current.autoplay = true;
                                    // refAudioEl.current.controls = true;
                                    await refAudioEl.current.play();
                                } catch (error) {
                                    alert(error);
                                    console.error(error);
                                }
                            };
                            const resumeAudioContext = async () => {
                                console.log("resumeAudioContext", audioContext.state);
                                if (audioContext.state === "suspended") {
                                    console.log("audio context was in suspended state. resuming...");
                                    await audioContext.resume();
                                }
                            };
                            try {
                                await Promise.all([playOutputTrack(), resumeAudioContext()]);
                                setShowConference(true);
                                await subscribe();
                            } catch (error) {
                                alert(error);
                            }
                        }}
                        className={css.buttonJoin}
                    >
                        Join <span role="img" aria-labelledby="phone">📞</span>
                    </button>
                </div>
            );
        }

        return (
            <div className={css.wrapper}>
                <div className={css.top}>{renderUsers()}</div>
                <div className={css.bottom}>
                    {user && (
                        <UserMe
                            volume={volume}
                            setVolume={setVolume}
                            user={user}
                            isMutedMicrophone={state.isMutedMicrophone}
                            isMutedSpeaker={state.isMutedSpeaker}
                            onClickMuteSpeaker={() => {
                                try {
                                update({ isMutedSpeaker: !state.isMutedSpeaker });
                                } catch (error) {
                                alert(error);
                                }
                            }}
                            onClickMuteMicrohone={async (event) => {
                                if (!mediaStreamManager.isMicrophoneRequested) {
                                    await mediaStreamManager.requestMicrophone();
                                }
                                if (mediaStreamManager.isMicrophoneMuted) {
                                    mediaStreamManager.microphoneUnmute();
                                    transport.sendEvent({ type: "unmute", user });
                                    update({ isMutedMicrophone: false });
                                } else {
                                    mediaStreamManager.microphoneMute();
                                    transport.sendEvent({ type: "mute", user });
                                    update({ isMutedMicrophone: true });
                                }
                            }}
                        />
                    )}
                </div>
            </div>
        );
    };
    return (
        <div style={{ width: "100%" }}>
        <audio ref={refAudioEl} />
            {renderContent()}
        </div>
    );
};

export default Conference;