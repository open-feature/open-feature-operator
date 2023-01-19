#!/bin/bash
FLAGD_VERSION=$1
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sed -i "s/INPUT_FLAGD_VERSION/${FLAGD_VERSION}/g" apis/core/v1alpha1/flagsourceconfiguration_types.go
elif [[ "$OSTYPE" == "darwin"* ]]; then
	sed -i '' -e "s/INPUT_FLAGD_VERSION/${FLAGD_VERSION}/g" apis/core/v1alpha1/flagsourceconfiguration_types.go
fi
