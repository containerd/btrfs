
.PHONY: clean binaries

all: vet lint test binaries

binaries: bin/btrfs-test

vet:
	go vet ./...

lint:
	golint ./...

test:
	go test -v ./...

bin/%: ./cmd/%
	go build -o ./$@ ./$<

clean:
	rm -rf bin/*
