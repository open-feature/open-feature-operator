# API Reference

Packages:

- [core.openfeature.dev/v1alpha1](#coreopenfeaturedevv1alpha1)
- [core.openfeature.dev/v1alpha2](#coreopenfeaturedevv1alpha2)
- [core.openfeature.dev/v1alpha3](#coreopenfeaturedevv1alpha3)

# core.openfeature.dev/v1alpha1

Resource Types:

- [FeatureFlagConfiguration](#featureflagconfiguration)

- [FlagSourceConfiguration](#flagsourceconfiguration)




## FeatureFlagConfiguration
<sup><sup>[↩ Parent](#coreopenfeaturedevv1alpha1 )</sup></sup>






FeatureFlagConfiguration is the Schema for the featureflagconfigurations API

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
      <td>core.openfeature.dev/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FeatureFlagConfiguration</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspec">spec</a></b></td>
        <td>object</td>
        <td>
          FeatureFlagConfigurationSpec defines the desired state of FeatureFlagConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FeatureFlagConfigurationStatus defines the observed state of FeatureFlagConfiguration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec
<sup><sup>[↩ Parent](#featureflagconfiguration)</sup></sup>



FeatureFlagConfigurationSpec defines the desired state of FeatureFlagConfiguration

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
        <td><b>featureFlagSpec</b></td>
        <td>string</td>
        <td>
          FeatureFlagSpec is the json representation of the feature flag<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspec">flagDSpec</a></b></td>
        <td>object</td>
        <td>
          FlagDSpec [DEPRECATED]: superseded by FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecserviceprovider">serviceProvider</a></b></td>
        <td>object</td>
        <td>
          ServiceProvider [DEPRECATED]: superseded by FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecsyncprovider">syncProvider</a></b></td>
        <td>object</td>
        <td>
          SyncProvider [DEPRECATED]: superseded by FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.flagDSpec
<sup><sup>[↩ Parent](#featureflagconfigurationspec)</sup></sup>



FlagDSpec [DEPRECATED]: superseded by FlagSourceConfiguration

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
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindex">envs</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>metricsPort</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.flagDSpec.envs[index]
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspec)</sup></sup>



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
          Name of the environment variable. Must be a C&#95;IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references &dollar;(VAR&#95;NAME) are expanded using the previously defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced to a single $, which allows for escaping the &dollar;(VAR&#95;NAME) syntax: i.e. "$&dollar;(VAR&#95;NAME)" will produce the string literal "&dollar;(VAR&#95;NAME)". Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindex)</sup></sup>



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
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['&lt;KEY&gt;']`, `metadata.annotations['&lt;KEY&gt;']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom)</sup></sup>



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


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom)</sup></sup>



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


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom)</sup></sup>



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


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom)</sup></sup>



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


### FeatureFlagConfiguration.spec.serviceProvider
<sup><sup>[↩ Parent](#featureflagconfigurationspec)</sup></sup>



ServiceProvider [DEPRECATED]: superseded by FlagSourceConfiguration

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
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: flagd<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecserviceprovidercredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs. 1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage. 2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted". Those cannot be well described when embedded. 3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen. 4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple and the version of the actual struct is irrelevant. 5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. 
 Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.serviceProvider.credentials
<sup><sup>[↩ Parent](#featureflagconfigurationspecserviceprovider)</sup></sup>



ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs. 1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage. 2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted". Those cannot be well described when embedded. 3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen. 4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple and the version of the actual struct is irrelevant. 5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. 
 Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .

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
        <td>
          API version of the referent.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceVersion</b></td>
        <td>string</td>
        <td>
          Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.syncProvider
<sup><sup>[↩ Parent](#featureflagconfigurationspec)</sup></sup>



SyncProvider [DEPRECATED]: superseded by FlagSourceConfiguration

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
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecsyncproviderhttpsyncconfiguration">httpSyncConfiguration</a></b></td>
        <td>object</td>
        <td>
          HttpSyncConfiguration defines the desired configuration for a http sync<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.syncProvider.httpSyncConfiguration
<sup><sup>[↩ Parent](#featureflagconfigurationspecsyncprovider)</sup></sup>



HttpSyncConfiguration defines the desired configuration for a http sync

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
        <td><b>target</b></td>
        <td>string</td>
        <td>
          Target is the target url for flagd to poll<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>bearerToken</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## FlagSourceConfiguration
<sup><sup>[↩ Parent](#coreopenfeaturedevv1alpha1 )</sup></sup>






FlagSourceConfiguration is the Schema for the FlagSourceConfigurations API

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
      <td>core.openfeature.dev/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FlagSourceConfiguration</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspec">spec</a></b></td>
        <td>object</td>
        <td>
          FlagSourceConfigurationSpec defines the desired state of FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FlagSourceConfigurationStatus defines the observed state of FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec
<sup><sup>[↩ Parent](#flagsourceconfiguration)</sup></sup>



FlagSourceConfigurationSpec defines the desired state of FlagSourceConfiguration

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
        <td><b><a href="#flagsourceconfigurationspecsourcesindex">sources</a></b></td>
        <td>[]object</td>
        <td>
          Sources defines the syncProviders and associated configuration to be applied to the sidecar<br/>
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
        <td><b><a href="#flagsourceconfigurationspecenvvarsindex">envVars</a></b></td>
        <td>[]object</td>
        <td>
          EnvVars define the env vars to be applied to the sidecar, any env vars in FeatureFlagConfiguration CRs are added at the lowest index, all values will have the EnvVarPrefix applied<br/>
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
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image allows for the sidecar image to be overridden, defaults to 'ghcr.io/open-feature/flagd'<br/>
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
        <td><b>metricsPort</b></td>
        <td>integer</td>
        <td>
          MetricsPort defines the port to serve metrics on, defaults to 8014<br/>
          <br/>
            <i>Format</i>: int32<br/>
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
      </tr><tr>
        <td><b>tag</b></td>
        <td>string</td>
        <td>
          Tag to be appended to the sidecar image, defaults to 'main'<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec.sources[index]
<sup><sup>[↩ Parent](#flagsourceconfigurationspec)</sup></sup>





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
          Provider type - kubernetes, http, grpc or filepath<br/>
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


### FlagSourceConfiguration.spec.envVars[index]
<sup><sup>[↩ Parent](#flagsourceconfigurationspec)</sup></sup>



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
          Name of the environment variable. Must be a C&#95;IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references &dollar;(VAR&#95;NAME) are expanded using the previously defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced to a single $, which allows for escaping the &dollar;(VAR&#95;NAME) syntax: i.e. "$&dollar;(VAR&#95;NAME)" will produce the string literal "&dollar;(VAR&#95;NAME)". Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec.envVars[index].valueFrom
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindex)</sup></sup>



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
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['&lt;KEY&gt;']`, `metadata.annotations['&lt;KEY&gt;']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec.envVars[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom)</sup></sup>



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


### FlagSourceConfiguration.spec.envVars[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom)</sup></sup>



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


### FlagSourceConfiguration.spec.envVars[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom)</sup></sup>



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


### FlagSourceConfiguration.spec.envVars[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom)</sup></sup>



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

# core.openfeature.dev/v1alpha2

Resource Types:

- [FeatureFlagConfiguration](#featureflagconfiguration)

- [FlagSourceConfiguration](#flagsourceconfiguration)




## FeatureFlagConfiguration
<sup><sup>[↩ Parent](#coreopenfeaturedevv1alpha2 )</sup></sup>






FeatureFlagConfiguration is the Schema for the featureflagconfigurations API

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
      <td>core.openfeature.dev/v1alpha2</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FeatureFlagConfiguration</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          FeatureFlagConfigurationSpec defines the desired state of FeatureFlagConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FeatureFlagConfigurationStatus defines the observed state of FeatureFlagConfiguration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec
<sup><sup>[↩ Parent](#featureflagconfiguration-1)</sup></sup>



FeatureFlagConfigurationSpec defines the desired state of FeatureFlagConfiguration

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
        <td><b><a href="#featureflagconfigurationspecfeatureflagspec">featureFlagSpec</a></b></td>
        <td>object</td>
        <td>
          FeatureFlagSpec is the structured representation of the feature flag specification<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspec-1">flagDSpec</a></b></td>
        <td>object</td>
        <td>
          FlagDSpec [DEPRECATED]: superseded by FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecserviceprovider-1">serviceProvider</a></b></td>
        <td>object</td>
        <td>
          ServiceProvider [DEPRECATED]: superseded by FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecsyncprovider-1">syncProvider</a></b></td>
        <td>object</td>
        <td>
          SyncProvider [DEPRECATED]: superseded by FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.featureFlagSpec
<sup><sup>[↩ Parent](#featureflagconfigurationspec-1)</sup></sup>



FeatureFlagSpec is the structured representation of the feature flag specification

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
        <td><b><a href="#featureflagconfigurationspecfeatureflagspecflagskey">flags</a></b></td>
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


### FeatureFlagConfiguration.spec.featureFlagSpec.flags[key]
<sup><sup>[↩ Parent](#featureflagconfigurationspecfeatureflagspec)</sup></sup>





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


### FeatureFlagConfiguration.spec.flagDSpec
<sup><sup>[↩ Parent](#featureflagconfigurationspec-1)</sup></sup>



FlagDSpec [DEPRECATED]: superseded by FlagSourceConfiguration

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
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindex-1">envs</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.flagDSpec.envs[index]
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspec-1)</sup></sup>



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
          Name of the environment variable. Must be a C&#95;IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references &dollar;(VAR&#95;NAME) are expanded using the previously defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced to a single $, which allows for escaping the &dollar;(VAR&#95;NAME) syntax: i.e. "$&dollar;(VAR&#95;NAME)" will produce the string literal "&dollar;(VAR&#95;NAME)". Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefrom-1">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindex-1)</sup></sup>



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
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromconfigmapkeyref-1">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromfieldref-1">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['&lt;KEY&gt;']`, `metadata.annotations['&lt;KEY&gt;']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromresourcefieldref-1">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecflagdspecenvsindexvaluefromsecretkeyref-1">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom-1)</sup></sup>



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


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom-1)</sup></sup>



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


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom-1)</sup></sup>



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


### FeatureFlagConfiguration.spec.flagDSpec.envs[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#featureflagconfigurationspecflagdspecenvsindexvaluefrom-1)</sup></sup>



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


### FeatureFlagConfiguration.spec.serviceProvider
<sup><sup>[↩ Parent](#featureflagconfigurationspec-1)</sup></sup>



ServiceProvider [DEPRECATED]: superseded by FlagSourceConfiguration

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
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: flagd<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecserviceprovidercredentials-1">credentials</a></b></td>
        <td>object</td>
        <td>
          ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs. 1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage. 2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted". Those cannot be well described when embedded. 3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen. 4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple and the version of the actual struct is irrelevant. 5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. 
 Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.serviceProvider.credentials
<sup><sup>[↩ Parent](#featureflagconfigurationspecserviceprovider-1)</sup></sup>



ObjectReference contains enough information to let you inspect or modify the referred object. --- New uses of this type are discouraged because of difficulty describing its usage when embedded in APIs. 1. Ignored fields.  It includes many fields which are not generally honored.  For instance, ResourceVersion and FieldPath are both very rarely valid in actual usage. 2. Invalid usage help.  It is impossible to add specific help for individual usage.  In most embedded usages, there are particular restrictions like, "must refer only to types A and B" or "UID not honored" or "name must be restricted". Those cannot be well described when embedded. 3. Inconsistent validation.  Because the usages are different, the validation rules are different by usage, which makes it hard for users to predict what will happen. 4. The fields are both imprecise and overly precise.  Kind is not a precise mapping to a URL. This can produce ambiguity during interpretation and require a REST mapping.  In most cases, the dependency is on the group,resource tuple and the version of the actual struct is irrelevant. 5. We cannot easily change it.  Because this type is embedded in many locations, updates to this type will affect numerous schemas.  Don't make new APIs embed an underspecified API type they do not control. 
 Instead of using this type, create a locally provided and used type that is well-focused on your reference. For example, ServiceReferences for admission registration: https://github.com/kubernetes/api/blob/release-1.17/admissionregistration/v1/types.go#L533 .

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
        <td>
          API version of the referent.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: "spec.containers{name}" (where "name" refers to the name of the container that triggered the event) or if no container name is specified "spec.containers[2]" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object. TODO: this design is not final and this field is subject to change in the future.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceVersion</b></td>
        <td>string</td>
        <td>
          Specific resourceVersion to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.syncProvider
<sup><sup>[↩ Parent](#featureflagconfigurationspec-1)</sup></sup>



SyncProvider [DEPRECATED]: superseded by FlagSourceConfiguration

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
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#featureflagconfigurationspecsyncproviderhttpsyncconfiguration-1">httpSyncConfiguration</a></b></td>
        <td>object</td>
        <td>
          HttpSyncConfiguration defines the desired configuration for a http sync<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FeatureFlagConfiguration.spec.syncProvider.httpSyncConfiguration
<sup><sup>[↩ Parent](#featureflagconfigurationspecsyncprovider-1)</sup></sup>



HttpSyncConfiguration defines the desired configuration for a http sync

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
        <td><b>target</b></td>
        <td>string</td>
        <td>
          Target is the target url for flagd to poll<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>bearerToken</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## FlagSourceConfiguration
<sup><sup>[↩ Parent](#coreopenfeaturedevv1alpha2 )</sup></sup>






FlagSourceConfiguration is the Schema for the FlagSourceConfigurations API

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
      <td>core.openfeature.dev/v1alpha2</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FlagSourceConfiguration</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          FlagSourceConfigurationSpec defines the desired state of FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FlagSourceConfigurationStatus defines the observed state of FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec
<sup><sup>[↩ Parent](#flagsourceconfiguration-1)</sup></sup>



FlagSourceConfigurationSpec defines the desired state of FlagSourceConfiguration

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
        <td><b>defaultSyncProvider</b></td>
        <td>string</td>
        <td>
          DefaultSyncProvider defines the default sync provider<br/>
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
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image allows for the sidecar image to be overridden, defaults to 'ghcr.io/open-feature/flagd'<br/>
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
        <td><b>metricsPort</b></td>
        <td>integer</td>
        <td>
          MetricsPort defines the port to serve metrics on, defaults to 8013<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port defines the port to listen on, defaults to 8014<br/>
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
      </tr><tr>
        <td><b>tag</b></td>
        <td>string</td>
        <td>
          Tag to be appended to the sidecar image, defaults to 'main'<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

# core.openfeature.dev/v1alpha3

Resource Types:

- [FlagSourceConfiguration](#flagsourceconfiguration)




## FlagSourceConfiguration
<sup><sup>[↩ Parent](#coreopenfeaturedevv1alpha3 )</sup></sup>






FlagSourceConfiguration is the Schema for the FlagSourceConfigurations API

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
      <td>core.openfeature.dev/v1alpha3</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>FlagSourceConfiguration</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          FlagSourceConfigurationSpec defines the desired state of FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          FlagSourceConfigurationStatus defines the observed state of FlagSourceConfiguration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec
<sup><sup>[↩ Parent](#flagsourceconfiguration-1)</sup></sup>



FlagSourceConfigurationSpec defines the desired state of FlagSourceConfiguration

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
        <td><b><a href="#flagsourceconfigurationspecsourcesindex-1">sources</a></b></td>
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
        <td><b><a href="#flagsourceconfigurationspecenvvarsindex-1">envVars</a></b></td>
        <td>[]object</td>
        <td>
          EnvVars define the env vars to be applied to the sidecar, any env vars in FeatureFlagConfiguration CRs are added at the lowest index, all values will have the EnvVarPrefix applied, default FLAGD<br/>
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
        <td><b>image</b></td>
        <td>string</td>
        <td>
          Image allows for the sidecar image to be overridden, defaults to 'ghcr.io/open-feature/flagd'<br/>
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
        <td><b>metricsPort</b></td>
        <td>integer</td>
        <td>
          MetricsPort defines the port to serve metrics on, defaults to 8014<br/>
          <br/>
            <i>Format</i>: int32<br/>
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
      </tr><tr>
        <td><b>tag</b></td>
        <td>string</td>
        <td>
          Tag to be appended to the sidecar image, defaults to 'main'<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec.sources[index]
<sup><sup>[↩ Parent](#flagsourceconfigurationspec-1)</sup></sup>





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


### FlagSourceConfiguration.spec.envVars[index]
<sup><sup>[↩ Parent](#flagsourceconfigurationspec-1)</sup></sup>



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
          Name of the environment variable. Must be a C&#95;IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references &dollar;(VAR&#95;NAME) are expanded using the previously defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced to a single $, which allows for escaping the &dollar;(VAR&#95;NAME) syntax: i.e. "$&dollar;(VAR&#95;NAME)" will produce the string literal "&dollar;(VAR&#95;NAME)". Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefrom-1">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec.envVars[index].valueFrom
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindex-1)</sup></sup>



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
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromconfigmapkeyref-1">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromfieldref-1">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['&lt;KEY&gt;']`, `metadata.annotations['&lt;KEY&gt;']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromresourcefieldref-1">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#flagsourceconfigurationspecenvvarsindexvaluefromsecretkeyref-1">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### FlagSourceConfiguration.spec.envVars[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom-1)</sup></sup>



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


### FlagSourceConfiguration.spec.envVars[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom-1)</sup></sup>



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


### FlagSourceConfiguration.spec.envVars[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom-1)</sup></sup>



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


### FlagSourceConfiguration.spec.envVars[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#flagsourceconfigurationspecenvvarsindexvaluefrom-1)</sup></sup>



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