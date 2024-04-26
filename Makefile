.PHONY: clean all

exe := build/main
src := $(shell find . -type f -name '*.go')

all: $(exe)

test:
	go test ./...

$(exe): $(src)
	mkdir -p build
	go build -compiler=gc -o $@ ./cmd/painter

clean:
	rm -rf build

