.PHONY: run build

run:
	go run cmd/server/main.go

build:
	CGO_ENABLED=0 GOOS=linux go build -o dist/server cmd/server/main.go
