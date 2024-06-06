#!/bin/bash

# This script is a hack to support helm flow control in kustomize overlays, which would otherwise break them.
# It allows us to render helm template bindings and add newlines.
# For instance, it transforms "___{{ .Value.myValue }}___" to {{ .Value.myValue }}.
# It also adds newlines wherever ___newline___ is found.

CHARTS_DIR='./chart/open-feature-operator/templates';

echo 'Running strip-kustomize-helm.sh script'
filenames=`find $CHARTS_DIR -name "*.yaml"`
for file in $filenames; do
    sed -i "s/___newline___/\\n/g" $file
    sed -i "s/\"___//g" $file
    sed -i "s/___\"//g" $file
    sed -i "s/___//g" $file
done
echo 'Done running strip-kustomize-helm.sh script'