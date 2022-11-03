#!/bin/bash
NAMESPACE=$1
NAMESPACE_REPLACE='namespace: {{ include "chart.namespace" . }}'
# NAMESPACE_DELIMITER=`---`
# NAMESPACE_DELIMITER_ADDITION=`{{ end }}`
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # replace all instances of the default namespace with NAMESPACE_REPLACE
    sed -i "s/namespace: ${NAMESPACE}/${NAMESPACE_REPLACE}/g" config/default/kustomization.yaml
    # add end to wrap the namespace, --- indicates the end of the namespace definition, only add on first match 
    # sed -i "0,/${NAMESPACE_DELIMITER}/s//${NAMESPACE_DELIMITER_ADDITION}\n&/" chart/templates/rendered.yaml
elif [[ "$OSTYPE" == "darwin"* ]]; then
# replace all instances of the default namespace with NAMESPACE_REPLACE
	sed -i '' -e "s/namespace: ${NAMESPACE}/${NAMESPACE_REPLACE}/g" config/default/kustomization.yaml
        # add end to wrap the namespace, --- indicates the end of the namespace definition
    # sed -i '' -e "0,/${NAMESPACE_DELIMITER}/s//${NAMESPACE_DELIMITER_ADDITION}\n&/" chart/templates/rendered.yaml chart/templates/rendered.yaml
fi


# flagD-aemon
