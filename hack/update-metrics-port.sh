#!/bin/bash
METRICS_PORT=$1
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sed -i "s/INPUT_METRICS_PORT/${METRICS_PORT}/g" config/manager/kustomization.yaml
elif [[ "$OSTYPE" == "darwin"* ]]; then
	sed -i '' -e "s/INPUT_METRICS_PORT/${METRICS_PORT}/g" config/manager/kustomization.yaml
fi
