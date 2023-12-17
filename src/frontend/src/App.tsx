import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import "./assets/styles.css";
import { VoiceChat } from "./components/VoiceChat";
import Homepage from './components/Homepage';
import Rooms from './components/Rooms';
import SearchPage from './api/axios';

export default function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Homepage />} />
        <Route path="/rooms" element={<Rooms />} />
        <Route path="/:id" element={<VoiceChat />} />
		<Route path="/search" element={<SearchPage />} />
      </Routes>
    </Router>
  );
}
