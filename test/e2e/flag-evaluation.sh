#!/bin/bash

FLAG_KEY="$1"
EXPECTED_RESPONSE="$2"

# attempt up to 5 times
MAX_ATTEMPTS=5
# retry every x seconds
RETRY_INTERVAL=1
if [[ "$3" =~ ^[0-9]+$ ]]
  then
    MAX_ATTEMPTS=$3
fi
if [[ "$4" =~ ^[0-9]+$ ]]
  then
    RETRY_INTERVAL=$4
fi


for (( ATTEMPT_COUNTER=0; ATTEMPT_COUNTER<${MAX_ATTEMPTS}; ATTEMPT_COUNTER++ ))
do 

    RESPONSE=$(curl -s -X POST "localhost:30000/schema.v1.Service/ResolveBoolean" -d "{\"flagKey\":\"$FLAG_KEY\",\"context\":{}}" -H "Content-Type: application/json")
    RESPONSE="${RESPONSE//[[:space:]]/}" # strip whitespace from response
   
    if [ "$RESPONSE" == "$EXPECTED_RESPONSE" ]
    then
      exit 0
    fi

    echo "Expected response for flag $FLAG_KEY: $EXPECTED_RESPONSE"
    echo "Got response for flag $FLAG_KEY: $RESPONSE"
    echo "Retrying in ${RETRY_INTERVAL} seconds"
    sleep "${RETRY_INTERVAL}"
done
echo "Max attempts reached"
exit 1