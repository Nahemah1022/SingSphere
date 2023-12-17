import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';
import logo from '../assets/logo.png';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import ListItemIcon from '@mui/material/ListItemIcon';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import PersonIcon from '@mui/icons-material/Person';

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
				<Paper className="scroll" style = {{backgroundColor: 'transparent'}}>
					<List className="room-list">
						{ktvrooms.map((room) => (
							<React.Fragment key={room.name}>
							<ListItem disablePadding className="list-item">
								<ListItemButton component={Link} to={`/${room.name}`}>
								<ListItemText className="list-text" primary={`Room ${room.name}`} />
								<ListItemIcon>
									<PersonIcon className="list-text" />
									<ListItemText className="list-text" primary={`${room.online}`} />
								</ListItemIcon>
								</ListItemButton>
							</ListItem>
							<Divider />
							</React.Fragment>
						))}
					</List>
				</Paper>
			</div>
    	</div>
	</div>
  );
}

export default Rooms;