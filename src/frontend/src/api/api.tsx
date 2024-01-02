import React, { useState } from 'react';
import { useContext } from 'react';

export interface TransportEvent {
  type:
    | 'offer'
    | 'offer_stereo'
    | 'answer'
    | 'candidate'
    | 'error'
    | 'user'
    | 'user_join'
    | 'user_leave'
    | 'room'
    | 'mute'
    | 'unmute'
    | 'enqueue'
    | 'next_song';

  offer?: RTCSessionDescriptionInit;
  answer?: RTCSessionDescriptionInit;
  candidate?: RTCIceCandidateInit;
  user?: User;
  room?: Room;
  song?: Song;
}

export interface Song {
  name: string;
  path: string;
  duration: number;
}

export interface User {
  id: string;
  emoji: string;
  mute: boolean;
}

export interface Room {
  users: User[];
}

export interface State {
  isMutedMicrophone: boolean;
  isMutedSpeaker: boolean;
  user?: User;
  room: Room;
}

interface Api {
  roomUserAdd: (user: User) => void;
  roomUserRemove: (user: User) => void;
  roomUserUpdate: (user: User) => void;
}
interface Store {
  state: State;
  update: (partial: Partial<State>) => void;
  api: Api;
}

const defaultState: State = {
  isMutedMicrophone: true,
  isMutedSpeaker: false,
  room: {
    users: [],
  },
};

const StoreContext = React.createContext<Store | undefined>(undefined);
export const StoreProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, setState] = useState<State>(defaultState);
  const update = (partial: Partial<State>) => setState({ ...state, ...partial });
  const updateRoom = (partial: Partial<Room>): void => {
    return update({ room: { ...state.room, ...partial } });
  };
  const api: Api = {
    roomUserAdd: (user) => {
      updateRoom({ users: [...state.room.users, user] });
    },
    roomUserRemove: (user) => {
      return updateRoom({
        users: state.room.users.filter((roomUser) => roomUser.id !== user.id),
      });
    },
    roomUserUpdate: (user) => {
      return updateRoom({
        users: state.room.users.map((roomUser) => {
          if (user.id === roomUser.id) {
            return { ...roomUser, ...user };
          }
          return roomUser;
        }),
      });
    },
  };
  return <StoreContext.Provider value={{ state, update, api }}>{children}</StoreContext.Provider>;
};
export const useStore = (): Store => {
  const context = useContext(StoreContext);
  return context as Store;
};
