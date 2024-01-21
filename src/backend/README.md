# Submodule: Music Upload and Search System

## Overview
This system provides a cloud-based solution for uploading, indexing, searching, and requesting audio files using a combination of AWS services. It consists of three main APIs behind an AWS API Gateway: Upload API, Search API, and Enqueue Music API.

## Components
***Make sure these services are enabled on the AWS account***
- **AWS API Gateway**: Serves as the entry point for all APIs, ensuring secure and efficient handling of requests.
- **S3 Bucket**: A storage service for audio files uploaded by users.
- **AWS Lambda**: Functions for indexing uploaded audio files in OpenSearch.
- **OpenSearch Service**: Used for indexing and searching audio files.
- **RabbitMQ on AWS**: A message broker service for handling song requests using a publish/subscribe model.

## APIs
### 1. Upload API
- **Functionality**: Users can upload audio files along with custom labels.
- **Process**:
  - The audio file is uploaded to a designated S3 bucket.
  - A Lambda function triggers to index the file and its metadata in OpenSearch.

### 2. Search API
- **Functionality**: Retrieve audio files based on search queries.
- **Process**:
  - Users send search queries through the API.
  - The system searches the indexed files in OpenSearch and returns relevant results.

### 3. Enqueue Music API
- **Functionality**: Request songs from search results to be played.
- **Process**:
  - Users select a song from search results.
  - The song's ID is published to RabbitMQ under a specific topic corresponding to the user's room.

## Usage Guidelines
### Uploading Audio Files
1. Send a POST request to the Upload API endpoint with the audio file and custom labels. (Ensure the audio file is in OPUS format for compatibility with the WebRTC protocol by voice server)
2. The file will be stored in S3 and indexed in OpenSearch.

### Searching for Audio Files
1. Use the Search API to submit search queries.
2. Review the returned list of existing audio files.

### Requesting Songs
1. Choose a song from the search results.
2. Send a request to the Enqueue Music API to add the song to a queue.
3. Specify the topic (e.g., user's room) for song playback.
