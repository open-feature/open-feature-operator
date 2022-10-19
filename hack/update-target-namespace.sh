#!/bin/bash
TARGET_NAMESPACE=$1
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sed -i "s/INPUT_TARGET_NAMESPACE/${TARGET_NAMESPACE}/g" config/manager/kustomization.yaml
elif [[ "$OSTYPE" == "darwin"* ]]; then
	sed -i '' -e "s/INPUT_TARGET_NAMESPACE/${TARGET_NAMESPACE}/g" config/manager/kustomization.yaml
fi
