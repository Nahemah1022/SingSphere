import * as React from "react";
import { useRef, useState, useEffect } from "react";


import css from "./VoiceChat.module.css";
import { UserMe, UsersRemoteList, EmptyRoom } from "./Components";
import { useStore, User, StoreProvider } from "./api";
import AudioContextProvider, { useAudioContext } from './context/audio';
import MediaStreamManagerProvider, { useMediaStreamManager } from './context/mediastream';
import WebSocketTransport from './transport';

const Conference = () => {
    const audioContext = useAudioContext();
    const mediaStreamManager = useMediaStreamManager();

    const refAudioEl = useRef<HTMLAudioElement | null>(null);
    const store = useStore();
    const { state, update } = store;

    const [user, setUser] = useState<User>();
    const refTransport = useRef<WebSocketTransport>();
    const WS_URL = `ws://sinsphere-api.nahemah.com:8000/${window.location.pathname.replace("/", "")}`
    // const WS_URL = `ws://127.0.0.1:8000/${window.location.pathname.replace("/", "")}`
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
            console.log(event.streams);
            const stream = event.streams[0];
            try {
                const audio = document.createElement("audio");
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

export const VoiceChat = () => {
    const refContainer = useRef<HTMLDivElement>(null);
    useEffect(() => {
        const set100vh = () => {
            if (refContainer.current) {
                refContainer.current.style.height = `${window.innerHeight}px`;
            }
        };
        window.addEventListener("resize", set100vh);
        set100vh();
        return () => {
            window.removeEventListener("resize", set100vh);
        };
    }, []);
    return (
        <AudioContextProvider>
        <MediaStreamManagerProvider>
            <StoreProvider>
            <div className={css.container} ref={refContainer}>
                <ErrorBoundary>
                    <Conference />
                </ErrorBoundary>
            </div>
            </StoreProvider>
        </MediaStreamManagerProvider>
        </AudioContextProvider>
    );
};

interface ErrorBoundaryProps {}
class ErrorBoundary extends React.Component<
    ErrorBoundaryProps,
    {
        errorMessage: string | undefined;
    }
> {
    constructor(props: ErrorBoundaryProps) {
        super(props);
        this.state = { errorMessage: undefined };
    }
    static getDerivedStateFromError(error: Error) {
        // Update state so the next render will show the fallback UI.
        console.log("getDerivedStateFromError", error);
        return { errorMessage: error.toString() };
    }
    componentDidCatch(error: Error, info: any) {
        console.log("error here", error, info);
    }
    render() {
        if (this.state.errorMessage) {
            return <div>err: {this.state.errorMessage}</div>;
        }
        return this.props.children;
    }
}
