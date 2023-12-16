import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';

function Rooms() {
  const navigate = useNavigate();

  const handleNewChatroomClick = () => {
    const newChatroomId = uuidv4();
    navigate(`/${newChatroomId}`);
  };

  return (
    <div>
      <h1>Chatrooms</h1>
      <button onClick={handleNewChatroomClick}>New Chatroom</button>
      <ul>
        <li>
          <Link to="/123">Chatroom 123</Link>
        </li>
        <li>
          <Link to="/456">Chatroom 456</Link>
        </li>
      </ul>
    </div>
  );
}

export default Rooms;