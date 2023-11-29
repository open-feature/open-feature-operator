# Changelog

## [0.3.0](https://github.com/open-feature/open-feature-operator/compare/operator-v0.2.36...operator/v0.3.0) (2023-11-29)


### ⚠ BREAKING CHANGES

* use v1beta1 in operator logic ([#539](https://github.com/open-feature/open-feature-operator/issues/539))

### ✨ New Features

* Introduce v1beta1 API version ([#535](https://github.com/open-feature/open-feature-operator/issues/535)) ([3acd492](https://github.com/open-feature/open-feature-operator/commit/3acd49289a40e8f07fd20aad46185ac42ceb1b7a))
* prepare apis for v1beta1 controllers onboarding ([#549](https://github.com/open-feature/open-feature-operator/issues/549)) ([e3c8b42](https://github.com/open-feature/open-feature-operator/commit/e3c8b4290be99d78b88ffef686531a38b97e61be))
* release APIs and Operator independently ([#541](https://github.com/open-feature/open-feature-operator/issues/541)) ([7b1af42](https://github.com/open-feature/open-feature-operator/commit/7b1af42ac41e63ccbb1820b31f579ffea679cff6))
* restricting sidecar image and tag setup ([#550](https://github.com/open-feature/open-feature-operator/issues/550)) ([233be79](https://github.com/open-feature/open-feature-operator/commit/233be79b56ccca32a1cb041bce53a9848f032a60))
* update api version to v0.2.38 ([#561](https://github.com/open-feature/open-feature-operator/issues/561)) ([d1f2477](https://github.com/open-feature/open-feature-operator/commit/d1f247727c5b6f4cb5154e94f1090aee0a442346))
* use v1beta1 in operator logic ([#539](https://github.com/open-feature/open-feature-operator/issues/539)) ([d234410](https://github.com/open-feature/open-feature-operator/commit/d234410a809760ba1c8592f95be56891e0cae855))


### 🐛 Bug Fixes

* Revert "chore: release apis 0.2.38" ([#557](https://github.com/open-feature/open-feature-operator/issues/557)) ([ccb8c1d](https://github.com/open-feature/open-feature-operator/commit/ccb8c1d6e12aa36e33239fd96bebbc57fc4ea3bc))


### 🧹 Chore

* clean up unused API code after moving to v1beta1 ([#543](https://github.com/open-feature/open-feature-operator/issues/543)) ([1287b07](https://github.com/open-feature/open-feature-operator/commit/1287b0785fd99cb8bfeaf9fe112aa8a0ed6f5cf9))
* **deps:** update actions/setup-node action to v3.8.1 ([#522](https://github.com/open-feature/open-feature-operator/issues/522)) ([32ddf00](https://github.com/open-feature/open-feature-operator/commit/32ddf002e6c20732d990283946ec124304827bd3))
* fix file source documentation ([#556](https://github.com/open-feature/open-feature-operator/issues/556)) ([318c52d](https://github.com/open-feature/open-feature-operator/commit/318c52d2ba38dbfee6deb3f06d3392dc14d80a6c))
* refactor code to decrease complexity ([#554](https://github.com/open-feature/open-feature-operator/issues/554)) ([17a547f](https://github.com/open-feature/open-feature-operator/commit/17a547f88595cb6c177ca93e1a8b4ad49f3c1a5f))
* release apis 0.2.37 ([#544](https://github.com/open-feature/open-feature-operator/issues/544)) ([854e72d](https://github.com/open-feature/open-feature-operator/commit/854e72d964fce51082220a60fc8a7319676e49c3))
* release apis 0.2.38 ([#548](https://github.com/open-feature/open-feature-operator/issues/548)) ([c6165d4](https://github.com/open-feature/open-feature-operator/commit/c6165d426b5be2af89e03695d24fe0b802fb1fe2))
* release apis 0.2.38 ([#558](https://github.com/open-feature/open-feature-operator/issues/558)) ([4ecbc9b](https://github.com/open-feature/open-feature-operator/commit/4ecbc9b8eeac4e1e86c0f4e11ffedf3dbc376f9a))
* release apis 0.2.38 ([#560](https://github.com/open-feature/open-feature-operator/issues/560)) ([069e275](https://github.com/open-feature/open-feature-operator/commit/069e2754210d1a71bc5b70c0d4e6e193f62a7bcb))
* revert recent release ([#559](https://github.com/open-feature/open-feature-operator/issues/559)) ([f7c79e4](https://github.com/open-feature/open-feature-operator/commit/f7c79e4c6f5a5dee05d7db1796bfb9891dbd53a0))
* use apis tag instead of local replace ([#546](https://github.com/open-feature/open-feature-operator/issues/546)) ([1856918](https://github.com/open-feature/open-feature-operator/commit/18569182c1f2eca3e29e9428a64239ac16ea3c08))
* use github-action for golangci-lint workflow ([#538](https://github.com/open-feature/open-feature-operator/issues/538)) ([a97d336](https://github.com/open-feature/open-feature-operator/commit/a97d336468d5a9b50662f4979784c8388ec10ec1))


### 📚 Documentation

* use v1beta1 API version ([#553](https://github.com/open-feature/open-feature-operator/issues/553)) ([ccc0471](https://github.com/open-feature/open-feature-operator/commit/ccc0471c15cb42a338cd4c1a69b0b1f7c7828837))

## [0.2.36](https://github.com/open-feature/open-feature-operator/compare/v0.2.35...v0.2.36) (2023-08-07)


### ✨ New Features

* add flagd sidecar resources attribute ([#514](https://github.com/open-feature/open-feature-operator/issues/514)) ([56ad0bd](https://github.com/open-feature/open-feature-operator/commit/56ad0bdc3a04457c35d906085e74b39e56970f82))
* add otel collector uri flag ([#513](https://github.com/open-feature/open-feature-operator/issues/513)) ([31d8d5a](https://github.com/open-feature/open-feature-operator/commit/31d8d5a4f9f1132d3b1b517c3acb76c2cb42e0c7))


### 🧹 Chore

* **deps:** update actions/setup-node action to v3.7.0 ([#504](https://github.com/open-feature/open-feature-operator/issues/504)) ([2f78b83](https://github.com/open-feature/open-feature-operator/commit/2f78b836de144234ef222af28069a543f1850eee))
* **deps:** update curlimages/curl docker tag to v8.2.1 ([#505](https://github.com/open-feature/open-feature-operator/issues/505)) ([ae1be55](https://github.com/open-feature/open-feature-operator/commit/ae1be55091086bc0791aaea8a3eed88dd47f5390))
* **deps:** update dependency bitnami-labs/readme-generator-for-helm to v2.5.1 ([#506](https://github.com/open-feature/open-feature-operator/issues/506)) ([54d59db](https://github.com/open-feature/open-feature-operator/commit/54d59db82ce834145cb1d21cdb6595920ad1a0d7))
* **deps:** update docker/login-action digest to a979406 ([#493](https://github.com/open-feature/open-feature-operator/issues/493)) ([22a1e55](https://github.com/open-feature/open-feature-operator/commit/22a1e557adee524006a4eef488a9e6c684a24464))
* **deps:** update helm/kind-action action to v1.8.0 ([#507](https://github.com/open-feature/open-feature-operator/issues/507)) ([e740068](https://github.com/open-feature/open-feature-operator/commit/e74006872ebbc6595332a3722657f64e34ef1f29))
* **deps:** update open-feature/flagd ([#516](https://github.com/open-feature/open-feature-operator/issues/516)) ([74dd65c](https://github.com/open-feature/open-feature-operator/commit/74dd65cd8fa3e45f6935c7bc9394f2341e593cd3))
* update K8s deps and fix api changes ([#518](https://github.com/open-feature/open-feature-operator/issues/518)) ([644144f](https://github.com/open-feature/open-feature-operator/commit/644144ffabfc4b7d527abf030223cef202c22bfe))

## [0.2.35](https://github.com/open-feature/open-feature-operator/compare/v0.2.34...v0.2.35) (2023-08-01)


### 🐛 Bug Fixes

* **deps:** update module github.com/stretchr/testify to v1.8.3 ([#488](https://github.com/open-feature/open-feature-operator/issues/488)) ([426be04](https://github.com/open-feature/open-feature-operator/commit/426be041d0530b8c3a77ba8176ec9e7e280dc162))
* **deps:** update module github.com/stretchr/testify to v1.8.4 ([#490](https://github.com/open-feature/open-feature-operator/issues/490)) ([660da11](https://github.com/open-feature/open-feature-operator/commit/660da11eccb6d6bf6d047d4bdb23225df6610da5))
* remove 'grpc://' prefix from proxy sync address ([#479](https://github.com/open-feature/open-feature-operator/issues/479)) ([50151ff](https://github.com/open-feature/open-feature-operator/commit/50151ffcfd239764da19e76cf657cd511ec882b0))
* use admission webhook namespace if pod namespace is empty ([#503](https://github.com/open-feature/open-feature-operator/issues/503)) ([ffd3e0a](https://github.com/open-feature/open-feature-operator/commit/ffd3e0a8ca1dbc1dbdbe81e36dec0921bd386dc9))


### 🧹 Chore

* adapt ServiceAccount only in case of K8s Provider ([#498](https://github.com/open-feature/open-feature-operator/issues/498)) ([786d511](https://github.com/open-feature/open-feature-operator/commit/786d51160292fcea6f1085891824091a4acb4fcb))
* adding troubleshooting guide ([#501](https://github.com/open-feature/open-feature-operator/issues/501)) ([0befb8f](https://github.com/open-feature/open-feature-operator/commit/0befb8fadbcb4f1925c29faac1e741b77c6ce6a7))
* attempt to improve documentation ([#496](https://github.com/open-feature/open-feature-operator/issues/496)) ([603e74e](https://github.com/open-feature/open-feature-operator/commit/603e74e62bf6d0e248130ac3eeb69e6c574134d1))
* **deps:** update curlimages/curl docker tag to v7.88.1 ([#459](https://github.com/open-feature/open-feature-operator/issues/459)) ([ea98e1e](https://github.com/open-feature/open-feature-operator/commit/ea98e1e77ac616acc4aebf1ea042fc812486ece7))
* **deps:** update curlimages/curl docker tag to v8 ([#461](https://github.com/open-feature/open-feature-operator/issues/461)) ([1271eab](https://github.com/open-feature/open-feature-operator/commit/1271eab2eb4ad6aaab226116cd317345c02f55ac))
* **deps:** update curlimages/curl docker tag to v8.1.2 ([#487](https://github.com/open-feature/open-feature-operator/issues/487)) ([b9720bb](https://github.com/open-feature/open-feature-operator/commit/b9720bb15737786fc1d207d104f2a42b2ec38d6e))
* **deps:** update docker/login-action digest to 40891eb ([#473](https://github.com/open-feature/open-feature-operator/issues/473)) ([630518a](https://github.com/open-feature/open-feature-operator/commit/630518a06b9439753c9a671271b9045d680083fd))
* **deps:** update docker/metadata-action digest to 35e9aff ([#494](https://github.com/open-feature/open-feature-operator/issues/494)) ([27a7efd](https://github.com/open-feature/open-feature-operator/commit/27a7efdc804a4d17531f8505f036978c24b5e2d1))
* **deps:** update docker/metadata-action digest to c4ee3ad ([#471](https://github.com/open-feature/open-feature-operator/issues/471)) ([5f3d98a](https://github.com/open-feature/open-feature-operator/commit/5f3d98a21484a6011a8dde20c9a8018c735cdb63))
* **deps:** update gcr.io/kubebuilder/kube-rbac-proxy docker tag to v0.14.1 ([#477](https://github.com/open-feature/open-feature-operator/issues/477)) ([8183725](https://github.com/open-feature/open-feature-operator/commit/818372531414cdd242b11016a177bc48635c7b28))
* **deps:** update helm/kind-action action to v1.7.0 ([#486](https://github.com/open-feature/open-feature-operator/issues/486)) ([09dcbc1](https://github.com/open-feature/open-feature-operator/commit/09dcbc1b181ae67f7b5e524fad0d2a55f3ded02d))
* **deps:** update module golang.org/x/net to v0.12.0 ([#484](https://github.com/open-feature/open-feature-operator/issues/484)) ([5af75bb](https://github.com/open-feature/open-feature-operator/commit/5af75bb6f4daf760d7869b24183d7b7bc4d2ee96))
* **deps:** update open-feature/flagd ([#480](https://github.com/open-feature/open-feature-operator/issues/480)) ([cfeddc8](https://github.com/open-feature/open-feature-operator/commit/cfeddc89cb8d83019246eb288b4ad4663a3c6cad))
* **deps:** update open-feature/flagd ([#499](https://github.com/open-feature/open-feature-operator/issues/499)) ([83fbb00](https://github.com/open-feature/open-feature-operator/commit/83fbb007ff1fb55c6da299ddfb5f4c0973a17ef1))
* extract flagd container injection into its own component ([#474](https://github.com/open-feature/open-feature-operator/issues/474)) ([9ed8e59](https://github.com/open-feature/open-feature-operator/commit/9ed8e598f8612f5f0935dbd115cd7a8053aa1210))
* generalize renovate configuration ([#495](https://github.com/open-feature/open-feature-operator/issues/495)) ([1ec3183](https://github.com/open-feature/open-feature-operator/commit/1ec3183f750ad929136b76131ff4711effefb398))


### 📚 Documentation

* add advanced flagd links ([#492](https://github.com/open-feature/open-feature-operator/issues/492)) ([eb44c61](https://github.com/open-feature/open-feature-operator/commit/eb44c6110333c0e0a8f39dc32c29245ab40b6bd2))
* add instruction for using OFO and GitOps ([#497](https://github.com/open-feature/open-feature-operator/issues/497)) ([244a625](https://github.com/open-feature/open-feature-operator/commit/244a62593445f5c057e1f098112ca9840cdf8449))
* Doc fixes ([#469](https://github.com/open-feature/open-feature-operator/issues/469)) ([5a7918a](https://github.com/open-feature/open-feature-operator/commit/5a7918a94615621b6c6430e7ddec28c3d39a45e1))
* replace `make deploy-demo` command with a link to the `cloud-native-demo` repo ([#476](https://github.com/open-feature/open-feature-operator/issues/476)) ([fff12a8](https://github.com/open-feature/open-feature-operator/commit/fff12a8dca900478c8f58762ce00ebaf23958dc6))
* update crd version in getting started guide ([#485](https://github.com/open-feature/open-feature-operator/issues/485)) ([eb3b950](https://github.com/open-feature/open-feature-operator/commit/eb3b9501cbfb0f5c2c70337dfc5e499a3b4d755f))

## [0.2.34](https://github.com/open-feature/open-feature-operator/compare/v0.2.33...v0.2.34) (2023-04-13)


### 🧹 Chore

* **deps:** update open-feature/flagd ([#466](https://github.com/open-feature/open-feature-operator/issues/466)) ([3b8d156](https://github.com/open-feature/open-feature-operator/commit/3b8d1564af5fa2991f3e26a0cb8fbf6ff722a9b1))

## [0.2.33](https://github.com/open-feature/open-feature-operator/compare/v0.2.32...v0.2.33) (2023-04-12)


### 🐛 Bug Fixes

* removed old prefix from flagd-proxy provider config ([#463](https://github.com/open-feature/open-feature-operator/issues/463)) ([39a99c6](https://github.com/open-feature/open-feature-operator/commit/39a99c622bb0a7a0fca63d07cc546b2a86f952a5))

## [0.2.32](https://github.com/open-feature/open-feature-operator/compare/v0.2.31...v0.2.32) (2023-04-12)


### 📚 Documentation

* add killercoda demo link ([#413](https://github.com/open-feature/open-feature-operator/issues/413)) ([bbeeea2](https://github.com/open-feature/open-feature-operator/commit/bbeeea27feb3bca805a8be504c6ad447a580582d))


### 🐛 Bug Fixes

* **deps:** update kubernetes packages to v0.26.3 ([#273](https://github.com/open-feature/open-feature-operator/issues/273)) ([abe56e1](https://github.com/open-feature/open-feature-operator/commit/abe56e14305309d4a4c776f4dfa3c8110cd16d23))
* **deps:** update module github.com/go-logr/logr to v1.2.4 ([#428](https://github.com/open-feature/open-feature-operator/issues/428)) ([8d07dab](https://github.com/open-feature/open-feature-operator/commit/8d07dab7eec3f467c84f09512bbf4c4cb066e35f))
* **deps:** update module github.com/onsi/gomega to v1.27.5 ([#357](https://github.com/open-feature/open-feature-operator/issues/357)) ([8624958](https://github.com/open-feature/open-feature-operator/commit/86249582d4bea32f9942c3940590ef399648e6e9))
* **deps:** update module github.com/onsi/gomega to v1.27.6 ([#429](https://github.com/open-feature/open-feature-operator/issues/429)) ([987815c](https://github.com/open-feature/open-feature-operator/commit/987815c05e933d3bfa4020a3864e4493b3b6e80d))
* **deps:** update module github.com/stretchr/testify to v1.8.2 ([#396](https://github.com/open-feature/open-feature-operator/issues/396)) ([f24b6c4](https://github.com/open-feature/open-feature-operator/commit/f24b6c4e536f56cde412827606eacd722637da89))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.14.6 ([#426](https://github.com/open-feature/open-feature-operator/issues/426)) ([0e779e8](https://github.com/open-feature/open-feature-operator/commit/0e779e8d8f53861b0c1a824701ff8668b9fb1907))
* remove unneeded OF namespace prefix from clusterrolebindings ([#453](https://github.com/open-feature/open-feature-operator/issues/453)) ([b23edef](https://github.com/open-feature/open-feature-operator/commit/b23edefc0d403e02dc2279bf275406bd988294f8))
* restrict permissions to only access specific CRB ([#436](https://github.com/open-feature/open-feature-operator/issues/436)) ([6f1f93c](https://github.com/open-feature/open-feature-operator/commit/6f1f93c98c7b8fbee534cc7db63fc396fa5b73c7))
* update flagd proxy env var prefix ([#440](https://github.com/open-feature/open-feature-operator/issues/440)) ([b451d47](https://github.com/open-feature/open-feature-operator/commit/b451d47184c37a5c218ce66a37a448f357dce11f))


### ✨ New Features

* flagd proxy resource ownership ([#442](https://github.com/open-feature/open-feature-operator/issues/442)) ([31b5f7b](https://github.com/open-feature/open-feature-operator/commit/31b5f7bdc62fde593c10797d0f177446aba5d71e))
* introduce debugLogging parameter to FlagSourceConfiguration CRD ([#434](https://github.com/open-feature/open-feature-operator/issues/434)) ([26ae125](https://github.com/open-feature/open-feature-operator/commit/26ae1257f7611ea78dc34247b2f866b0d2043525))
* kube-flagd-proxy deployment ([#412](https://github.com/open-feature/open-feature-operator/issues/412)) ([651c63c](https://github.com/open-feature/open-feature-operator/commit/651c63c5feeb00349db3233554ece2d289e9ccf2))
* migrate flagd startup argument to sources flag ([#427](https://github.com/open-feature/open-feature-operator/issues/427)) ([1c67f34](https://github.com/open-feature/open-feature-operator/commit/1c67f34dca6a6f58e09a7e8b56ce2a2523c1d260))
* **test:** substitute kuttl to bash e2e test ([#411](https://github.com/open-feature/open-feature-operator/issues/411)) ([ff199f1](https://github.com/open-feature/open-feature-operator/commit/ff199f1ae3c72d5472937eef7c2409b186bbb314))


### 🧹 Chore

* add unit tests to pod webhook ([#419](https://github.com/open-feature/open-feature-operator/issues/419)) ([4290978](https://github.com/open-feature/open-feature-operator/commit/42909784b6a3a0642f07b5c5e093f9d4c549a21c))
* attempt renovate fix ([48b6c7f](https://github.com/open-feature/open-feature-operator/commit/48b6c7fabce54270b06f53c033801be5ec100633))
* attempt versioning fix in test ([58d0145](https://github.com/open-feature/open-feature-operator/commit/58d0145f0a3ae1d67be002961faf82d8ef050015))
* **deps:** update actions/setup-go action to v4 ([#398](https://github.com/open-feature/open-feature-operator/issues/398)) ([ee9ecb9](https://github.com/open-feature/open-feature-operator/commit/ee9ecb9d693cdccbcac38a5c6c97d20a8a9c769f))
* **deps:** update dependency open-feature/flagd to v0.2.1 ([#462](https://github.com/open-feature/open-feature-operator/issues/462)) ([d2d53b7](https://github.com/open-feature/open-feature-operator/commit/d2d53b75791eef407ba0b1dd5377aff8277301ea))
* **deps:** update docker/login-action digest to 65b78e6 ([#421](https://github.com/open-feature/open-feature-operator/issues/421)) ([8d2ebe2](https://github.com/open-feature/open-feature-operator/commit/8d2ebe27193379fb54e5a39455e8db787f8eae89))
* **deps:** update docker/metadata-action digest to 3f6690a ([#432](https://github.com/open-feature/open-feature-operator/issues/432)) ([991b2bd](https://github.com/open-feature/open-feature-operator/commit/991b2bd3c320b8b576812f72a2d98ab30436f6c8))
* **deps:** update golang docker tag to v1.20.3 ([#445](https://github.com/open-feature/open-feature-operator/issues/445)) ([b8f6c5b](https://github.com/open-feature/open-feature-operator/commit/b8f6c5b9e7bfc986f2208b2d7a2f402d7210ca7a))
* **deps:** update module golang.org/x/net to v0.8.0 ([#397](https://github.com/open-feature/open-feature-operator/issues/397)) ([096c889](https://github.com/open-feature/open-feature-operator/commit/096c889c87e80b5cfef0254869dc1e096ee23ad8))
* **deps:** update module golang.org/x/net to v0.9.0 ([#451](https://github.com/open-feature/open-feature-operator/issues/451)) ([4cbe4f1](https://github.com/open-feature/open-feature-operator/commit/4cbe4f1a02517d89a53fde6ca1a5861da2691747))
* **deps:** update open-feature/flagd ([#457](https://github.com/open-feature/open-feature-operator/issues/457)) ([db9af7a](https://github.com/open-feature/open-feature-operator/commit/db9af7a02dbfcd4be10b170dab4bb5e65614221f))
* **deps:** update open-feature/flagd to v0.5.0 ([#422](https://github.com/open-feature/open-feature-operator/issues/422)) ([6846aa2](https://github.com/open-feature/open-feature-operator/commit/6846aa206a9ffb4aa9b1cff1ca7078b93ede927c))
* fix renovate config, add recommended preset ([#418](https://github.com/open-feature/open-feature-operator/issues/418)) ([78c5970](https://github.com/open-feature/open-feature-operator/commit/78c597024241158ebf2e9b07e82610766efd85de))
* improve container build layer caching ([#414](https://github.com/open-feature/open-feature-operator/issues/414)) ([3212eba](https://github.com/open-feature/open-feature-operator/commit/3212eba809744c8dc1c94d8bf558523a0fbbf326))
* increase backoffLimit for inject-flagd ([#423](https://github.com/open-feature/open-feature-operator/issues/423)) ([29d7cf0](https://github.com/open-feature/open-feature-operator/commit/29d7cf069d68ce2b81718b0297194b3ba53c3ed9))
* introduce additional unit tests for api packages ([#420](https://github.com/open-feature/open-feature-operator/issues/420)) ([5ba5bc9](https://github.com/open-feature/open-feature-operator/commit/5ba5bc97faa8bf18a07a380d685c518f6e093145))
* refactor admission webhook tests ([#409](https://github.com/open-feature/open-feature-operator/issues/409)) ([29c7c28](https://github.com/open-feature/open-feature-operator/commit/29c7c28b4a6fb76bc565e32f46d0ab74fc2e5371))
* refactor pod webhook mutator ([#410](https://github.com/open-feature/open-feature-operator/issues/410)) ([2a86b03](https://github.com/open-feature/open-feature-operator/commit/2a86b032888fef4bd3e7d93e3a5cb1cc376fcd22))
* refactored component test using fake client ([#435](https://github.com/open-feature/open-feature-operator/issues/435)) ([08a50ac](https://github.com/open-feature/open-feature-operator/commit/08a50accff516be1f8226c4f1051eef8843c9190))
* remove ignored renovate paths ([#441](https://github.com/open-feature/open-feature-operator/issues/441)) ([c1d8929](https://github.com/open-feature/open-feature-operator/commit/c1d89291d75ef0d594a071ef5055b55a404d9b73))
* reorder containers in e2e assertion ([1d895c3](https://github.com/open-feature/open-feature-operator/commit/1d895c33c32cefc9858cf2ef0f283d1ba62a4f00))
* split controllers to separate packages + cover them with unit tests ([#404](https://github.com/open-feature/open-feature-operator/issues/404)) ([6ed4cef](https://github.com/open-feature/open-feature-operator/commit/6ed4cef4a7d1ec889300459f73e930d4b6d2ba6f))
* troubleshoot renovate ([de4ac14](https://github.com/open-feature/open-feature-operator/commit/de4ac1475717201ec6a828ffc7700d3c28de4d33))
* troubleshoot renovate ([89a7b5b](https://github.com/open-feature/open-feature-operator/commit/89a7b5b9890f127a5af1d321f40b8f2a8635fcb5))
* troubleshoot renovate ([244bd3a](https://github.com/open-feature/open-feature-operator/commit/244bd3ade508c476a9783c9ee11d608e2536bb9f))
* troubleshoot renovate ([eafa670](https://github.com/open-feature/open-feature-operator/commit/eafa6702e1663a02b24b48e3b61ea6252b2a9b40))
* troubleshoot renovate ([c3d9523](https://github.com/open-feature/open-feature-operator/commit/c3d95232d0f1ca6e8c898ffffb165537462fe2e9))
* troubleshoot renovatge ([35054cb](https://github.com/open-feature/open-feature-operator/commit/35054cb6917dcacbafb9fbccb00a85493922f245))
* troubleshoot renvoate ([7ac3c90](https://github.com/open-feature/open-feature-operator/commit/7ac3c90a358baf6f0dd00bd2f7295665ebf46a59))
* update codeowners to use cloud native team ([6133060](https://github.com/open-feature/open-feature-operator/commit/613306011016a3cbb7fbc23a2273aecfd26a3bbf))
* update flagd renovate detection ([#439](https://github.com/open-feature/open-feature-operator/issues/439)) ([3d1540c](https://github.com/open-feature/open-feature-operator/commit/3d1540c67c7d43c69feb61654b7d2a3c8a72a5a1))
* update renovate config to watch the assert yaml directly ([9ef25a0](https://github.com/open-feature/open-feature-operator/commit/9ef25a0abbdeb15666679fd43d4f2c032b825722))
* use renovate to bump flagd version ([#395](https://github.com/open-feature/open-feature-operator/issues/395)) ([fd5b072](https://github.com/open-feature/open-feature-operator/commit/fd5b072214f1c3c74dfc4bc53ca1ff6c14d72ffa))

## [0.2.31](https://github.com/open-feature/open-feature-operator/compare/v0.2.30...v0.2.31) (2023-03-16)


### 📚 Documentation

* fix rendering issue with operator resource config table ([#401](https://github.com/open-feature/open-feature-operator/issues/401)) ([71ea8a6](https://github.com/open-feature/open-feature-operator/commit/71ea8a68bbb97052822552ffce3c498c3da0e52d))


### 🐛 Bug Fixes

* update flagd version ([#402](https://github.com/open-feature/open-feature-operator/issues/402)) ([dc6aa3c](https://github.com/open-feature/open-feature-operator/commit/dc6aa3c3dd9fec6c508b34608384247b63b42eeb))

## [0.2.30](https://github.com/open-feature/open-feature-operator/compare/v0.2.29...v0.2.30) (2023-03-16)


### 📚 Documentation

* add AND operator to sequential commands ([#368](https://github.com/open-feature/open-feature-operator/issues/368)) ([6f73a62](https://github.com/open-feature/open-feature-operator/commit/6f73a6214d87771f9555469fe4d60dbb6d301198))


### ✨ New Features

* enable flagd probes ([#390](https://github.com/open-feature/open-feature-operator/issues/390)) ([41efb15](https://github.com/open-feature/open-feature-operator/commit/41efb155994b3cfb768cc39e59bfc09781c57f2e))
* improve deployment pattern ([#344](https://github.com/open-feature/open-feature-operator/issues/344)) ([572ba96](https://github.com/open-feature/open-feature-operator/commit/572ba961912ada2c07eb6143925d16ab6a6a85a3))


### 🐛 Bug Fixes

* **deps:** update module sigs.k8s.io/controller-runtime to v0.14.5 ([#279](https://github.com/open-feature/open-feature-operator/issues/279)) ([8a80bff](https://github.com/open-feature/open-feature-operator/commit/8a80bff886af404e897e6a247cea2f4c88d88499))


### 🧹 Chore

* add additional sections to the release notes ([4bec5af](https://github.com/open-feature/open-feature-operator/commit/4bec5af5fc5fc589d920f0c17a1213a036b558a0))
* add artifact hub metadata ([#372](https://github.com/open-feature/open-feature-operator/issues/372)) ([c6f539f](https://github.com/open-feature/open-feature-operator/commit/c6f539f5bdd9dc18ac197eb3303d91131e863011))
* **deps:** update dependency open-feature/flagd to v0.4.0 ([#342](https://github.com/open-feature/open-feature-operator/issues/342)) ([0640f46](https://github.com/open-feature/open-feature-operator/commit/0640f469daa3c0adce920bb73e901fe83bc275e7))
* **deps:** update dependency open-feature/flagd to v0.4.1 ([#373](https://github.com/open-feature/open-feature-operator/issues/373)) ([756cf7a](https://github.com/open-feature/open-feature-operator/commit/756cf7a96c05fdfa8ffa2bf933225b84af400e37))
* **deps:** update dependency open-feature/flagd to v0.4.4 ([#400](https://github.com/open-feature/open-feature-operator/issues/400)) ([3e0a666](https://github.com/open-feature/open-feature-operator/commit/3e0a666f2824071c49250a4467d62b96a5af5ee7))
* **deps:** update docker/login-action digest to 219c305 ([#365](https://github.com/open-feature/open-feature-operator/issues/365)) ([ee84954](https://github.com/open-feature/open-feature-operator/commit/ee849546322516019ea19a205c22c4ee38ac36ed))
* **deps:** update docker/metadata-action digest to 766400c ([#267](https://github.com/open-feature/open-feature-operator/issues/267)) ([38a1464](https://github.com/open-feature/open-feature-operator/commit/38a14644e687b928e51d1350f6d57ef9d493330c))
* **deps:** update docker/metadata-action digest to 9ec57ed ([#366](https://github.com/open-feature/open-feature-operator/issues/366)) ([884d444](https://github.com/open-feature/open-feature-operator/commit/884d44422ad7bfa28a8fb88156cd66e252e2eba5))
* **deps:** update gcr.io/kubebuilder/kube-rbac-proxy docker tag to v0.14.0 ([#376](https://github.com/open-feature/open-feature-operator/issues/376)) ([708e4bc](https://github.com/open-feature/open-feature-operator/commit/708e4bc44d8493d4f0aaa7f7036c2b7ecd2efd32))
* **deps:** update ghcr.io/open-feature/flagd docker tag to v0.4.4 ([#381](https://github.com/open-feature/open-feature-operator/issues/381)) ([a253761](https://github.com/open-feature/open-feature-operator/commit/a253761af8565fdcf6e6f9ca92c740f25b4b0620))
* **deps:** update golang docker tag to v1.20.2 ([#374](https://github.com/open-feature/open-feature-operator/issues/374)) ([e2de529](https://github.com/open-feature/open-feature-operator/commit/e2de52997b44835a4a8515e9fd37c976d3539272))
* e2e test for openfeature.dev/enabled annotation set to false ([#375](https://github.com/open-feature/open-feature-operator/issues/375)) ([b03fb14](https://github.com/open-feature/open-feature-operator/commit/b03fb145e317f987727d76b98041fa783e5c2202))
* improve formatting and content ([#384](https://github.com/open-feature/open-feature-operator/issues/384)) ([c5a6a32](https://github.com/open-feature/open-feature-operator/commit/c5a6a32f0ccccc6449fc581de08c283434c1adb6))
* remove unneeded conversion webhooks + introduce unit tests for conversion functions ([#385](https://github.com/open-feature/open-feature-operator/issues/385)) ([dd34801](https://github.com/open-feature/open-feature-operator/commit/dd34801fd71ac4f1e6c0b0f39f78ddf738f5601d))

## [0.2.29](https://github.com/open-feature/open-feature-operator/compare/v0.2.28...v0.2.29) (2023-02-23)


### Features

* add log format configuration options through helm chart ([#346](https://github.com/open-feature/open-feature-operator/issues/346)) ([bcef736](https://github.com/open-feature/open-feature-operator/commit/bcef7368fc4905b351f81f5dfa10eb4c26bf8764))
* Introduced context to the readyz endpoint, added wait to test suite ([#336](https://github.com/open-feature/open-feature-operator/issues/336)) ([ed81c02](https://github.com/open-feature/open-feature-operator/commit/ed81c0284f8d759eb228d3af7030efb0b94ee280))


### Bug Fixes

* Security issues ([#348](https://github.com/open-feature/open-feature-operator/issues/348)) ([5bd0b19](https://github.com/open-feature/open-feature-operator/commit/5bd0b192a5db7f1557e1161e4bb425bbf0e31e2a))
* set defaultTag to INPUT_FLAGD_VERSION ([#332](https://github.com/open-feature/open-feature-operator/issues/332)) ([23547a1](https://github.com/open-feature/open-feature-operator/commit/23547a1e155e0cde2f085882bfd43128681466cd))

## [0.2.28](https://github.com/open-feature/open-feature-operator/compare/v0.2.27...v0.2.28) (2023-01-28)


### Bug Fixes

* mount dirs not files ([#326](https://github.com/open-feature/open-feature-operator/issues/326)) ([089ab3c](https://github.com/open-feature/open-feature-operator/commit/089ab3c48c0937e64060057e43ff07cf8fd47f67))

## [0.2.27](https://github.com/open-feature/open-feature-operator/compare/v0.2.26...v0.2.27) (2023-01-27)


### Features

* default sync provider configuration ([#320](https://github.com/open-feature/open-feature-operator/issues/320)) ([7cba7e1](https://github.com/open-feature/open-feature-operator/commit/7cba7e14c223a083f02ff8313b899583253120f3))


### Bug Fixes

* gave configmaps volume mounts a subpath to allow for multiple mounts ([#321](https://github.com/open-feature/open-feature-operator/issues/321)) ([2ec454c](https://github.com/open-feature/open-feature-operator/commit/2ec454c036149ebeaf34f81cbf4ad7895f0bb995))
* uniqueness of featureflagconfiguration file path ([#323](https://github.com/open-feature/open-feature-operator/issues/323)) ([2b10945](https://github.com/open-feature/open-feature-operator/commit/2b109452893abd053640ffbb9c79b834b78feb7b))

## [0.2.26](https://github.com/open-feature/open-feature-operator/compare/v0.2.25...v0.2.26) (2023-01-26)


### Bug Fixes

* **deps:** update module github.com/open-feature/schemas to v0.2.8 ([#269](https://github.com/open-feature/open-feature-operator/issues/269)) ([ed48060](https://github.com/open-feature/open-feature-operator/commit/ed48060b1f9e591ddadca4f9478728a823e10685))

## [0.2.25](https://github.com/open-feature/open-feature-operator/compare/v0.2.24...v0.2.25) (2023-01-25)


### Features

* Helm configuration ([#304](https://github.com/open-feature/open-feature-operator/issues/304)) ([99edfeb](https://github.com/open-feature/open-feature-operator/commit/99edfeb8c32ada435f830c6799540ebdf3b5fcdd))


### Bug Fixes

* removed duplicate config map generation, resolve permissions issue ([#305](https://github.com/open-feature/open-feature-operator/issues/305)) ([eec16af](https://github.com/open-feature/open-feature-operator/commit/eec16af28eb963a3d0f276d382e808079e663a50))
* update x/net for CVE-2022-41721 ([#301](https://github.com/open-feature/open-feature-operator/issues/301)) ([bbe9837](https://github.com/open-feature/open-feature-operator/commit/bbe983786ff74b59046b95082d79f71089fe2b67))

## [0.2.24](https://github.com/open-feature/open-feature-operator/compare/v0.2.23...v0.2.24) (2023-01-16)


### Features

* backfill flagd-kubernetes-sync cluster role binding on startup ([#295](https://github.com/open-feature/open-feature-operator/pull/295))
* decouple feature flag spec from flagd config ([#276](https://github.com/open-feature/open-feature-operator/pull/276))


### Features

* upgrade flagd to v0.3.0 ([20571e1](https://github.com/open-feature/open-feature-operator/commit/20571e1018e102ffbcf01b2518fcbf8b66a287be))

## [0.2.22](https://github.com/open-feature/open-feature-operator/compare/v0.2.21...v0.2.22) (2022-12-16)


### Bug Fixes

* **deps:** update module go.uber.org/zap to v1.24.0 ([#268](https://github.com/open-feature/open-feature-operator/issues/268)) ([b7bdde8](https://github.com/open-feature/open-feature-operator/commit/b7bdde8944446621751e6ef70e6b0f0646adee21))
* Version fix ([#284](https://github.com/open-feature/open-feature-operator/issues/284)) ([a9c6f15](https://github.com/open-feature/open-feature-operator/commit/a9c6f154589f1e00e60883c229b3ee29d7d2e9aa))

## [0.2.21](https://github.com/open-feature/open-feature-operator/compare/v0.2.20...v0.2.21) (2022-12-16)


### Features

* add ff shortname, commit httpSyncConfiguration ([11e4652](https://github.com/open-feature/open-feature-operator/commit/11e46528fcd06cdc0c8e6f46944656224cd97441))
* introduce configurable resource limits for flagd sidecar ([e4affcf](https://github.com/open-feature/open-feature-operator/commit/e4affcfb0ccf13dc0406ef1c21c2b884a836f71f))


### Bug Fixes

* **deps:** update github.com/open-feature/schemas digest to 302d0fa ([#246](https://github.com/open-feature/open-feature-operator/issues/246)) ([7d22374](https://github.com/open-feature/open-feature-operator/commit/7d22374afb7a5e2e166550544d327ec7b5b3d1bf))
* **deps:** update kubernetes packages to v0.25.4 ([75bab2d](https://github.com/open-feature/open-feature-operator/commit/75bab2d441c945d51f17f0d32195a217072c3c15))
* include release tag in helm charts publishing ([2746716](https://github.com/open-feature/open-feature-operator/commit/27467164dcd05b0220e0857bf79e42d62e7a40a9))

## [0.2.20](https://github.com/open-feature/open-feature-operator/compare/v0.2.19...v0.2.20) (2022-11-18)


### Bug Fixes

* **deps:** update module sigs.k8s.io/controller-runtime to v0.13.1 ([edeffcd](https://github.com/open-feature/open-feature-operator/commit/edeffcd3ef6fe9a8d52d0d5c414512ef8cd80629))

## [0.2.19](https://github.com/open-feature/open-feature-operator/compare/v0.2.18...v0.2.19) (2022-11-15)


### Features

* introduced v1beta1 of featureflagconfiguration CRD with conversion webhook to v1alpha1 ([a45bdef](https://github.com/open-feature/open-feature-operator/commit/a45bdef5eec87738ce731af5825daffeb69eb6cb))
* structured the featureflagconfiguration CRD ([b056c7c](https://github.com/open-feature/open-feature-operator/commit/b056c7cdd76f4653c1a728342687beaa8279e314))

## [0.2.18](https://github.com/open-feature/open-feature-operator/compare/v0.2.17...v0.2.18) (2022-11-10)


### Bug Fixes

* nil pointer dereference ([#216](https://github.com/open-feature/open-feature-operator/issues/216)) ([d975066](https://github.com/open-feature/open-feature-operator/commit/d975066f96a5f9caf8af8d513076480a33943257))

## [0.2.17](https://github.com/open-feature/open-feature-operator/compare/v0.2.16...v0.2.17) (2022-11-07)


### Bug Fixes

* **deps:** update github.com/open-feature/schemas digest to d638ecf ([a984836](https://github.com/open-feature/open-feature-operator/commit/a98483696f467270783858046132f02b3d338ac2))
* for helm issues ([#206](https://github.com/open-feature/open-feature-operator/issues/206)) ([39febd7](https://github.com/open-feature/open-feature-operator/commit/39febd76d1b996afdbc24399bcd08b502621c6cc))

## [0.2.16](https://github.com/open-feature/open-feature-operator/compare/v0.2.15...v0.2.16) (2022-10-27)


### Bug Fixes

* resolve issue with templated DNS name in cert ([65068df](https://github.com/open-feature/open-feature-operator/commit/65068df3019312a965271e50c52bbb90b68665c0))

## [0.2.15](https://github.com/open-feature/open-feature-operator/compare/v0.2.14...v0.2.15) (2022-10-25)


### Bug Fixes

* artifact name and output file ([#187](https://github.com/open-feature/open-feature-operator/issues/187)) ([4dee157](https://github.com/open-feature/open-feature-operator/commit/4dee157d44c20fc925f9e33dbaae16c18f3d9b48))
* remove redundant name ([#189](https://github.com/open-feature/open-feature-operator/issues/189)) ([664bd73](https://github.com/open-feature/open-feature-operator/commit/664bd7314e376b23a01247b5c027c04a9ac26329))

## [0.2.14](https://github.com/open-feature/open-feature-operator/compare/v0.2.13...v0.2.14) (2022-10-25)


### Bug Fixes

* add sbom to ouput name ([#182](https://github.com/open-feature/open-feature-operator/issues/182)) ([5e939a8](https://github.com/open-feature/open-feature-operator/commit/5e939a8f67fbd095c18a6a2172bb856fe61dd173))

## [0.2.13](https://github.com/open-feature/open-feature-operator/compare/v0.2.12...v0.2.13) (2022-10-25)


### Bug Fixes

* set sbom dir ([#180](https://github.com/open-feature/open-feature-operator/issues/180)) ([616272d](https://github.com/open-feature/open-feature-operator/commit/616272d6d693115a22839cf52eb8fd448609ad6c))

## [0.2.12](https://github.com/open-feature/open-feature-operator/compare/v0.2.11...v0.2.12) (2022-10-25)


### Bug Fixes

* set sbom dir ([#178](https://github.com/open-feature/open-feature-operator/issues/178)) ([143adf9](https://github.com/open-feature/open-feature-operator/commit/143adf910fe15a8b8af31dff48743352ab203d83))

## [0.2.11](https://github.com/open-feature/open-feature-operator/compare/v0.2.10...v0.2.11) (2022-10-25)


### Bug Fixes

* Upload sbom ([#175](https://github.com/open-feature/open-feature-operator/issues/175)) ([813c646](https://github.com/open-feature/open-feature-operator/commit/813c6469ecc18101f60c593282ed32d7579f5880))
* Upload sbom by name ([#176](https://github.com/open-feature/open-feature-operator/issues/176)) ([7d0fcd0](https://github.com/open-feature/open-feature-operator/commit/7d0fcd0ba7eeee1b2424189c7e5f5f92bc1fffac))

## [0.2.10](https://github.com/open-feature/open-feature-operator/compare/v0.2.9...v0.2.10) (2022-10-25)


### Bug Fixes

* correcrt needs in asset release ([5ed4571](https://github.com/open-feature/open-feature-operator/commit/5ed45718ca189a15f7cdf4f8ddfc5864f189b1ce))

## [0.2.9](https://github.com/open-feature/open-feature-operator/compare/v0.2.8...v0.2.9) (2022-10-25)


### Bug Fixes

* Package signing should happen in the oci workflow. ([a04a110](https://github.com/open-feature/open-feature-operator/commit/a04a110e29b1725a66d0f4b529741947ebb7c798))

## [0.2.8](https://github.com/open-feature/open-feature-operator/compare/v0.2.7...v0.2.8) (2022-10-25)


### Bug Fixes

* package signing fixes ([36597f4](https://github.com/open-feature/open-feature-operator/commit/36597f484c85effd6a993f44b97fcd541d34c515))

## [0.2.7](https://github.com/open-feature/open-feature-operator/compare/v0.2.6...v0.2.7) (2022-10-25)


### Features

* adding artifacthub information ([#144](https://github.com/open-feature/open-feature-operator/issues/144)) ([65a5244](https://github.com/open-feature/open-feature-operator/commit/65a524445d1db8bb5608b88282a4d97a9bb6b74f))
* builds helm chart ([#137](https://github.com/open-feature/open-feature-operator/issues/137)) ([1525421](https://github.com/open-feature/open-feature-operator/commit/1525421229d43b17636dddb65d7b124e6477fe79))

## [0.2.7](https://github.com/open-feature/open-feature-operator/compare/v0.2.6...v0.2.7) (2022-10-24)


### Features

* adding artifacthub information ([#144](https://github.com/open-feature/open-feature-operator/issues/144)) ([65a5244](https://github.com/open-feature/open-feature-operator/commit/65a524445d1db8bb5608b88282a4d97a9bb6b74f))
* builds helm chart ([#137](https://github.com/open-feature/open-feature-operator/issues/137)) ([1525421](https://github.com/open-feature/open-feature-operator/commit/1525421229d43b17636dddb65d7b124e6477fe79))

## [0.2.6](https://github.com/open-feature/open-feature-operator/compare/v0.2.5...v0.2.6) (2022-10-24)


### Features

* adding artifacthub information ([#144](https://github.com/open-feature/open-feature-operator/issues/144)) ([65a5244](https://github.com/open-feature/open-feature-operator/commit/65a524445d1db8bb5608b88282a4d97a9bb6b74f))
* builds helm chart ([#137](https://github.com/open-feature/open-feature-operator/issues/137)) ([1525421](https://github.com/open-feature/open-feature-operator/commit/1525421229d43b17636dddb65d7b124e6477fe79))


### Bug Fixes

* CVE-2022-32149 ([015c19a](https://github.com/open-feature/open-feature-operator/commit/015c19ac4455673902c365111816b021f893c485))

## [0.2.6](https://github.com/open-feature/open-feature-operator/compare/v0.2.5...v0.2.6) (2022-10-20)


### Bug Fixes

* CVE-2022-32149 ([015c19a](https://github.com/open-feature/open-feature-operator/commit/015c19ac4455673902c365111816b021f893c485))

## [0.2.5](https://github.com/open-feature/open-feature-operator/compare/v0.2.4...v0.2.5) (2022-10-19)


### Features

* stop creation and mounting of flagd-config config map in case of kubernetes sync-provider ([#126](https://github.com/open-feature/open-feature-operator/issues/126)) ([a1d9fe2](https://github.com/open-feature/open-feature-operator/commit/a1d9fe276a37259d01e6ed6239c0ebcd3a1e6611))

## [0.2.4](https://github.com/open-feature/open-feature-operator/compare/v0.2.3...v0.2.4) (2022-10-18)


### Bug Fixes

* build and push to docker registry with tag as current release ([#123](https://github.com/open-feature/open-feature-operator/issues/123)) ([d4abda1](https://github.com/open-feature/open-feature-operator/commit/d4abda119e4a7c2dab7a2e0d335d44b1df07ec62))

## [0.2.3](https://github.com/open-feature/open-feature-operator/compare/v0.2.2...v0.2.3) (2022-10-18)


### Bug Fixes

* build and push to docker registry on tag creation ([#121](https://github.com/open-feature/open-feature-operator/issues/121)) ([27c6f9c](https://github.com/open-feature/open-feature-operator/commit/27c6f9cbc298fb8bf578464e4c3f9f07402b87ab))

## [0.2.2](https://github.com/open-feature/open-feature-operator/compare/v0.2.1...v0.2.2) (2022-10-14)


### Bug Fixes

* bump flagd version to include change detection fix ([421cab6](https://github.com/open-feature/open-feature-operator/commit/421cab651f6ebe2ece1380fda7dc24d92838d6b5))

## [0.2.1](https://github.com/open-feature/open-feature-operator/compare/v0.2.0...v0.2.1) (2022-10-13)


### Features

* metrics ([#111](https://github.com/open-feature/open-feature-operator/issues/111)) ([6016669](https://github.com/open-feature/open-feature-operator/commit/6016669ec46984d127951ee5d0ff02e7685f4d80))
* pr github action workflow ([#96](https://github.com/open-feature/open-feature-operator/issues/96)) ([a719f8a](https://github.com/open-feature/open-feature-operator/commit/a719f8a33abc9b9599987314282cc4e7ac202d67))


### Bug Fixes

* include assets in release ([#109](https://github.com/open-feature/open-feature-operator/issues/109)) ([b835abb](https://github.com/open-feature/open-feature-operator/commit/b835abb48ae8ca3c9c63abd51ae5614a4068c003))

## [0.2.0](https://github.com/open-feature/open-feature-operator/compare/v0.1.1...v0.2.0) (2022-10-10)


### ⚠ BREAKING CHANGES

* bump flagd version to 0.2.0 (connect refactor) (#97)

### Features

* bump flagd version to 0.2.0 (connect refactor) ([#97](https://github.com/open-feature/open-feature-operator/issues/97)) ([8118b9f](https://github.com/open-feature/open-feature-operator/commit/8118b9fcbaf0d3c66d6869369add645e388989de))


### Bug Fixes

* upgrade dependencies with vulnerabilities ([#90](https://github.com/open-feature/open-feature-operator/issues/90)) ([58cdd4e](https://github.com/open-feature/open-feature-operator/commit/58cdd4ee7c6989e44258bad3e9ed75a3bb465cae))
