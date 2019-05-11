#!/bin/bash

### Kinesis helper functions

DUMMY_REGION=dummyregion
KINESITE_ENDPOINT=http://localhost:4567

kCreateStream() {
    local STREAM_NAME="$1"
    aws kinesis create-stream \
        --region $DUMMY_REGION \
        --endpoint-url $KINESITE_ENDPOINT \
        --shard-count 1 \
        --stream-name "$STREAM_NAME"
}

kListStreams() {
    aws kinesis list-streams \
        --region $DUMMY_REGION \
        --endpoint-url $KINESITE_ENDPOINT
} 


kGetRecords() {
    local STREAM_NAME="$1"
    local SHARD_ITERATOR=$(aws kinesis get-shard-iterator \
        --region $DUMMY_REGION \
        --endpoint-url $KINESITE_ENDPOINT \
        --shard-id shardId-000000000000 \
        --shard-iterator-type TRIM_HORIZON \
        --stream-name "$STREAM_NAME" | jq -r .ShardIterator)
    aws kinesis get-records \
        --region $DUMMY_REGION \
        --endpoint-url $KINESITE_ENDPOINT \
        --shard-iterator "$SHARD_ITERATOR"
}

kGetLastRecord() {
    local STREAM_NAME="$1"
    kGetRecords "$STREAM_NAME" | jq -r .Records[-1].Data | base64 -d
}
