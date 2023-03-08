#!/bin/bash

# delete existing deployment
kubectl -n open-feature-operator-system delete deployment open-feature-e2e-test-deployment

# set openfeature.dev/enabled annotation to false and redeploy
cat <<EOF | kubectl -n open-feature-operator-system apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: open-feature-e2e-test-deployment
  labels:
    app: open-feature-e2e-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: open-feature-e2e-test
  template:
    metadata:
      labels:
        app: open-feature-e2e-test
      annotations:
        openfeature.dev/enabled: "false"
    spec:
      serviceAccountName: open-feature-e2e-test-sa
      volumes:
        - name: open-feature-e2e-nginx-conf
          configMap:
            name: open-feature-e2e-nginx-conf
            items:
              - key: nginx.conf
                path: nginx.conf
      containers:
        - name: open-feature-e2e-test
          image: nginx:stable-alpine
          ports:
            - containerPort: 80
          volumeMounts:
            - name: open-feature-e2e-nginx-conf
              mountPath: /etc/nginx
              readOnly: true
EOF

# wait until deployment is ready
kubectl wait --for=condition=Available=True deploy --all -n 'open-feature-operator-system'

# curl to nginx reverse proxy should return non 200 status code (flagd hasn't been deployed)
STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "localhost:30000/schema.v1.Service/ResolveBoolean" -d "{\"flagKey\":\"simple-flag\",\"context\":{}}" -H "Content-Type: application/json")

# delete deployment then reset state and wait until ready
kubectl -n open-feature-operator-system delete deployment open-feature-e2e-test-deployment
kubectl -n open-feature-operator-system apply -f ./test/e2e/e2e.yml > /dev/null
kubectl wait --for=condition=Available=True deploy --all -n 'open-feature-operator-system'

if [ "$STATUS_CODE" -eq 200 ]; then
  echo "Expected curl to nginx reverse proxy to return non 200 status code when openfeature.dev/enabled annotation is false."
  exit 1
else
  exit 0
fi
