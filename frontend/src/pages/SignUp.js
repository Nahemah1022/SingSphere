// SignUp.js
import React, { useState } from 'react';
import AWS from 'aws-sdk';
import S3Uploader from 'react-s3-uploader';
import './SignUp.css'; // Include your CSS file for styling

const SignUp = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [profilePicture, setProfilePicture] = useState('');

  const handleUploadFinish = (info) => {
    // This function is called when the profile picture upload finishes
    console.log('Upload finished:', info);
  };

  const handleSignUp = async () => {
    if (password !== confirmPassword) {
      console.error("Passwords don't match");
      return;
    }

    const cognito = new AWS.CognitoIdentityServiceProvider();

    const params = {
      ClientId: 'YOUR_COGNITO_CLIENT_ID',
      Username: username,
      Password: password,
      UserAttributes: [
        {
          Name: 'email',
          Value: username,
        },
      ],
    };

    try {
      const data = await cognito.signUp(params).promise();
      console.log('Successfully signed up:', data);
    } catch (error) {
      console.error('Error signing up:', error);
    }
  };

  return (
    <div className="signup-container">
      <h1>Sign Up</h1>
      <div className="input-container">
        <label htmlFor="username">Username:</label>
        <input
          type="text"
          id="username"
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>
      <div className="input-container">
        <label htmlFor="password">Password:</label>
        <input
          type="password"
          id="password"
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>
      <div className="input-container">
        <label htmlFor="confirmPassword">Confirm Password:</label>
        <input
          type="password"
          id="confirmPassword"
          onChange={(e) => setConfirmPassword(e.target.value)}
        />
      </div>
      <div className="input-container">
        <label htmlFor="profilePicture">Profile Picture:</label>
        <S3Uploader
          signingUrl="/s3/sign"
          signingUrlMethod="GET"
          accept="image/*"
          onFinish={handleUploadFinish}
          uploadRequestHeaders={{ 'x-amz-acl': 'public-read' }}
          contentDisposition="auto"
          scrubFilename={(filename) => filename.replace(/[^\w\d_\-.]+/gi, '')}
        />
      </div>
      <button onClick={handleSignUp}>Sign Up</button>
    </div>
  );
};

export default SignUp;