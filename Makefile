BINARY := tomu
ifeq ($(OS),Windows_NT)
    BINARY := tomu.exe
endif

.PHONY: build test clean

build:
	go build -o $(BINARY) .

test:
	go test ./... -count=1

clean:
	go clean
	rm -f tomu tomu.exe
