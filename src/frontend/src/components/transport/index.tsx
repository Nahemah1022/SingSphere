import { TransportEvent } from '../../api/api'

interface Transport {
    sendOffer: (sessionDescription: RTCSessionDescriptionInit) => void;
    sendAnswer: (sessionDescription: RTCSessionDescriptionInit) => void;
    sendCandidate: (candidate: RTCIceCandidateInit) => void;
    sendEvent: (event: TransportEvent) => void;

    onOpen: (callback: () => void) => void;
    onOffer: (
      callback: (sessionDescription: RTCSessionDescriptionInit) => void
    ) => void;
    onAnswer: (
      callback: (sessionDescription: RTCSessionDescriptionInit) => void
    ) => void;
    onCandidate: (callback: (candidate: RTCIceCandidateInit) => void) => void;
}

export default class WebSocketTransport implements Transport {
    private ws: WebSocket;
    private onOfferCallback: (
        sessionDescription: RTCSessionDescriptionInit
    ) => void;
    private onAnswerCallback: (
        sessionDescription: RTCSessionDescriptionInit
    ) => void;
    private onCandidateCallback: (candidate: RTCIceCandidateInit) => void;
    private onOpenCallback: () => void;
    private onEventCallback: (event: TransportEvent) => void;
    constructor(path: string) {
        this.onOfferCallback = () => undefined;
        this.onAnswerCallback = () => undefined;
        this.onCandidateCallback = () => undefined;
        this.onOpenCallback = () => undefined;
        this.onEventCallback = () => undefined;
        this.ws = new WebSocket(path);
        this.ws.addEventListener("message", (event) => this.onMessage(event));
        this.ws.addEventListener("open", () => this.onOpenCallback());
        this.ws.addEventListener("close", () => console.log("ws is closed"));
        this.ws.addEventListener("error", (error) => console.error(error));
    }
    public sendOffer(sessionDescription: RTCSessionDescriptionInit): void {
        this.sendEvent({ type: "offer", offer: sessionDescription });
    }
    public sendOfferStereo(sessionDescription: RTCSessionDescriptionInit): void {
        this.sendEvent({ type: "offer_stereo", offer: sessionDescription });
    }
    public sendAnswer(sessionDescription: RTCSessionDescriptionInit): void {
        this.sendEvent({ type: "answer", answer: sessionDescription });
    }
    public sendCandidate(candidate: RTCIceCandidateInit) {
        this.sendEvent({ type: "candidate", candidate });
    }
    public sendEvent(event: TransportEvent) {
        console.log("[transport]sendEvent", event.type);
        this.ws.send(JSON.stringify(event));
    }

    private onMessage(event: MessageEvent) {
        const data = JSON.parse(event.data) as TransportEvent;

        if (data.type === "answer" && data.answer) {
            return this.onAnswerCallback(data.answer);
        } else if (data.type === "offer" && data.offer) {
            return this.onOfferCallback(data.offer);
        } else if (data.type === "candidate" && data.candidate) {
            return this.onCandidateCallback(data.candidate);
        } else if (data.type === "error") {
            console.error(data);
        } else {
            this.onEventCallback(data);
        }
    }

    public onOpen(callback: () => void): void {
        this.onOpenCallback = callback;
    }
    public onOffer(callback: WebSocketTransport["onOfferCallback"]): void {
        this.onOfferCallback = callback;
    }
    public onAnswer(callback: WebSocketTransport["onAnswerCallback"]): void {
        this.onAnswerCallback = callback;
    }
    public onCandidate(callback: WebSocketTransport["onCandidateCallback"]): void {
        this.onCandidateCallback = callback;
    }
    public onEvent(callback: WebSocketTransport["onEventCallback"]): void {
        this.onEventCallback = callback;
    }
}