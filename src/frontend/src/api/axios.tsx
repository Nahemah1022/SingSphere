import React, { useState } from 'react';
import axios, { AxiosRequestConfig, AxiosResponse, Method} from 'axios';
//import { sign} from 'aws4';
import {aws4Interceptor} from 'aws4-axios';
import AWSCredentials from '../aws_creds_local';

interface Song {
  url: string;
  search_term: string;
  labels: string[];
}

//interface SignedRequest {
//  method: Method;
//  //service: string;
//  region: string;
//  host: string;
//  headers: Record<string, string>;
//  body: string;
//}

const client = axios.create();

const interceptor = aws4Interceptor({
  options: {
    region: "us-east-1",
	service: "execute-api",
//	assumeRoleArn: "arn:aws:iam::601912694676:user/Josephine"
  },
//  credentials: AWSCredentials,
});

client.interceptors.request.use(interceptor);

const SearchPage = () => {
  const [songName, setSongName] = useState('');
  const [songs, setSongs] = useState<Song[]>([]);

  const handleSearch = async () => {
    try {
      const response = await axios.get(
        `https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/GET/?song=${songName}`
      );
      console.log(response);

      setSongs(response.data.results);
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  const queueSong = async (result: Song) => {
    try {
      const path = '/Dev/POST/final-music';
      const url = 'https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/Dev/POST/final-music';
	  //const url = 'https://zuooeb1uui.execute-api.us-east-1.amazonaws.com/';

      const requestBody = {
        song: result.search_term,
        room: '123',
      };

      const base64RequestBody = Buffer.from(JSON.stringify(requestBody)).toString('base64');

	  const response = await axios({
		method: 'POST',
		url: url,
		data: base64RequestBody,
		headers: {
			'Content-Type': 'text/plain'
		}
	  })

    //  const request = {
    //    //host: 'zuooeb1uui.execute-api.us-east-1.amazonaws.com',
	//	host,
    //    method: 'POST',
	//	region: 'us-east-1',
    //    body: base64RequestBody,
    //    headers: {
    //      'content-type': 'application/json',
    //    },
    //  };

	//  let signedRequest: AxiosRequestConfig = {}
    //  signedRequest.data = sign(request, AWSCredentials);

	  //  if (signedRequest.headers) {
		//	delete signedRequest.headers['Host'];
		//    delete signedRequest.headers['Content-Length'];  }

    //  // Use the axios request for the API call
    //  let response: AxiosResponse = await axios(signedRequest);

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
          {songs.map((result, index) => (
            <li key={index}>
              {result.search_term} - {result.labels.join(', ')}
              <button onClick={() => queueSong(result)}>Queue Song</button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default SearchPage;