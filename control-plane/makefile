PROTOC = protoc
PROTO_DIR = internal/application/adapter/grpc
OUT_DIR = .

TEST_CONFIG_PATH = ../config/config.yml
DEV_CONFIG_PATH = ./config/config.yml

.PHONY: proto
protoc:
	$(PROTOC) --go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		--go_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto



.PHONY: test
test-uc:
	CONFIG_PATH=$(TEST_CONFIG_PATH) go test ./test/usecase/

.PHONY: run-dev

run-dev:
	CONFIG_PATH=$(DEV_CONFIG_PATH) go run cmd/main.go