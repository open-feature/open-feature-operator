#!/bin/bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
kubectl wait --for=condition=Available=True deploy --all -n 'cert-manager'
kubectl create ns 'open-feature-operator-system' --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.5/certificate.yaml
kubectl apply -f https://github.com/open-feature/open-feature-operator/releases/download/v0.2.5/release.yaml
kubectl wait --for=condition=Available=True deploy --all -n 'open-feature-operator-system'
