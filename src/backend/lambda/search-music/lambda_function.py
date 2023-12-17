import json
import boto3
import requests
import logging
from requests_aws4auth import AWS4Auth
from opensearchpy import OpenSearch, RequestsHttpConnection

REGION = 'us-east-1'
SERVICE = 'es'
HOST = 'search-music-266fztddlp63e4bo5tdfhebwvu.aos.us-east-1.on.aws'
BUCKET = 'final-music'
PORT = 443
INDEX = 'music'
ALL = 'all'

# Set up logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

def lambda_handler(event, context):
    logger.info('Enter search-music, event: %s', event)
    search_term = event.get('queryStringParameters', {}).get('song', '').lower()
    
    s3_client = boto3.client("s3")

    if not search_term:
        logger.error("Empty seearch term received")
        return build_response(400, {"message": "Search term cannot be empty."})
    
    if search_term == ALL:
        song_list = list_all_songs(s3_client)
        if not song_list:
            logger.info("No available songs in s3.")
            return build_response(404, {"message": "No available songs, please contact the development."})
        return build_response(200, {"message": "Retrieved all available songs.", "results": song_list})
    
    try:
        open_search_results = search_songs_info(search_term)
        song_list = get_songs(s3_client, open_search_results)

        if not song_list:
            logger.info("Songs related to %s are not available", search_term)
            return build_response(404, {"message": f"No songs found matching the search term '{search_term}'. Please try a different search."})
        
        return build_response(200, {"message": f"Retrievd song based on the search term '{search_term}'.", "results": song_list})

    except Exception as err:
        logger.error("Failed to search the song, error: %s", err)
        return build_response(500, {"message": "Failed to search the song, please contact development team."})


# list all songs from s3 without OpenSearch info
def list_all_songs(s3_client):
    logger.info('Enter list_all_songs')

    objects = s3_client.list_objects_v2(Bucket=BUCKET)
    song_list = []

    for obj in objects.get('Contents', []):
        meta_data = s3_client.head_object(Bucket=BUCKET, Key=obj['Key']).get('Metadata', {})
        list_url = s3_client.generate_presigned_url(ClientMethod='get_object', Params={'Bucket': BUCKET, 'Key': obj['Key']}, ExpiresIn=36000)
        info = {'url': list_url, 'search_term': obj['Key'], 'labels': meta_data.get('customlabels', '').split(',')}
        song_list.append(info)
    
    return song_list


# get songs from s3 based on the OpenSearch response
def get_songs(s3_client, open_search_results):
    logger.info("Enter get_songs")
    song_list = []

    for res in open_search_results:
        for hit in res['hits']['hits']:
            key = hit['_source']['objectKey']
            labels = hit['_source'].get('labels', "")
            url = s3_client.generate_presigned_url(ClientMethod="get_object", Params={"Bucket": BUCKET, "Key": key}, ExpiresIn=36000) 
            song_list.append({"url": url, "search_term": key, "labels": labels[0] if labels else ""})
    
    return song_list


# search songs info from OpenSearch based on search term
def search_songs_info(search_term):
    logger.info("Enter search_songs_info")

    open_search_client = OpenSearch(
        hosts=[{'host': HOST, 'port': PORT}],
        http_auth=get_aws_auth(),
        use_ssl=True,
        verify_certs=True,
        connection_class=RequestsHttpConnection
    )
    query = build_query(search_term)

    result = open_search_client.search(index=INDEX, body=query)
    return [result]


# build a query for OpenSearch
def build_query(search_term):
    if '.mp3' in search_term:
        return {"size": 5, "query": {"match": {"_id": search_term}}}
    else:
        return {"size": 5, "query": {'multi_match': {'query': search_term}}}
    

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
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Credentials': True,
            'Access-Control-Request-Headers':'*',
            'Access-Control-Allow-Headers':'*'
        },
        'body': json.dumps(body)
    }
