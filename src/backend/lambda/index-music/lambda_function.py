import json
import os
import boto3
import logging
from opensearchpy import OpenSearch, RequestsHttpConnection, OpenSearchException
from requests_aws4auth import AWS4Auth

REGION = 'us-east-1'
SERVICE = 'es'
HOST = 'search-music-266fztddlp63e4bo5tdfhebwvu.aos.us-east-1.on.aws'
PORT = 443
INDEX = 'music'
BUCKET = 'final-music' 

# Set up logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)


def lambda_handler(event, context):
    logger.info('Enter index-music, event: %s', event)
    s3_client = boto3.client('s3')
    song_names = list_all_songs(s3_client) 

    try:
        bucket, name, created_timestamp, labels = extract_song_info(s3_client, event)
        name_lowercase = name.lower()

        result = index_song_info(name_lowercase, bucket, created_timestamp, labels)
        logger.info("Indexed song successfully, result: %s", result)

        return build_response(200, {"message": "Indexed song successfully.", "results": result})
    
    except Exception as err:
        logger.error("Failed to index song, error: %s", err)
        return build_response(500, {"message": "Failed to index song due to server error. Please contact development team."})


# list name of all songs from s3 without OpenSearch info
def list_all_songs(s3_client):
    logger.info('Enter list_all_songs')

    objects = s3_client.list_objects_v2(Bucket=BUCKET)
    song_names = [obj['Key'] for obj in objects.get('Contents', [])]

    logger.info("Existing songs: %s", song_names)
    
    return song_names


# extract song info from s3
def extract_song_info(s3_client, event):
    logger.info("Enter extract_song_info")

    record = event['Records'][0]['s3']
    bucket = record['bucket']['name']
    name = record['object']['key']
      
    try:
        head_object = s3_client.head_object(Bucket=bucket, Key=name)
        custom_labels_arr = head_object["Metadata"].get("customlabels", "")
        custom_labels = [label.strip().lower() for label in custom_labels_arr.split(',')] if custom_labels_arr else []
        created_timestamp = head_object["LastModified"].strftime("%Y-%m-%dT%H:%M:%S")
        logger.info("bucket: %s, name: %s, created_timestamp: %s, custom_labels: %s", bucket, name, created_timestamp, custom_labels)
        return bucket, name, created_timestamp, custom_labels
    
    except s3_client.exceptions.NoSuchKey:
        logger.error("S3 object not found: %s", name)
        raise
    
    except Exception as e:
        logger.error("Error extracting song info: %s", e)
        raise


# index song info to OpenSearch
def index_song_info(name, bucket, created_timestamp, labels):
    logger.info("Enter index_song_info")

    open_search_object = {
        "objectKey": name,
        "bucket": bucket,
        "createdTimestamp": created_timestamp,
        "labels": labels
    }

    open_search_client = OpenSearch(
        hosts=[{'host': HOST,'port': PORT}],
        http_auth=get_aws_auth(),
        use_ssl=True,
        verify_certs=True,
        connection_class=RequestsHttpConnection
    )

    logger.info("OpenSearch object: %s", open_search_object)

    try:
        open_search_client.index(index=INDEX, id=open_search_object["objectKey"], body=json.dumps(open_search_object))
        result = open_search_client.get(index=INDEX, id=open_search_object["objectKey"])
        return result
    
    except OpenSearchException as e:
        logger.error("OpenSearch exception occurred: %s", e)
        raise
    
    except Exception as e:
        logger.error("An unexpected error occurred: %s", e)
        raise


# get AWS authentication for OpenSearch
def get_aws_auth():
    credentials = boto3.Session().get_credentials()

    return AWS4Auth(credentials.access_key,
                    credentials.secret_key,
                    REGION,
                    SERVICE,
                    session_token=credentials.token)


# util function to build http response
def build_response(status_code, body):

    return {
        'statusCode': status_code,
        'headers': {
            'Access-Control-Allow-Origin':'*',
            'Access-Control-Allow-Credentials':True,
            'Access-Control-Allow-Methods': 'OPTIONS,POST,GET',
            'Access-Control-Request-Headers':'*',
            'Access-Control-Allow-Headers':'*'
        },
        'body': json.dumps(body)
    }
