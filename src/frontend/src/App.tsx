import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import "./styles.css";
import { VoiceChat } from "./VoiceChat";
import Homepage from './Homepage';
import Rooms from './Rooms';

export default function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Homepage />} />
        <Route path="/rooms" element={<Rooms />} />
        <Route path="/:id" element={<VoiceChat />} />
      </Routes>
    </Router>
  );
}
