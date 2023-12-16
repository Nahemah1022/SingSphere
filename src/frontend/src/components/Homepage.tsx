import React from 'react';
import { useNavigate } from 'react-router-dom';

const Homepage: React.FC = () => {
  const navigate = useNavigate();

  const handleStartButtonClick = () => {
    // Use navigate to go to the /rooms route
    navigate('/rooms');
  };

  return (
    <div>
      <h1>Welcome to the Homepage</h1>
      <button onClick={handleStartButtonClick}>Start</button>
    </div>
  );
};

export default Homepage;