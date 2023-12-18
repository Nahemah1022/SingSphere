import React, {useEffect} from 'react';
import { useNavigate } from 'react-router-dom';
import logo from '../assets/logo.png';

const Homepage: React.FC = () => {
  const navigate = useNavigate();

  const handleStartButtonClick = () => {
    navigate('/rooms');
  };

  useEffect(() => {
    const singsphereCaption = document.querySelector('.home-singsphere');
	const caption = document.querySelector('.home-caption');
	const button = document.querySelector('.home-button');
    if (singsphereCaption && caption && button){
		singsphereCaption.classList.add('loaded');
		caption.classList.add('loaded');
		button.classList.add('loaded');
	}
  }, []);

  return (
    <div className="bg-container">
		<div className="home-background">
			<img className='home-logo' src={logo} alt="Logo" />
			<div className='title-container'>
				<h1 className='home-title'>Welcome <br /> Back</h1>
				<div className='home-singsphere'>SingSphere</div>
				<div className='home-caption'>The all-in-one virtual <br /> karaoke app</div>
			</div>

			<button className='home-button' onClick={handleStartButtonClick}>Join</button>
		</div>
	</div>
  );
};

export default Homepage;