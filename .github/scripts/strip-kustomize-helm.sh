#!/usr/bin/env bash

# This script is a hack to support helm flow control in kustomize overlays, which would otherwise break them.
# It allows us to render helm template bindings and add newlines.
# For instance, it transforms "___{{ .Value.myValue }}___" to {{ .Value.myValue }}.
# It also adds newlines wherever ___newline___ is found, and other operations. See
# sed_expressions below.

echo 'Running strip-kustomize-helm.sh script'
CHARTS_DIR='./chart/open-feature-operator/templates'

# Careful! Ordering of these expressions matter!
sed_expressions=(
    "s/___newline___/\\n/g"
    "s/___space___/ /g"
    "s/\"___//g"
    "s/___\"//g"
    "/___delete_me___/d"
    "s/___//g"
)

find $CHARTS_DIR -name "*.yaml" | while read file; do
    for expr in "${sed_expressions[@]}"; do
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS (BSD) version
            sed -i '' "$expr" "$file"
        else
            # Linux (GNU) version
            sed -i "$expr" "$file"
        fi
    done
done

echo 'Done running strip-kustomize-helm.sh script'
