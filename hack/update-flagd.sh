#!/bin/bash
FLAGD_VERSION=$1
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sed -i "s/INPUT_FLAGD_VERSION/${FLAGD_VERSION}/g" config/manager/manager.yaml
    sed -i "s/INPUT_FLAGD_VERSION/${FLAGD_VERSION}/g" chart/template/deployment.yaml
elif [[ "$OSTYPE" == "darwin"* ]]; then
	sed -i '' -e "s/INPUT_FLAGD_VERSION/${FLAGD_VERSION}/g" config/manager/manager.yaml
    sed -i '' -e "s/INPUT_FLAGD_VERSION/${FLAGD_VERSION}/g" chart/template/deployment.yaml
fi
