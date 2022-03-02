.PHONY: test test_ci tidy install buildprep build buildmac buildwin

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
	mkdir -p bin

build: buildprep
	GOOS=linux go build -o bin/configmanager-Linux ./cmd/configmanager

buildmac: buildprep
	GOOS=darwin go build -o bin/configmanager-Darwin ./cmd/configmanager

buildwin: buildprep
	GOOS=windows go build -o bin/configmanager.exe ./cmd/configmanager
