import * as React from "react";
import { useRef, useState, useEffect } from "react";
import css from "./VoiceChat.module.css";
import bggif from "../assets/bggif2.gif"
import { UsersRemoteList, EmptyRoom, ButtonMicrohone, ButtonSpeaker} from "./Components";
import { useStore, User } from "../api/api";
import { useAudioContext } from './context/audio';
import { useMediaStreamManager } from './context/mediastream';
import WebSocketTransport from './transport';
import AppBar from '@mui/material/AppBar';
import MusicNoteIcon from '@mui/icons-material/MusicNote';
import Modal from '@mui/material/Modal';
import Box from '@mui/material/Box';
import axios from 'axios';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Paper from '@mui/material/Paper';
import AddIcon from '@mui/icons-material/Add';
import InputRange from 'react-input-range';
import { ColorRing } from 'react-loader-spinner';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

interface ConferenceProps {
  roomId: string;
}

interface Song {
  url: string;
  search_term: string;
  labels: string[];
}

const Conference = ({ roomId }: ConferenceProps) => {
    const audioContext = useAudioContext();
    const mediaStreamManager = useMediaStreamManager();
	const [user, setUser] = useState<User>();
	const [showConference, setShowConference] = useState<boolean>(false);
    const [volume, setVolume] = useState<number>(5);

	const store = useStore();
    const { state, update } = store;

	// For searching and queueing songs
	const [searchTerm, setSearchTerm] = useState<string>('');
	const [songs, setSongs] = useState<Song[]>([]);
	const [open, setOpen] = useState<boolean>(false);
    const [isSearching, setIsSearching] = useState<boolean>(false);
	const handleOpen = () => setOpen(true);
	const handleClose = () => setOpen(false);

	// For counting down song duration
	const [startCounting, setStartCounting] = useState<boolean>(false);
	const [duration, setDuration] = useState<number>(0);


	// For formatting the result list by removing '_', capitalizing, and removing '.mp3'
	const capitalizeAndFormat = (str: string): string => (
		str
			.replace(/_/g, ' ')
			.replace(/\.mp3$/, '')
			.replace(/\b\w/g, (match) => match.toUpperCase())
	);

	// Format the seconds into MM:SS format
	const formatTime = (timeInSeconds: number): string => {
		const minutes: number = Math.floor(timeInSeconds / 60);
		const seconds: number = timeInSeconds % 60;
		const formattedMinutes: string = minutes < 10 ? `0${minutes}` : minutes.toString();
		const formattedSeconds: string = seconds < 10 ? `0${seconds}` : seconds.toString();
		return `${formattedMinutes}:${formattedSeconds}`;
	};

	useEffect(() => {
		let countdownInterval: NodeJS.Timeout;

		if (startCounting && duration > 0) {
			countdownInterval = setInterval(() => {
				setDuration((prevDuration) => prevDuration - 1);
			}, 1000);
		}

		return () => {
			// Clear the interval when the component unmounts or when startCounting becomes false
			clearInterval(countdownInterval);
		};
  	}, [startCounting]);

	// Transport / Websocket
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
            console.log(event);
            const stream = event.streams[0];
            try {
                const audio = document.createElement("audio");
                if (event.track.label === "stereo_audio") {
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
			// Alert when a song is enqueued
            } else if (event.type === "enqueue") {
				const song = event?.song?.name ?? "";
				if (song) {
					toast(`Enqueued: ${capitalizeAndFormat(song)} 🎸`);
				}

                console.log(event.song);
			// Alert when the next song is coming up
            } else if (event.type === "next_song") {
				const song = event?.song?.name ?? "";
				const duration = event?.song?.duration ?? "";
				if (song) {
					toast(`Next Up: ${capitalizeAndFormat(song)} 🎤`);
				}

				if (duration) {
					setDuration(duration);
					setStartCounting(true);
				}

                console.log(event.song);
            } else {
                throw new Error(`type ${event.type} not implemented`);
            }
        });
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
		//Sent GET request to retrieve songs
		const handleSearch = async (option: string) => {
			try {
				let formattedTerm = ""
				if (option == 'song'){
					// Change all the spaces to underscore and make all characters lowercase
					formattedTerm = searchTerm.replace(/ /g, '_').toLowerCase();

					// Add .mp3 if there's no mp3 extension on the search term
					if (!formattedTerm.endsWith('.mp3')) {
						formattedTerm += '.mp3';
					}
				}
				else {
					formattedTerm = searchTerm;
				}

                setIsSearching(true);
				const response = await axios.get(`https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/GET/?song=${formattedTerm}`);
				console.log(response);
                setIsSearching(false);
				setSongs(response.data.results);
			} catch (error) {
				setSongs([]);
                setIsSearching(false);
				console.error('Error fetching data:', error);
			}
		};

		// Queue song by sending POST request to lambda function
		const queueSong = async (result: Song) => {
			try {
			const url = 'https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/POST/final-music';

			const requestBody = {
				song: result.search_term,
				room: roomId,
			};

			const base64RequestBody = Buffer.from(JSON.stringify(requestBody)).toString('base64');
            console.log(base64RequestBody)
            let res = await axios.post("https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/final-music", base64RequestBody)
            console.log(res)

			// console.log('Song queued successfully:', response.data);
			} catch (error) {
			console.error('Error queuing song:', error);
			}
		};

        return (
            <div className={css.wrapper}>
				<div className={css.countDown}>🪩 Time Remaining: {formatTime(duration)}</div>
				<ToastContainer />
                <div className={css.userContainer}>{renderUsers()}</div>
                <div className={css.displayContainer}>

					<img src={bggif} width="100%" height="100%" />
					<AppBar style={{
						height: '5em',
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
						<div className={css.buttons}>
							<div className={css.mutebuttons}>
								<ButtonMicrohone
									muted={state.isMutedMicrophone}
									onClick={async (event) => {
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
								<ButtonSpeaker
									muted={state.isMutedSpeaker}
									onClick={() => {
										try {
										update({ isMutedSpeaker: !state.isMutedSpeaker });
										} catch (error) {
										alert(error);
										}
									}}
								/>
							</div>
							<button className={css.addSong} onClick={handleOpen}>
								<MusicNoteIcon style={{ color: 'white'}} />
							</button>
							<div className={css.volumeContainer}>
								<InputRange
									classNames={{
										activeTrack: "input-range__track input-range__track--active",
										disabledInputRange:"input-range input-range--disabled",
										inputRange:"input-range",
										labelContainer:"input-range__label-container",
										maxLabel:"input-range__label input-range__label--max",
										minLabel:"input-range__label input-range__label--min",
										slider:"input-range__slider",
										sliderContainer:"input-range__slider-container",
										track:"input-range__track input-range__track--background",
										valueLabel:"input-range__label input-range__label--value",
									}}
									maxValue={10}
									minValue={0}
									value={volume}
									onChange={(value) => {
										setVolume(value as any)
										let stereoAudio = document.querySelectorAll(".stereo_audio") as NodeListOf<HTMLAudioElement>;
										if (stereoAudio.length !== 0) {
											stereoAudio[0].volume = value as any / 10;
										}
										console.log(value)
									}}
								/>
							</div>
						</div>
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
									id="searchTerm"
									value={searchTerm}
									onChange={(e) => setSearchTerm(e.target.value)}
									className={css.searchInput}
								/>
								<button className={css.searchArtist} onClick={() => handleSearch("artist")}>Search Artist</button>
								<button className={css.searchSong} onClick={() => handleSearch("song")}>Search Song</button>
								<button className={css.searchCategory} onClick={() => handleSearch("category")}>Search Category</button>
							</div>
                            <ColorRing
                                visible={isSearching}
                                height="80"
                                width="80"
                                ariaLabel="blocks-loading"
                                wrapperStyle={{}}
                                wrapperClass="blocks-wrapper"
                                colors={['#e15b64', '#f47e60', '#f8b26a', '#abbd81', '#849b87']}
                                />
							<div className={css.searchResults}>
								{songs.length === 0 ? (
									<p className={css.searchMsg}>No search results</p>
								) : (
									<div>
										<p className={css.searchMsg}>
											{songs.length === 1
											? `${songs.length} song was found`
											: `${songs.length} songs were found`}
										</p>
										<Paper className="scroll" style = {{backgroundColor: 'transparent', color: 'white', boxShadow: 'none'}}>
											<List>
												{songs.map((result,index) => (
													<ListItem key={index} className={css.resultItems}>
													<ListItemText primary={capitalizeAndFormat(result.search_term)} />
													<ListItemText primary={capitalizeAndFormat(result.labels[0])} className={css.artist} />
													<button className={css.addButton} onClick={() => queueSong(result)}><AddIcon /></button>
													</ListItem>
												))}
											</List>
										</Paper>
									</div>
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