MAKEFLAGS := --print-directory
SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

BINARY=starlet

# for CircleCI, GitHub Actions, GitLab CI build number
ifeq ($(origin CIRCLE_BUILD_NUM), environment)
	BUILD_NUM ?= cc$(CIRCLE_BUILD_NUM)
else ifeq ($(origin GITHUB_RUN_NUMBER), environment)
	BUILD_NUM ?= gh$(GITHUB_RUN_NUMBER)
else ifeq ($(origin CI_PIPELINE_IID), environment)
	BUILD_NUM ?= gl$(CI_PIPELINE_IID)
endif

# for go dev
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GODOC=$(GOCMD) doc
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# for go build
# export CGO_ENABLED=0
export TZ=Asia/Shanghai
export PACK=main
export FLAGS="-s -w -X '$(PACK).AppName=$(BINARY)' -X '$(PACK).BuildDate=`date '+%Y-%m-%dT%T%z'`' -X '$(PACK).BuildHost=`hostname`' -X '$(PACK).GoVersion=`go version`' -X '$(PACK).GitBranch=`git symbolic-ref -q --short HEAD`' -X '$(PACK).GitCommit=`git rev-parse --short HEAD`' -X '$(PACK).GitSummary=`git describe --tags --dirty --always`' -X '$(PACK).CIBuildNum=${BUILD_NUM}'"

# commands
.PHONY: default ci test test_loop bench
default:
	@echo "build target is required for $(BINARY)"
	@exit 1

ci:
	$(GOTEST) -v -race -cover -covermode=atomic -coverprofile=coverage.txt -count 1 ./...

test:
	$(GOTEST) -v -race -cover -covermode=atomic -count 1 ./...

test_loop:
	while true; do \
		$(GOTEST) -v -race -cover -covermode=atomic -count 1 ./...; \
		if [[ $$? -ne 0 ]]; then \
			break; \
		fi; \
	done

bench:
	$(GOTEST) -parallel=4 -run="none" -benchtime="2s" -benchmem -bench=./...
