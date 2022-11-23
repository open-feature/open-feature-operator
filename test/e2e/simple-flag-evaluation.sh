#!/bin/bash

EXPECTED_RESPONSE="$1"

# attempt up to 5 times
ATTEMPT_COUNTER=0
MAX_ATTEMPTS=5
until RESPONSE=$(curl -s -X POST "localhost:30000/schema.v1.Service/ResolveBoolean" -d '{"flagKey":"simple-flag","context":{}}' -H "Content-Type: application/json"); do
    if [ ${ATTEMPT_COUNTER} -eq ${MAX_ATTEMPTS} ];then
      echo "Max attempts reached"
      exit 1
    fi

    printf '.'
    ATTEMPT_COUNTER=$((ATTEMPT_COUNTER+1))
    sleep 1
done

RESPONSE="${RESPONSE//[[:space:]]/}" # strip whitespace from response

if [ "$RESPONSE" == "$EXPECTED_RESPONSE" ]
then
  exit 0
else
  echo "Expected response: $EXPECTED_RESPONSE"
  echo "Got response: $RESPONSE"
  exit 1
fi
