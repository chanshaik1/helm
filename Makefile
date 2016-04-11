# Copyright 2015 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO_DIRS ?= $(shell glide nv -x )
GO_PKGS ?= $(shell glide nv)

BIN_DIR := bin
PATH_WITH_BIN = PATH="$(shell pwd)/$(BIN_DIR):$(PATH)"
ROOTFS := rootfs
CLIENT := cmd/helm

.PHONY: info
info:
	$(MAKE) -C $(ROOTFS) $@

.PHONY: gocheck
ifndef GOPATH
	$(error No GOPATH set)
endif

.PHONY: build
build: gocheck
	@scripts/build-go.sh

.PHONY: build-static
build-static: gocheck
	@BUILD_TYPE=STATIC scripts/build-go.sh

.PHONY: build-cross
build-cross: gocheck
	@BUILD_TYPE=CROSS scripts/build-go.sh

.PHONY: all
all: build

.PHONY: clean
clean:
	$(MAKE) -C $(ROOTFS) $@
	go clean -v $(GO_PKGS)
	rm -rf bin

.PHONY: test
test: build test-style test-unit test-flake8

.PHONY: quicktest
quicktest: test-style
	go test $(GO_PKGS)

.PHONY: push
push: push-server push-client

.PHONY: push-server
push-server: build-static
	$(MAKE) -C $(ROOTFS) push

.PHONY: push-client
push-client: gocheck
	@BUILD_TYPE=CROSS scripts/build-go.sh $(CLIENT)
	$(MAKE) -C $(CLIENT) push

.PHONY: container
container: build-static
	$(MAKE) -C $(ROOTFS) $@

.PHONY: test-unit
test-unit:
	@echo Running tests...
	go test -race -v $(GO_PKGS)

.PHONY: test-flake8
test-flake8:
	@echo Running flake8...
	flake8 expansion
	@echo ----------------

.PHONY: test-style
test-style:
	@scripts/validate-go.sh

.PHONY: test-e2e
test-e2e: container local-cluster-up
	$(PATH_WITH_BIN) go test -tags=e2e ./test/e2e -v --manager-image=${DOCKER_REGISTRY}/manager:${TAG} --resourcifier-image=${DOCKER_REGISTRY}/resourcifier:${TAG} --expandybird-image=${DOCKER_REGISTRY}/expandybird:${TAG}

.PHONY: local-cluster-up
local-cluster-up:
	@scripts/kube-up.sh

.PHONY: local-cluster-down
local-cluster-down:
	@scripts/kube-down.sh

HAS_GLIDE := $(shell command -v glide;)
HAS_GOLINT := $(shell command -v golint;)
HAS_GOVET := $(shell command -v go tool vet;)
HAS_GOX := $(shell command -v gox;)
HAS_PIP := $(shell command -v pip;)
HAS_FLAKE8 := $(shell command -v flake8;)

.PHONY: bootstrap
bootstrap:
	@echo Installing deps
ifndef HAS_PIP
	$(error Please install the latest version of Python pip)
endif
ifndef HAS_GLIDE
	go get -u github.com/Masterminds/glide
endif
ifndef HAS_GOLINT
	go get -u github.com/golang/lint/golint
endif
ifndef HAS_GOVET
	go get -u golang.org/x/tools/cmd/vet
endif
ifndef HAS_GOX
	go get -u github.com/mitchellh/gox
endif
ifndef HAS_FLAKE8
	pip install flake8
endif
	glide install
	pip install --user -r expansion/requirements.txt
