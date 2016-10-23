Name := recorder
Version := $(shell git describe --tags --abbrev=0)
GOOS := linux
GOARCH := amd64
OWNER := heptagon-inc
.DEFAULT_GOAL := help

## Setup
setup:
	go get github.com/kardianos/govendor
	go get github.com/Songmu/make2help/cmd/make2help

## Install dependencies
deps: setup
	govendor sync

## Initialize and Update dependencies
update: setup
	rm -rf /vendor/vendor.json
	govendor fetch +outside

## Vet
vet: setup
	govendor vet +local

## Lint
lint: setup
	go get github.com/golang/lint/golint
	govendor vet +local
	for pkg in $$(govendor list -p -no-status +local); do \
		golint -set_exit_status $$pkg || exit $$?; \
	done

## Run tests
test: deps
	govendor test +local

## Build
build: deps
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-X main.Version=$(Version) -X main.Name=$(Name)"

## Build
build-local: deps
	go build -ldflags "-X main.Version=$(Version) -X main.Name=$(Name)"

## Release
release: build
	zip $(Name)_$(GOOS)_$(GOARCH).zip $(Name)
	ghr -t ${GITHUB_TOKEN} -u $(OWNER) -r $(Name) --replace $(Version) $(Name)_$(GOOS)_$(GOARCH).zip

## Show help
help:
	@make2help $(MAKEFILE_LIST)

.PHONY: setup deps update vet lint test build build-local release help
