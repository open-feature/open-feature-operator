#!/bin/bash

cat <<EOF | kubectl -n open-feature-operator-system apply -f -
apiVersion: core.openfeature.dev/v1alpha2
kind: FeatureFlagConfiguration
metadata:
  name: end-to-end-test-filepath
spec:
  syncProvider:
    name: filepath
  featureFlagSpec:
    flags:
      simple-flag-filepath:
        state: ENABLED
        variants:
          "on": true
          "off": false
        defaultVariant: "off"
EOF

# filepath sync provider takes up to 2 minutes to synchronize, retry every 10 seconds up to 12 times
./"$(dirname "${BASH_SOURCE[0]}")"/../flag-evaluation.sh simple-flag-filepath '{"reason":"STATIC","variant":"off"}' 12 10
EXIT_CODE=$?

kubectl -n open-feature-operator-system apply -f ./test/e2e/e2e.yml > /dev/null # reset state quietly

exit $EXIT_CODE
