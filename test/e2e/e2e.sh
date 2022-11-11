#!/bin/bash

RESPONSE=$(curl -s -X POST "localhost:30000/schema.v1.Service/ResolveBoolean" -d '{"flagKey":"simple-flag","context":{}}' -H "Content-Type: application/json")
EXPECTED_RESPONSE='{"value":true,"reason":"DEFAULT","variant":"on"}'

# let's tear down regardless of the result of the curl
make undeploy ignore-not-found=true

if [ "$RESPONSE" == "$EXPECTED_RESPONSE" ]
then
  echo "Success."
  exit 0
else
  echo "Unexpected response: $RESPONSE"
  exit 1
fi
