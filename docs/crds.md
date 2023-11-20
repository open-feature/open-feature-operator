# API Reference

Packages:

- [core.openfeature.dev/v1beta1](#coreopenfeaturedevv1beta1)

# core.openfeature.dev/v1beta1

Resource Types:

- [FeatureFlag](#featureflag)

- [FeatureFlagSource](#featureflagsource)




## FeatureFlag
<sup><sup>[↩ Parent](#coreopenfeaturedevv1beta1 )</sup></sup>






FeatureFlag is the Schema for the featureflags API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>core.openfeature.dev/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FeatureFlag</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagspec">spec</a></b></td>
        <td>object</td>
        <td>
          FeatureFlagSpec defines the desired state of FeatureFlag<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FeatureFlagStatus defines the observed state of FeatureFlag<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlag.spec
<sup><sup>[↩ Parent](#featureflag)</sup></sup>



FeatureFlagSpec defines the desired state of FeatureFlag

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#featureflagspecflagspec">flagSpec</a></b></td>
        <td>object</td>
        <td>
          FlagSpec is the structured representation of the feature flag specification<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlag.spec.flagSpec
<sup><sup>[↩ Parent](#featureflagspec)</sup></sup>



FlagSpec is the structured representation of the feature flag specification

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#featureflagspecflagspecflagskey">flags</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>$evaluators</b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlag.spec.flagSpec.flags[key]
<sup><sup>[↩ Parent](#featureflagspecflagspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultVariant</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>state</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: ENABLED, DISABLED<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>variants</b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>targeting</b></td>
        <td>object</td>
        <td>
          Targeting is the json targeting rule<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## FeatureFlagSource
<sup><sup>[↩ Parent](#coreopenfeaturedevv1beta1 )</sup></sup>






FeatureFlagSource is the Schema for the FeatureFlagSources API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>core.openfeature.dev/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FeatureFlagSource</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagsourcespec">spec</a></b></td>
        <td>object</td>
        <td>
          FeatureFlagSourceSpec defines the desired state of FeatureFlagSource<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FeatureFlagSourceStatus defines the observed state of FeatureFlagSource<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec
<sup><sup>[↩ Parent](#featureflagsource)</sup></sup>



FeatureFlagSourceSpec defines the desired state of FeatureFlagSource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#featureflagsourcespecsourcesindex">sources</a></b></td>
        <td>[]object</td>
        <td>
          SyncProviders define the syncProviders and associated configuration to be applied to the sidecar<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>debugLogging</b></td>
        <td>boolean</td>
        <td>
          DebugLogging defines whether to enable --debug flag of flagd sidecar. Default false (disabled).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>defaultSyncProvider</b></td>
        <td>string</td>
        <td>
          DefaultSyncProvider defines the default sync provider<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>envVarPrefix</b></td>
        <td>string</td>
        <td>
          EnvVarPrefix defines the prefix to be applied to all environment variables applied to the sidecar, default FLAGD<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagsourcespecenvvarsindex">envVars</a></b></td>
        <td>[]object</td>
        <td>
          EnvVars define the env vars to be applied to the sidecar, any env vars in FeatureFlag CRs are added at the lowest index, all values will have the EnvVarPrefix applied, default FLAGD<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>evaluator</b></td>
        <td>string</td>
        <td>
          Evaluator sets an evaluator, defaults to 'json'<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>logFormat</b></td>
        <td>string</td>
        <td>
          LogFormat allows for the sidecar log format to be overridden, defaults to 'json'<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>managementPort</b></td>
        <td>integer</td>
        <td>
          ManagemetPort defines the port to serve management on, defaults to 8014<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>otelCollectorUri</b></td>
        <td>string</td>
        <td>
          OtelCollectorUri defines whether to enable --otel-collector-uri flag of flagd sidecar. Default false (disabled).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port defines the port to listen on, defaults to 8013<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>probesEnabled</b></td>
        <td>boolean</td>
        <td>
          ProbesEnabled defines whether to enable liveness and readiness probes of flagd sidecar. Default true (enabled).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagsourcespecresources">resources</a></b></td>
        <td>object</td>
        <td>
          Resources defines flagd sidecar resources. Default to operator sidecar-cpu-* and sidecar-ram-* flags.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>rolloutOnChange</b></td>
        <td>boolean</td>
        <td>
          RolloutOnChange dictates whether annotated deployments will be restarted when configuration changes are detected in this CR, defaults to false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>socketPath</b></td>
        <td>string</td>
        <td>
          SocketPath defines the unix socket path to listen on<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>syncProviderArgs</b></td>
        <td>[]string</td>
        <td>
          SyncProviderArgs are string arguments passed to all sync providers, defined as key values separated by =<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.sources[index]
<sup><sup>[↩ Parent](#featureflagsourcespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>source</b></td>
        <td>string</td>
        <td>
          Source is a URI of the flag sources<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>certPath</b></td>
        <td>string</td>
        <td>
          CertPath is a path of a certificate to be used by grpc TLS connection<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>httpSyncBearerToken</b></td>
        <td>string</td>
        <td>
          HttpSyncBearerToken is a bearer token. Used by http(s) sync provider only<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>provider</b></td>
        <td>string</td>
        <td>
          Provider type - kubernetes, http(s), grpc(s) or filepath<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>providerID</b></td>
        <td>string</td>
        <td>
          ProviderID is an identifier to be used in grpc provider<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>selector</b></td>
        <td>string</td>
        <td>
          Selector is a flag configuration selector used by grpc provider<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tls</b></td>
        <td>boolean</td>
        <td>
          TLS - Enable/Disable secure TLS connectivity. Currently used only by GRPC sync<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.envVars[index]
<sup><sup>[↩ Parent](#featureflagsourcespec)</sup></sup>



EnvVar represents an environment variable present in a Container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable. Must be a C&lowbar;IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR&lowbar;NAME) are expanded using the previously defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced to a single $, which allows for escaping the $(VAR&lowbar;NAME) syntax: i.e. "$$(VAR&lowbar;NAME)" will produce the string literal "$(VAR&lowbar;NAME)". Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagsourcespecenvvarsindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.envVars[index].valueFrom
<sup><sup>[↩ Parent](#featureflagsourcespecenvvarsindex)</sup></sup>



Source for the environment variable's value. Cannot be used if value is not empty.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#featureflagsourcespecenvvarsindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagsourcespecenvvarsindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['&lt;KEY&gt;']`, `metadata.annotations['&lt;KEY&gt;']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagsourcespecenvvarsindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagsourcespecenvvarsindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.envVars[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#featureflagsourcespecenvvarsindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.envVars[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#featureflagsourcespecenvvarsindexvaluefrom)</sup></sup>



Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['&lt;KEY&gt;']`, `metadata.annotations['&lt;KEY&gt;']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.envVars[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#featureflagsourcespecenvvarsindexvaluefrom)</sup></sup>



Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.envVars[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#featureflagsourcespecenvvarsindexvaluefrom)</sup></sup>



Selects a key of a secret in the pod's namespace

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.resources
<sup><sup>[↩ Parent](#featureflagsourcespec)</sup></sup>



Resources defines flagd sidecar resources. Default to operator sidecar-cpu-* and sidecar-ram-* flags.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#featureflagsourcespecresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. 
 This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. 
 This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagSource.spec.resources.claims[index]
<sup><sup>[↩ Parent](#featureflagsourcespecresources)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name must match the name of one entry in pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>