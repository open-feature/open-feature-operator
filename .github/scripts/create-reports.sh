#!/bin/bash

ignore="--ignore-not-found"
logsDir="logs"

createResourceReport () {
    path=$1
    namespace=$2
    resource=$3
    withLogs=$4

    mkdir -p "$path/$resource"

    kubectl get "$resource" -n "$namespace" "$ignore" > "$path/$resource/list-$resource.txt"

    for r in $(kubectl get "$resource" -n "$namespace" "$ignore" -o jsonpath='{.items[*].metadata.name}'); do
        kubectl describe "$resource/$r" -n "$namespace" > "$path/$resource/$r-describe.txt"

        if $withLogs ; then
        kubectl logs "$resource/$r" --all-containers=true -n "$namespace" > "$path/$resource/$r-logs.txt"
        fi
    done
}

# Go through each namespace in the cluster
for namespace in $(kubectl get namespaces -o jsonpath='{.items[*].metadata.name}'); do

    mkdir -p "$logsDir/$namespace"
    createResourceReport "$logsDir/$namespace" "$namespace" "Pods" true
    createResourceReport "$logsDir/$namespace" "$namespace" "Deployments" false
    createResourceReport "$logsDir/$namespace" "$namespace" "Daemonsets" false
    createResourceReport "$logsDir/$namespace" "$namespace" "Statefulsets" false
    createResourceReport "$logsDir/$namespace" "$namespace" "Jobs" false
    createResourceReport "$logsDir/$namespace" "$namespace" "FeatureFlagConfiguration" false
    createResourceReport "$logsDir/$namespace" "$namespace" "FlagSourceConfiguration" false
    
done
