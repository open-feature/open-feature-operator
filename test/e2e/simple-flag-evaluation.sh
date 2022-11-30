#!/bin/bash

EXPECTED_RESPONSE="$1"
# attempt up to 5 times
MAX_ATTEMPTS=5
# retry every x seconds
RETRY_INTERVAL=1

for (( ATTEMPT_COUNTER=0; ATTEMPT_COUNTER<=${MAX_ATTEMPTS}; ATTEMPT_COUNTER++ ))
do 

    RESPONSE=$(curl -s -X POST "localhost:30000/schema.v1.Service/ResolveBoolean" -d '{"flagKey":"simple-flag","context":{}}' -H "Content-Type: application/json")
    RESPONSE="${RESPONSE//[[:space:]]/}" # strip whitespace from response
   
    if [ "$RESPONSE" == "$EXPECTED_RESPONSE" ]
    then
      exit 0
    fi

    echo "Expected response: $EXPECTED_RESPONSE"
    echo "Got response: $RESPONSE"
    echo "Retrying in ${RETRY_INTERVAL} seconds"
    sleep ${RETRY_INTERVAL}
done
echo "Max attempts reached"
exit 1