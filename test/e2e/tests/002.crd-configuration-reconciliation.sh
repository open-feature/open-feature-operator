#!/bin/bash

cat <<EOF | kubectl -n open-feature-operator-system apply -f -
apiVersion: core.openfeature.dev/v1alpha1
kind: FeatureFlagConfiguration
metadata:
  name: end-to-end-test
spec:
  featureFlagSpec: |
    {
      "flags": {
        "simple-flag": {
          "state": "ENABLED",
          "variants": {
            "on": true,
            "off": false
          },
          "defaultVariant": "off"
        }
      }
    }
EOF

./"$(dirname "${BASH_SOURCE[0]}")"/../simple-flag-evaluation.sh '{"reason":"STATIC","variant":"off"}'
EXIT_CODE=$?

kubectl -n open-feature-operator-system apply -f ./test/e2e/e2e.yml > /dev/null # reset state quietly

exit $EXIT_CODE
