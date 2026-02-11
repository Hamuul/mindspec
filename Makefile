BINARY  := mindspec
BINDIR  := ./bin
PKG     := ./cmd/mindspec

.PHONY: build install test clean

build:
	go build -o $(BINDIR)/$(BINARY) $(PKG)

install:
	go install $(PKG)

test:
	go test ./...

clean:
	rm -rf $(BINDIR)
