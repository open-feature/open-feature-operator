# Flagd

The CRD `Flagd` at version `v1beta1` is used to create a standalone flagd deployment,
accompanied by a `Service` and an optional `Ingress` to expose its API
endpoint to clients outside the cluster.

The handling of this resource can be enabled/disabled by setting `managerConfig.flagdResourceEnabled` Helm value of the operator
chart to `true` or `false` respectively.

Below is an example of a `Flagd` resource:

```yaml
apiVersion: core.openfeature.dev/v1beta1
kind: Flagd
metadata:
  name: flagd-sample
spec:
  replicas: 2
  serviceType: ClusterIP
  serviceAccountName: default
  featureFlagSource: end-to-end
  ingress:
    enabled: true
    annotations:
      nginx.ingress.kubernetes.io/force-ssl-redirect: "false"
    hosts:
      - flagd-sample
    ingressClassName: nginx
    pathType: ImplementationSpecific
```

In the example above, we have created a `Flagd` resource called `flagd-sample`,
which results the following resources to be created by the operator
after applying it:

- A `flagd-sample` `Deployment` with two replicas, running an instance of `flagd` each:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: flagd-sample
    app.kubernetes.io/managed-by: open-feature-operator
    app.kubernetes.io/name: flagd-sample
  name: flagd-sample
  ownerReferences:
    - apiVersion: core.openfeature.dev/v1beta1
      kind: Flagd
      name: flagd-sample
spec:
  replicas: 2
  selector:
    matchLabels:
      app: flagd-sample
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: flagd-sample
        app.kubernetes.io/managed-by: open-feature-operator
        app.kubernetes.io/name: flagd-sample
    spec:
      containers:
          - name: flagd
            # renovate: datasource=github-tags depName=open-feature/flagd/flagd
            image: ghcr.io/open-feature/flagd:v0.10.1
            ports:
              - containerPort: 8014
                name: management
                protocol: TCP
              - containerPort: 8013
                name: flagd
                protocol: TCP
              - containerPort: 8016
                name: ofrep
                protocol: TCP
              - containerPort: 8015
                name: sync
                protocol: TCP
      serviceAccount: default
      serviceAccountName: default
```

- A `flagd-sample` `Service` with the type set to `ClusterIP`, that enables access to the pods
running the flagd instance:

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: flagd-sample
    app.kubernetes.io/managed-by: open-feature-operator
    app.kubernetes.io/name: flagd-sample
  name: flagd-sample
  ownerReferences:
    - apiVersion: core.openfeature.dev/v1beta1
      kind: Flagd
      name: flagd-sample
spec:
  ports:
    - name: flagd
      port: 8013
      protocol: TCP
      targetPort: 8013
    - name: ofrep
      port: 8016
      protocol: TCP
      targetPort: 8016
    - name: sync
      port: 8015
      protocol: TCP
      targetPort: 8015
    - name: metrics
      port: 8014
      protocol: TCP
      targetPort: 8014
  selector:
    app: flagd-sample
  type: ClusterIP
```

- A `flagd-sample` `Ingress` enabling the communication between outside clients and the `flagd-sample` `Service`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    app: flagd-sample
    app.kubernetes.io/managed-by: open-feature-operator
    app.kubernetes.io/name: flagd-sample
  name: flagd-sample
  annotations:
    nginx.ingress.kubernetes.io/force-ssl-redirect: "false"
  ownerReferences:
    - apiVersion: core.openfeature.dev/v1beta1
      kind: Flagd
      name: flagd-sample
spec:
  ingressClassName: nginx
  rules:
    - host: flagd-sample
      http:
        paths:
          - backend:
              service:
                name: flagd-sample
                port:
                  number: 8013
            path: /flagd
            pathType: ImplementationSpecific
          - backend:
              service:
                name: flagd-sample
                port:
                  number: 8016
            path: /ofrep
            pathType: ImplementationSpecific
          - backend:
              service:
                name: flagd-sample
                port:
                  number: 8015
            path: /sync
            pathType: ImplementationSpecific
```

Note that if the flagd service is intended only for cluster-internal use, the creation of the `Ingress` can be disabled
by setting the `spec.ingress.enabled` parameter of the `Flagd` resource to `false`.
