import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';
import logo from '../assets/logo.png';
import Button from '@mui/material/Button';

interface KTVroom {
  name: string;
  online: number;
}

function Rooms() {
  const navigate = useNavigate();
  const [ktvrooms, setKTVrooms] = useState<KTVroom[]>([]);

  const fetchKTVrooms = async () => {
    try {
      const response = await fetch('https://sinsphere-api.nahemah.com/api/stats');
      const data: { rooms: KTVroom[] } = await response.json();

      if (data.rooms && Array.isArray(data.rooms)) {
        setKTVrooms(data.rooms);
      }
    } catch (error) {
      console.error('Error fetching ktvrooms:', error);
    }
  };

  const handleNewRoom = async () => {
    const newKTVroomId = uuidv4();
    navigate(`/${newKTVroomId}`);

    // Fetch ktv rooms after creating a new room
    await fetchKTVrooms();
  };

  const handleEnterCode = async () => {

  };

  // Fetch ktv rooms when mounting
  useEffect(() => {
    fetchKTVrooms();
  }, []);

  return (
    <div className="bg-container">
		<div className="rooms-background">
			<img className='home-logo' src={logo} alt="Logo" />
			<div className="room-list-container">
				<div className="button-container">
					<button className="new-room-button" onClick={handleNewRoom}>Create Room</button>
					<button className="enter-code-button" onClick={handleNewRoom}>Enter Room Code</button>
				</div>
				<div className="rooms-caption">Join a room to start singing!</div>
				<Button variant="contained" color="primary">
					Test
				</Button>
				<ul>
					{ktvrooms.map((room) => (
					<li key={room.name}>
						<Link to={`/${room.name}`}>{`KTVroom ${room.name} (${room.online} online)`}</Link>
					</li>
					))}
				</ul>
			</div>
    	</div>
	</div>
  );
}

export default Rooms;