import React, { useRef, useContext } from "react";
import { useAudioContext } from "./audio";

const MediaStreamManagerContext = React.createContext<MediaStreamManager | undefined>(undefined);
const MediaStreamManagerProvider: React.FC = ({ children }) => {
    const audioContext = useAudioContext();
    const refMediaStreamManager = useRef<MediaStreamManager>();
    if (!refMediaStreamManager.current) {
        refMediaStreamManager.current = new MediaStreamManager(audioContext);
    }
    return (
        <MediaStreamManagerContext.Provider value={refMediaStreamManager.current}>
        {children}
        </MediaStreamManagerContext.Provider>
    );
};
export const useMediaStreamManager = (): MediaStreamManager => {
    const mediaStreamManager = useContext(MediaStreamManagerContext);
    if (!mediaStreamManager) {
        throw new Error("Media stream manager is not connected");
    }
    return mediaStreamManager;
};

export class MediaStreamManager {
    public audioContext: AudioContext;
    public inputGain: GainNode;
    public outputGain: GainNode;

    private inputStreamDestination: MediaStreamAudioDestinationNode;
    private outputStreamDestination: MediaStreamAudioDestinationNode;

    private microphone: MediaStreamAudioSourceNode | undefined;
    private microphoneGain: GainNode | undefined;

    // private oscillator: OscillatorNode;
    // private oscillatorGain: GainNode;

    public isMicrophoneRequested: boolean;

    constructor(audioContext: AudioContext) {
        this.isMicrophoneRequested = false;
        this.audioContext = audioContext;

        // this.oscillator = this.audioContext.createOscillator();
        // this.oscillatorGain = this.audioContext.createGain();
        // this.disableOscillator();

        this.inputGain = this.audioContext.createGain();
        this.outputGain = this.audioContext.createGain();

        // this.oscillator.connect(this.oscillatorGain);
        // this.oscillatorGain.connect(this.inputGain);

        // this.oscillator.detune.value = 100;
        // this.oscillator.frequency.value = sample([200, 250, 300, 350, 400, 450, 500, 550,]);
        // this.oscillator.start(0);

        this.inputStreamDestination = this.audioContext.createMediaStreamDestination();
        this.inputGain.connect(this.inputStreamDestination);
        this.inputGain.gain.value = 1;

        this.outputStreamDestination = this.audioContext.createMediaStreamDestination();
        this.outputGain.connect(this.outputStreamDestination);
        this.outputGain.gain.value = 1;
    }

    public getInputStream(): MediaStream {
        return this.inputStreamDestination.stream;
    }
    public getOutputStream(): MediaStream {
        return this.outputStreamDestination.stream;
    }

    public async requestMicrophone(): Promise<void> {
        try {
        this.isMicrophoneRequested = true;
        const mediaStream = await navigator.mediaDevices.getUserMedia({
            audio: true,
        });
        this.microphone = this.audioContext.createMediaStreamSource(mediaStream);
        this.microphoneGain = this.audioContext.createGain();
        this.microphoneGain.gain.value = 0; // mute by default
        this.microphone.connect(this.microphoneGain);
        this.microphoneGain.connect(this.inputGain);
        } catch (error) {
        this.isMicrophoneRequested = false;
        return undefined;
        }
    }

    public addOutputTrack(stream: MediaStream) {
        // const outputStreamSource = this.audioContext.createMediaStreamSource(
        //   stream
        // );
        // const outputStreamGain = this.audioContext.createGain();
        // outputStreamGain.gain.value = 0.5;
        // outputStreamSource.connect(outputStreamGain);
        // outputStreamSource.connect(this.audioContext.destination);
        // outputStreamGain.connect(this.outputGain);
        // outputStreamSource.connect(this.outputGain);

        const audio = new Audio();
        audio.srcObject = stream;
        const gainNode = this.audioContext.createGain();
        gainNode.gain.value = 0.5;
        audio.onloadedmetadata = () => {
        const source = this.audioContext.createMediaStreamSource(
            audio.srcObject as MediaStream
        );
        audio.play();
        audio.muted = true;
        source.connect(gainNode);
        gainNode.connect(this.outputGain);
        };
    }

    mute() {
        this.inputGain.gain.value = 0;
    }
    unmute() {
        this.inputGain.gain.value = 1;
    }

    get isMicrophoneMuted(): boolean {
        if (!this.microphoneGain) {
        throw new Error("Microphone is not connected");
        }
        return this.microphoneGain.gain.value === 0;
    }

    microphoneMute(): void {
        if (!this.microphoneGain) {
        throw new Error("Microphone is not connected");
        }
        this.microphoneGain.gain.value = 0;
    }
    microphoneUnmute(): void {
        if (!this.microphoneGain) {
        throw new Error("Microphone is not connected");
        }
        this.microphoneGain.gain.value = 1;
    }

    enableOscillator() {
        // this.oscillatorGain.gain.value = 1;
    }
    disableOscillator() {
        // this.oscillatorGain.gain.value = 0;
    }
}

export default MediaStreamManagerProvider;
