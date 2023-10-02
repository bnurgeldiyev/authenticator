run:
	@go run ./cmd/daemon

gen-proto:
	protoc --proto_path=proto proto/*.proto --go_out=auth --go-grpc_out=auth
