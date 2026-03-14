.PHONY: install test

install:
	go install ./cmd/dev

test:
	go test ./...
