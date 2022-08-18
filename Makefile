
OWNER := dnitsch
NAME := configmanager
GIT_TAG := "1.4.2"
VERSION := "v1.4.2"
# VERSION := "$(shell git describe --tags --abbrev=0)"
REVISION := $(shell git rev-parse --short HEAD)

LDFLAGS := -ldflags="-s -w -X \"github.com/$(OWNER)/$(NAME)/cmd/configmanager.Version=$(VERSION)\" -X \"github.com/$(OWNER)/$(NAME)/cmd/configmanager.Revision=$(REVISION)\" -extldflags -static"

.PHONY: test test_ci tidy install cross-build 

test: test_prereq
	go test `go list ./... | grep -v */generated/` -v -mod=readonly -coverprofile=.coverage/out | go-junit-report > .coverage/report-junit.xml && \
	gocov convert .coverage/out | gocov-xml > .coverage/report-cobertura.xml

test_ci:
	go test ./... -mod=readonly

test_prereq: 
	mkdir -p .coverage
	go install github.com/jstemmer/go-junit-report@v0.9.1 && \
	go install github.com/axw/gocov/gocov@v1.0.0 && \
	go install github.com/AlekSi/gocov-xml@v1.0.0

tidy: 
	go mod tidy

install: tidy
	go mod vendor

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist/*
	rm -rf vendor/*

cross-build:
	for os in darwin linux windows; do \
	    [ $$os = "windows" ] && EXT=".exe"; \
		GOOS=$$os CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)-$$os$$EXT ./cmd; \
	done

release:
	OWNER=$(OWNER) NAME=$(NAME) PAT=$(PAT) VERSION=$(VERSION) . hack/release.sh 

tag: 
	git tag "v$(GIT_TAG)"
	git push origin "v$(GIT_TAG)"

tagbuildrelease: tag cross-build release
