PACKAGE_NAME := github.com/appoptics/appoptics-kubernetes-controller
REGISTRY := hub.docker.com/r/appoptics
APP_NAME := appoptics_kubernetes_controller
IMAGE_TAG ?= 0.1
GOPATH ?= $HOME/go
HACK_DIR ?= hack
BUILD_TAG := build
BINPATH := ./bin
NAMESPACE := default

# Path to dockerfiles directory
DOCKERFILES := $(HACK_DIR)/build
# A list of all types.go files in pkg/apis
TYPES_FILES := $(shell find pkg/apis -name types.go)

# Go build flags
GOOS := linux
GOARCH := amd64
GIT_COMMIT := $(shell git rev-parse HEAD)
GOLDFLAGS := -ldflags "-X $(PACKAGE_NAME)/pkg/util.AppGitCommit=${GIT_COMMIT} -X $(PACKAGE_NAME)/pkg/util.AppVersion=${IMAGE_TAG}"

.PHONY: verify build docker_build push generate generate_verify \
	appoptics_controller go_test go_fmt e2e_test go_verify   \
	docker_build docker_push

# Alias targets
###############

build: go_dep generate generate_verify go_test appoptics_controller # docker_build
verify: generate_verify go_verify
#push: build docker_push

# Code generation
#################
# This target runs all required generators against our API types.
generate: $(TYPES_FILES)
	$(HACK_DIR)/update-codegen.sh

generate_verify:
	$(HACK_DIR)/verify-codegen.sh

# Go targets
#################
go_verify: go_fmt go_test

go_dep:
	dep ensure -v

appoptics_controller:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-a -tags netgo \
		-o $(BINPATH)/${APP_NAME}-$(GOOS)_$(GOARCH) \
		$(GOLDFLAGS) \
		./

go_test:
	go test -v \
	    -race \
		-cover \
		-coverprofile=coverage.out \
		$$(go list ./... | \
			grep -v '/vendor/' | \
			grep -v '/pkg/client' \
		)

coverage: go_test
	go tool cover -html=coverage.out

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

add_token:
	kubectl --namespace $(NAMESPACE) create secret generic appoptics --from-literal=token=$(APPOPTICS_TOKEN)

# Docker targets
################
#docker_build:
#	docker build \
#		--build-arg VCS_REF=$(GIT_COMMIT) \
#		-t $(REGISTRY)/$(APP_NAME):$(BUILD_TAG) \
#		-f $(DOCKERFILES)/Dockerfile \
#		./
#
#docker_push:
#	set -e; \
#	docker tag $(REGISTRY)/$(APP_NAME):$(BUILD_TAG) $(REGISTRY)/$(APP_NAME):$(IMAGE_TAG) ; \
#	docker push $(REGISTRY)/$(APP_NAME):$(IMAGE_TAG);
#
#create_secret:
#	ansible-vault decrypt secret/solarwinds-appoptics_controllercontrollerrobot-secret.yml.enc \
#		--vault-password-file=secret/vault-secret \
#		--output=- | kubectl apply -f - --namespace $(NAMESPACE)