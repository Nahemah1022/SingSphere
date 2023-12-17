import ssl
import pika
import json
import base64
import requests
import boto3
import logging

HOST = 'b-dd7ec1e7-2096-4e6a-9dec-f8a6dc939959.mq.us-east-1.amazonaws.com'
USERNAME = 'singsphere'
PASSWORD = 'singsphere123'
REGION = 'us-east-1'
PORT = 5671
EXCHANGE_NAME = 'songs_exchange'
EXCHANGE_TYPE = 'direct'
BUCKET = 'final-music'
ROOM_STATS_API = 'https://sinsphere-api.nahemah.com/api/stats'

# Set up logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)


def lambda_handler(event, context):
    logger.info("Enter enqueue-music, event: %s", event)
    decrypted_body = json.loads(base64.b64decode(event['body']))
    logger.info('Decrypted body: %s', decrypted_body)

    song = decrypted_body['song']
    room = decrypted_body['room']

    song_names = list_song() 
    room_names = list_room()

    if not room_names:
        logger.error("No rooms are currently available")
        return build_response(404, {"message": "No rooms available, please create a room first."})
    
    if not room:
        logger.error("Empty room received")
        return build_response(400, {"message": "Room cannot be empty, please try again."})
    
    if room not in room_names:
        logger.error("Room: %s is not available", room)
        return build_response(404, {"message": f"No room found matching the name '{room}', please try again."})
    
    if not song_names:
        logger.error("No available songs")
        return build_response(400, {"message": "No songs available, please index a song first."})
    
    if not song:
        logger.error("Empty song received")
        return build_response(400, {"message": "Song cannot be empty, please try again."})
    
    if song not in song_names:
        logger.error("Song: %s is not available", song)
        return build_response(404, {"message": f"No song found matching the name '{song}'. Please try again."})

    connection = connect()
    
    try:
        message = send_message(connection, room, song)
        logger.info(message)
        return build_response(200, {"message": "Message sent successfully", "result": message})
    except Exception as e:
        logger.error("Error in sending message: %s", e)
        return build_response(500, {"message": "Failed to send message due to server error. Please contact development team."})


# list name of all rooms
def list_room():
    logger.info('Enter list_room')
    response = requests.get(ROOM_STATS_API)
    
    if response.status_code == 200:
        data = response.json()
        room_names = [room['name'] for room in data.get('rooms', [])]
        if room_names:
            logger.info('Rooms: %s', room_names)
            return room_names
        logger.error('Room data is present but the room list is empty, data: %s', data)
        return []
    
    logger.error('Failed to retrieve room stats, status code: %s, data: %s', response.status_code, data)
    return []

# list name of all songs from s3 without OpenSearch info
def list_song():
    logger.info('Enter list_song')

    s3_client = boto3.client("s3")
    objects = s3_client.list_objects_v2(Bucket=BUCKET)
    song_names = [obj['Key'] for obj in objects.get('Contents', [])]
    
    if song_names:
        logger.info('Songs: %s', song_names)
        return song_names
    
    return []


# connect to RabbitMQ
def connect():
    logger.info("Enter connect")
    
    ssl_context = ssl.create_default_context(cafile=None, capath=None, cadata=None)
    ssl_context.check_hostname = False
    ssl_context.verify_mode = ssl.CERT_NONE

    credentials = pika.PlainCredentials(USERNAME, PASSWORD)
    parameters = pika.ConnectionParameters(
        host=HOST, 
        port=PORT, 
        credentials=credentials, 
        ssl_options=pika.SSLOptions(ssl_context)
    )

    try:
        connection = pika.BlockingConnection(parameters)
        logger.info("Established connection successfully")
        return connection
    except Exception as err:
        logger.error("Failed to connect to RabbitMQ, error: %s", err)
        raise


# send message to RabbitMQ exchange based on room 
def send_message(connection, room, song):
    logger.info("Enter send_message")

    try:
        channel = connection.channel()
        channel.exchange_declare(exchange=EXCHANGE_NAME, exchange_type=EXCHANGE_TYPE)
        channel.basic_publish(
            exchange = EXCHANGE_NAME,
            routing_key = room,
            body = song,
            properties=pika.BasicProperties(delivery_mode=2) # 2 stands for Persistent message delivery
        )
        logger.info("Published message: routing_key: %s, body: %s", room, song)
        return f"Song '{song}' successfully published to room '{room}'."
    except pika.exceptions.AMQPError as err:
        logger.error("Failed to publish message, error : %s", err)
        raise 
    finally:
        if connection and not connection.is_closed:
            connection.close()
            logger.info("Connection closed")


# util function to build http response
def build_response(status_code, result):

    return {
        'statusCode': status_code,
        'headers': {
            'Access-Control-Allow-Origin':'*',
            'Access-Control-Allow-Credentials':True,
            'Access-Control-Allow-Methods': 'OPTIONS,POST,GET',
            'Access-Control-Request-Headers':'*',
            'Access-Control-Allow-Headers':'*'
        },
        'body': json.dumps({"results": result})
    }