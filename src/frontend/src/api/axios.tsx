import React, { useState } from 'react';
import axios, {AxiosRequestConfig, AxiosResponse} from 'axios';
import {sign} from 'aws4';
//import { aws4Interceptor } from "aws4-axios";
import AWSCredentials from '../aws_creds_local';

interface Song {
  url: string;
  name: string;
  labels: string[];
}

// Only for POST because GET doens't have CORS set
const client = axios.create();

//const interceptor = aws4Interceptor({
//  options: {
//    region: "us-east-1",
//	service: "execute-api",
//	assumeRoleArn: "arn:aws:iam::601912694676:user/Josephine"
//  },
//  credentials: AWSCredentials
//});

//client.interceptors.request.use(interceptor);

const SearchPage = () => {
  const [songName, setSongName] = useState('');
  const [searchResults, setSearchResults] = useState<Song[]>([]);

  const handleSearch = async () => {
    try {
      const response = await axios.get(
        `https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/GET/?song=${songName}`
      );
	  console.log(response);

      setSearchResults(response.data.results);
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  // Explicitly define the song parameter type
  const queueSong = async (song: Song) => {
    try {
		const path = `/Dev/POST/final-music`;
		const url = `https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/POST/final-music`;

		const requestBody = {
			song: song.name,
			room: '123',
		};

		// Convert requestBody to Base64
		const base64RequestBody = Buffer.from(JSON.stringify(requestBody)).toString('base64');

		// Create a request object
		const request = {
			host: 'zuooeb1uui.execute-api.us-east-1.amazonaws.com',
			method: 'POST',
			url: url,
			data: base64RequestBody,
			headers: {
				'content-type': 'application/json'
			}
		};

		// Sign the request
		const signedRequest = sign(
			request,
			AWSCredentials
		);

		if (signedRequest.headers) {
			delete signedRequest.headers['Host'];
			delete signedRequest.headers['Content-Length'];
		}

		// Use the signed request for the API call
		const response: AxiosResponse = await client(request);

		console.log('Song queued successfully:', response.data);
	} catch (error) {
		console.error('Error queuing song:', error);
	}
  };

  return (
    <div>
      <h1>Search Page</h1>
      <div>
        <label htmlFor="songName">Song Name:</label>
        <input
          type="text"
          id="songName"
          value={songName}
          onChange={(e) => setSongName(e.target.value)}
        />
        <button onClick={handleSearch}>Search</button>
      </div>
      <div>
        <h2>Search Results:</h2>
        <ul>
          {searchResults.map((result, index) => (
            <li key={index}>
              {result.name} - {result.labels.join(', ')}
              <button onClick={() => queueSong(result)}>Queue Song</button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default SearchPage;