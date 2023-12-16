import * as React from "react";
import { useRef, useEffect } from "react";
import { useParams } from 'react-router-dom';
import css from "./VoiceChat.module.css";
import { ErrorBoundary } from "./Components";
import { StoreProvider } from "../api/api";
import AudioContextProvider from './context/audio';
import MediaStreamManagerProvider from './context/mediastream';
import Conference from './Conference';

export const VoiceChat = () => {
	const { id } = useParams<{id: string | undefined}>();
	const roomId = id || "";
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
                   {roomId && <Conference roomId={roomId}/>}
                </ErrorBoundary>
            </div>
            </StoreProvider>
        </MediaStreamManagerProvider>
        </AudioContextProvider>
    );
};
