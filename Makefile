MAINPKG=./cmd/youtubetoolkit
TARGET=./bin/youtubetoolkit

all: clean build test

acceptance: clean build integrationtests

build:
	go build -o $(TARGET) $(MAINPKG)
test:
	go test ./...
integrationtests:
	INTEGRATION_TESTS=true go test -v $(MAINPKG)/cli_test.go
clean:
	$(RM) $(TARGET)
install:
	go install $(MAINPKG)
