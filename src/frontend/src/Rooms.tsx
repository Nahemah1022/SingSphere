import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';

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

  const handleNewKTVroomClick = async () => {
    const newKTVroomId = uuidv4();
    navigate(`/${newKTVroomId}`);

    // Fetch ktvrooms after creating a new room
    await fetchKTVrooms();
  };

  // Fetch ktvrooms when mounting
  useEffect(() => {
    fetchKTVrooms();
  }, []);

  return (
    <div>
      <h1>KTV Rooms</h1>
      <button onClick={handleNewKTVroomClick}>New KTVroom</button>
      <ul>
        {ktvrooms.map((room) => (
          <li key={room.name}>
            <Link to={`/${room.name}`}>{`KTVroom ${room.name} (${room.online} online)`}</Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default Rooms;