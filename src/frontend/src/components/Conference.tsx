import * as React from "react";
import { useRef, useState, useEffect } from "react";
import css from "./VoiceChat.module.css";
import { UserMe, UsersRemoteList, EmptyRoom} from "./Components";
import { useStore, User } from "../api/api";
import { useAudioContext } from './context/audio';
import { useMediaStreamManager } from './context/mediastream';
import WebSocketTransport from './transport';
import AppBar from '@mui/material/AppBar';
import MusicNoteIcon from '@mui/icons-material/MusicNote';
import Modal from '@mui/material/Modal';
import Box from '@mui/material/Box';
import axios from 'axios';
import {aws4Interceptor} from 'aws4-axios';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import TextField from '@mui/material/TextField';
import AddIcon from '@mui/icons-material/Add';
import SearchIcon from '@mui/icons-material/Search';

interface ConferenceProps {
  roomId: string;
}

interface Song {
  url: string;
  search_term: string;
  labels: string[];
}

const client = axios.create();

const interceptor = aws4Interceptor({
	options: {
		region: "us-east-1",
		service: "execute-api",
		assumeRoleArn: "arn:aws:iam::601912694676:user/Josephine"
	},
	credentials: {
		accessKeyId: process.env.AWS_ACCESS_KEY_ID ?? "",
		secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY ?? ""
	}
});

client.interceptors.request.use(interceptor);

const Conference = ({ roomId }: ConferenceProps) => {
    const audioContext = useAudioContext();
    const mediaStreamManager = useMediaStreamManager();
	const [user, setUser] = useState<User>();
	const [showConference, setShowConference] = useState<boolean>(false);

	const store = useStore();
    const { state, update } = store;

	// For searching and queueing songs
	const [songName, setSongName] = useState('');
	const [songs, setSongs] = useState<Song[]>([]);
	const [open, setOpen] = useState(false);
	const handleOpen = () => setOpen(true);
	const handleClose = () => setOpen(false);

    const refAudioEl = useRef<HTMLAudioElement | null>(null);

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

		//For searching and queueing songs
		const handleSearch = async () => {
			try {
			const response = await axios.get(
				`https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/GET/?song=${songName}`
			);
			console.log(response);

			setSongs(response.data.results);
			}
			catch (error) {
				console.error('Error fetching data:', error);
			}
		};

		const queueSong = async (result: Song) => {
			try {
			const url = 'https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/POST/final-music';

			const requestBody = {
				song: result.search_term,
				room: roomId,
			};

			const base64RequestBody = Buffer.from(JSON.stringify(requestBody)).toString('base64');
            let res = await axios.post("https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/final-music", base64RequestBody)
            console.log(res)

			// const response = await client({
			// 	method: 'POST',
			// 	url: url,
			// 	data: base64RequestBody,
			// 	headers: {
			// 		'Content-Type': 'text/plain'
			// 	}
			// })

			// console.log('Song queued successfully:', response.data);
			} catch (error) {
			console.error('Error queuing song:', error);
			}
		};

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
					<AppBar style={{
						height: '4em',
						backgroundColor: '#1F2937',
						boxShadow: 'none',
						outline: 'none',
						border: 'none',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						top: 'auto',
						bottom: '0',
						flexGrow: '1'
					}}>
						<button className={css.addSong} onClick={handleOpen}>
							<MusicNoteIcon style={{ color: 'white'}} />
						</button>
					</AppBar>

					<Modal
					open={open}
					onClose={handleClose}
					aria-labelledby="modal-modal-title"
					aria-describedby="modal-modal-description"
					>
						<Box className={css.roomModal}>
							<div className={css.searchBar}>
							<input
								type="text"
								id="songName"
								value={songName}
								onChange={(e) => setSongName(e.target.value)}
								className={css.searchInput}
							/>
							{/*
							<TextField
								label="Search"
								variant="outlined"
								fullWidth
								onChange={(e) => setSongName(e.target.value)}
								className={css.searchInput}
							/>
							*/}
							<button className={css.searchIcon} onClick={handleSearch}><SearchIcon /></button>
							</div>

							<div className={css.searchResults}>
							{songs.length === 0 ? (
								<p>No search results</p>
							) : (
								<List>
								{songs.map((result,index) => (
									<ListItem key={index} className={css.resultItems}>
									<ListItemText primary={result.search_term} />
									<ListItemText primary={result.labels[0]}/>
									<button className={css.addButton} onClick={() => queueSong(result)}><AddIcon /></button>
									</ListItem>
								))}
								</List>
							)}
							</div>
						</Box>
					</Modal>
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