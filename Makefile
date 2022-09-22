.PHONY: lint build test integrationtest clean install

MAINPKG=./cmd/youtubetoolkit
TARGET=./bin/youtubetoolkit

all: clean build test

acceptance: clean build integrationtest

lint:
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.49.0 golangci-lint run -v
build:
	go build -o $(TARGET) $(MAINPKG)
test:
	go test ./...
integrationtest:
	go test -v -tags integration -count=1 ./...
clean:
	$(RM) $(TARGET)
install:
	go install $(MAINPKG)
