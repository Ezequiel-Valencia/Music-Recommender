build:
	@go build -o bin/out cmd/main.go

test:
	@go test ./...

testv:
	@go test -v ./...

run:
	@go run ./cmd/main.go

check:
	@golangci-lint run
	@golangci-lint fmt
