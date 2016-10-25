Name := recorder
Version := $(shell git describe --tags --abbrev=0)
OWNER := heptagon-inc
.DEFAULT_GOAL := help

## Setup
setup:
	go get github.com/kardianos/govendor
	go get github.com/Songmu/make2help/cmd/make2help
	go get github.com/mitchellh/gox

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
	gox -osarch="darwin/amd64 linux/amd64" -ldflags="-X main.Version=$(Version) -X main.Name=$(Name)" -output="pkg/{{.OS}}_{{.Arch}}/$(Name)"

## Build
build-local: deps
	go build -ldflags "-X main.Version=$(Version) -X main.Name=$(Name)" -o pkg/$(Name)

## Release
release: build
	mkdir -p pkg/release
	zip pkg/release/$(Name)_darwin_amd64.zip pkg/darwin_amd64/$(Name)
	zip pkg/release/$(Name)_linux_amd64.zip pkg/linux_amd64/$(Name)
	ghr -t ${GITHUB_TOKEN} -u $(OWNER) -r $(Name) --replace $(Version) pkg/release/

## Show help
help:
	@make2help $(MAKEFILE_LIST)

.PHONY: setup deps update vet lint test build build-local release help
