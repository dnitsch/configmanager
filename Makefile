.PHONY: codegen

# export GO111MODULE=on

codegen_verify:
	./hack/verify-codegen.sh

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

tidy: install 
	go mod tidy

install:
	go mod vendor

buildprep: tidy
	rm -rf bin && mkdir -p bin

build: buildprep
	GOOS=linux go build -o bin/genvars-linux ./cmd/genvars

buildmac: buildprep
	GOOS=darwin go build -o bin/genvars-darwin ./cmd/genvars

buildwin: buildprep
	GOOS=windows go build -o bin/genvars.exe ./cmd/genvars
