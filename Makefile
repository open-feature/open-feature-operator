RELEASE_REGISTRY?=ghcr.io/openfeature
TAG?=latest
RELEASE_NAME?=operator
RELEASE_IMAGE?=$(RELEASE_NAME):$(TAG)
ARCH?=amd64
IMG?=$(RELEASE_REGISTRY)/$(RELEASE_IMAGE)
# customize overlay to be used in the build, DEFAULT or HELM
KUSTOMIZE_OVERLAY ?= DEFAULT
CHART_VERSION=v0.2.35# x-release-please-version
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26.1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: unit-test
unit-test: manifests fmt vet generate envtest ## Run tests.
	go test ./... -v -short -coverprofile cover.out

## Requires the operator to be deployed
.PHONY: e2e-test
e2e-test: manifests generate fmt vet
	kubectl -n open-feature-operator-system apply -f ./test/e2e/e2e.yml
	kubectl wait --for=condition=Available=True deploy --all -n 'open-feature-operator-system'
	./test/e2e/run.sh

.PHONY: e2e-test-kuttl #these tests should run on a real cluster!
e2e-test-kuttl:
	kubectl kuttl test --start-kind=false ./test/e2e/kuttl --config=./kuttl-test.yaml

.PHONY: e2e-test-kuttl-local #these tests should run on a real cluster!
e2e-test-kuttl-local:
	kubectl kuttl test --start-kind=false ./test/e2e/kuttl/scenarios --config=./kuttl-test-local.yaml

.PHONY: lint
lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	${GOPATH}/bin/golangci-lint run --deadline=3m --timeout=3m ./... # Run linters

.PHONY: generate-crdocs
generate-crdocs: kustomize crdocs
	$(KUSTOMIZE) build config/crd > tmpcrd.yaml
	perl -i -pe "s/\_/\&lowbar;/gm" tmpcrd.yaml #escape _
	perl -i -pe "s/\</\&lt;/gm" tmpcrd.yaml #escape <
	perl -i -pe "s/\>/\&gt;/gm" tmpcrd.yaml #escape <
	$(CRDOC) --resources tmpcrd.yaml --output docs/crds.md


##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: clean  ## Build docker image with the manager.
	DOCKER_BUILDKIT=1 docker build \
		-t $(IMG)-$(ARCH)  \
		--platform linux/$(ARCH) \
		.
	docker tag $(IMG)-$(ARCH) $(IMG)

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push $(IMG)

.PHONY: clean
clean:
	rm -rf ./bin

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: release-manifests
release-manifests: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	mkdir -p config/rendered/
	@if [ ${KUSTOMIZE_OVERLAY} = DEFAULT ]; then\
		echo building default overlay;\
        $(KUSTOMIZE) build config/default > config/rendered/release.yaml;\
    fi
	@if [ ${KUSTOMIZE_OVERLAY} = HELM ]; then\
		echo building helm overlay;\
		$(KUSTOMIZE) build config/overlays/helm -o chart/open-feature-operator/templates/ ;\
    fi
	
.PHONY: deploy
deploy: generate manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: generate ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-operator
deploy-operator:
	kubectl create ns 'open-feature-operator-system' --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
	kubectl wait --for=condition=Available=True deploy --all -n 'cert-manager'
	make deploy
	kubectl wait --for=condition=Available=True deploy --all -n 'open-feature-operator-system'

.PHONY: build-deploy-operator
build-deploy-operator:
	make docker-build
	make docker-push
	make deploy-operator

deploy-demo:
	kubectl apply -f https://raw.githubusercontent.com/open-feature/playground/main/config/k8s/end-to-end.yaml
	kubectl wait -l app=open-feature-demo --for=condition=Available=True deploy
	kubectl port-forward service/open-feature-demo-service 30000:30000

delete-demo-deployment:
	kubectl delete -f https://raw.githubusercontent.com/open-feature/playground/main/config/k8s/end-to-end.yaml

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
HELM ?= $(LOCALBIN)/HELM
ENVTEST ?= $(LOCALBIN)/setup-envtest
CRDOC ?= $(LOCALBIN)/crdoc

## Tool Versions
# renovate: datasource=github-tags depName=kubernetes-sigs/kustomize
KUSTOMIZE_VERSION ?= v4.5.7
# renovate: datasource=github-releases depName=kubernetes-sigs/controller-tools
CONTROLLER_TOOLS_VERSION ?= v0.10.0
CRDOC_VERSION ?= v0.6.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	[ -e "$(KUSTOMIZE)" ] && rm -rf "$(KUSTOMIZE)" || true
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: crdocs
crdocs: $(CRDOC) ## Download crdoc locally if necessary.
$(CRDOC): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install fybrik.io/crdoc@$(CRDOC_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

OSARCH=$(shell ./hack/get-os.sh)
HELM = $(shell pwd)/bin/$(OSARCH)/helm
HELM_INSTALLER ?= "https://get.helm.sh/helm-v3.10.1-$(OSARCH).tar.gz"
.PHONY: helm
helm: $(HELM) ## Download helm locally if necessary.
$(HELM): $(LOCALBIN)
	[ -e "$(HELM)" ] && rm -rf "$(HELM)" || true
	cd $(LOCALBIN) && curl -s $(HELM_INSTALLER) | tar -xzf - -C $(LOCALBIN)

.PHONY: set-helm-overlay
set-helm-overlay:
	${eval KUSTOMIZE_OVERLAY = HELM}

helm-package: set-helm-overlay generate release-manifests helm
	$(HELM) package --version $(CHART_VERSION) chart/open-feature-operator
	mkdir -p charts && mv open-feature-operator-*.tgz charts
	$(HELM) repo index --url https://open-feature.github.io/open-feature-operator/charts charts
	mv charts/index.yaml index.yaml

install-mockgen:
	go install github.com/golang/mock/mockgen@v1.6.0
mockgen: install-mockgen
	mockgen -source=controllers/common/flagd-injector.go -destination=controllers/common/mock/flagd-injector.go -package=commonmock
