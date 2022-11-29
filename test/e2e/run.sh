#!/bin/bash

FAILURE=0

for FILE in "$(dirname "${BASH_SOURCE[0]}")"/tests/*;
do
  echo "Running ${FILE##*/}";
  ./"${FILE}"
  EXIT_CODE=$?
  if [ $EXIT_CODE -ne 0 ];
  then
    FAILURE=1
    echo "${FILE##*/} failed."
  else
    echo "${FILE##*/} succeeded."
  fi
done

if [ $FAILURE -eq 1 ];
then
  exit 1
fi

