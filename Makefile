run:
	@go run ./cmd/daemon

gen-proto:
	protoc --proto_path=proto proto/*.proto --go_out=internal --go-grpc_out=internal
