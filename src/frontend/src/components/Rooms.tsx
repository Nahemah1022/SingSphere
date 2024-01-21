import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import logo from '../assets/logo.png';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import ListItemIcon from '@mui/material/ListItemIcon';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import Modal from '@mui/material/Modal';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import PersonIcon from '@mui/icons-material/Person';

interface KTVroom {
  name: string;
  online: number;
}

function Rooms() {
  const navigate = useNavigate();
  const [ktvrooms, setKTVrooms] = useState<KTVroom[]>([]);
  const [open, setOpen] = React.useState(false);
  const [roomCode, setRoomCode] = useState('');
  const handleOpen = () => setOpen(true);
  const handleClose = () => setOpen(false);

  const fetchKTVrooms = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/stats');
      const data: { rooms: KTVroom[] } = await response.json();

      if (data.rooms && Array.isArray(data.rooms)) {
        setKTVrooms(data.rooms);
      }
    } catch (error) {
      console.error('Error fetching ktvrooms:', error);
    }
  };

  const handleNewRoom = async () => {
    const newKTVroomId = Math.random().toString(36).slice(2, 6);
    navigate(`/${newKTVroomId}`);

    // Fetch ktv rooms after creating a new room
    await fetchKTVrooms();
  };

  const handleEnterRoom = async () => {
    try {
      // Fetch data from the API
      const response = await fetch('http://localhost:8080/api/stats');
      const data = await response.json();

      // Check if the entered code matches any room name
      const isValidCode = data.rooms.some((room: KTVroom) => room.name === roomCode);

      if (isValidCode) {
        navigate(`/${roomCode}`);
        setOpen(false);
      } else {
        alert('Invalid room code. Please try again.');
      }
    } catch (error) {
      console.error('Error fetching or validating room code:', error);
      alert('An error occurred. Please try again.');
    }
  };

  // Fetch ktv rooms when mounting
  useEffect(() => {
    fetchKTVrooms();
  }, []);

  return (
    <div className='bg-container'>
      <div className='rooms-background'>
        <img className='home-logo' src={logo} alt='Logo' />
        <div className='room-list-container'>
          <div className='button-container'>
            <button className='new-room-button' onClick={handleNewRoom}>
              Create Room
            </button>
            <button className='enter-code-button' onClick={handleOpen}>
              Enter Room Code
            </button>
          </div>
          <Modal
            open={open}
            onClose={handleClose}
            aria-labelledby='modal-modal-title'
            aria-describedby='modal-modal-description'
          >
            <Box className='modal'>
              <div className='modal-text'>Please input your room code</div>
              <TextField
                variant='filled'
                value={roomCode}
                onChange={(e) => setRoomCode(e.target.value)}
                className='modal-input'
              />
              <button className='modal-button' onClick={handleEnterRoom}>
                Join Room
              </button>
            </Box>
          </Modal>
          <div className='rooms-caption'>Join a room to start singing!</div>
          <Paper className='scroll' style={{ backgroundColor: 'transparent', boxShadow: 'none' }}>
            <List>
              {ktvrooms.map((room) => (
                <React.Fragment key={room.name}>
                  <ListItem disablePadding className='list-item'>
                    <ListItemButton component={Link} to={`/${room.name}`}>
                      <ListItemText className='list-text' primary={`Room ${room.name}`} />
                      <ListItemIcon>
                        <PersonIcon className='list-text' />
                        <ListItemText className='list-text' primary={`${room.online}`} />
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
