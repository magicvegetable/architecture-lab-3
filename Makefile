.PHONY: clean all

all: build/main

src := cmd/painter/main.go painter/op.go painter/loop.go painter/lang/http.go painter/lang/parser.go ui/window.go

test: *.go
	go test ./...

build/main: $(src)
	mkdir -p build
	go build -compiler=gc -o build/main ./cmd/painter

clean:
	rm -rf build