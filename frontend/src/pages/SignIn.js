// SignIn.js
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import AWS from 'aws-sdk';

const SignIn = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const signIn = async () => {
    const cognito = new AWS.CognitoIdentityServiceProvider();

    const params = {
      AuthFlow: 'USER_PASSWORD_AUTH',
      ClientId: 'YOUR_COGNITO_CLIENT_ID',
      AuthParameters: {
        USERNAME: username,
        PASSWORD: password,
      },
    };

    try {
      const data = await cognito.initiateAuth(params).promise();
      console.log('Successfully signed in', data);
      // Redirect or perform actions after successful sign-in
    } catch (error) {
      console.error('Error signing in:', error);
    }
  };

  return (
    <div className="signin-container">
      <h1>Sign In</h1>
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
      <button onClick={signIn}>Sign In</button>
      <div className="signup-link">
        Don't have an account yet? <Link to="/signup">Sign up</Link>
      </div>
    </div>
  );
};

export default SignIn;
