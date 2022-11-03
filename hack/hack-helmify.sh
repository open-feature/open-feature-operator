#!/bin/bash
NAMESPACE=$1
NAMESPACE_REPLACE="{{ .Release.Namespace }}"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sed -i "s/${NAMESPACE}/${NAMESPACE_REPLACE}/g" chart/templates/rendered.yaml
elif [[ "$OSTYPE" == "darwin"* ]]; then
	sed -i '' -e "s/${NAMESPACE}/${NAMESPACE_REPLACE}/g" chart/templates/rendered.yaml
fi
