# Submodule: GoLang Audio Streaming Service

## Overview
This project describes a GoLang-based music streaming service, leveraging an EC2 instance, WebRTC protocol, and WebSocket for communication. The service is designed to establish voice communication channels for clients in the same room, while stream music to clients in different rooms via a message queue.

## System Components
1. **EC2 Instance**: Hosts the streaming service.
2. **WebSocket**: Facilitates real-time communication between clients and the server.
3. **WebRTC Protocol**: Handles streaming of binary audio data.
4. **Streaming Server**: Core module for establishing peer connections.
5. **Room Manager**: Helper module to manage peer connections based on room IDs.
6. **Media Manager**: Module to process audio requests and stream audio from S3.

## Workflow
### 1. User Connection
- Users connect to the EC2 server through WebSocket when they join a room.
- This connection enables the Streaming Server to establish peer connections.

### 2. Peer Connection Management
- WebRTC peer connections are established via "pion" package.
- The Room Manager partitions these connections based on room IDs.

### 3. Audio Streaming
- The Media Manager listens to the message queue for audio requests.
- Upon receiving a request, it locates the corresponding audio file in the S3 bucket (mounted directly on the EC2 as storage).
- The Media Manager then streams the audio to clients in the appropriate room.

## Modules
### Streaming Server
- Manages peer connections via WebRTC.
- Responsible for establishing and maintaining streaming channels.

### Room Manager
- Allocates and manages peer connections based on room IDs.
- Ensures that clients in a room only receive audio from that specific room.

### Media Manager
- Subscribes to the message queue for song requests.
- Fetches audio files from the S3 bucket.
- Streams audio to clients based on their room ID.

## Getting Started
1. **Set Up EC2 Instance**: Deploy an EC2 instance with necessary network configurations.
2. **S3 Bucket Integration**: Mount the S3 bucket onto the EC2 for direct file access through by executing `s3-mount.sh` file.
3. **Deploy Modules**: Simply run `docker-compose up -d` and you are all set!!

---

## TODO
[] Separate the music streaminig module out, performing its function as a normal room user
[] Testing
