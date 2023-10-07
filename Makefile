run:
	@go run ./cmd/daemon

start:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o daemon cmd/daemon/main.go
	docker-compose up -d
	rm ./daemon

stop:
	docker-compose down

build:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o daemon cmd/daemon/main.go

gen-proto:
	protoc --proto_path=proto proto/*.proto --go_out=internal --go-grpc_out=internal
