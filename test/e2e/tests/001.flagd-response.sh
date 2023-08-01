#!/bin/bash

FAILURE=0

flagKeys=('simple-flag' 'simple-flag-filepath' 'simple-flag-filepath2')
expectedResponseContain=('value":true,"reason":"STATIC","variant":"on"' '"value":true,"reason":"STATIC","variant":"on"' '"value":true,"reason":"STATIC","variant":"on"')

for i in "${!flagKeys[@]}"; do
  ./"$(dirname "${BASH_SOURCE[0]}")"/../flag-evaluation.sh "${flagKeys[$i]}" "${expectedResponseContain[$i]}"
  EXIT_CODE=$?
  if [ $EXIT_CODE -ne 0 ];
    then
      FAILURE=1
  fi
done

if [ $FAILURE -eq 1 ];
then
  exit 1
fi
