.PHONY: build test examples clean

build:
	go build ./...

test:
	go test -v ./...

examples:
	cd examples/basic && go build -o ../../bin/example-basic
	cd examples/interactive && go build -o ../../bin/example-interactive

clean:
	rm -rf bin/

deps:
	go mod tidy
	go mod download

lint:
	go vet ./...
	gofmt -w .

check: lint test
