PROTOC = protoc
PROTO_DIR = internal/application/adapter/grpc
OUT_DIR = .

protoc:
	$(PROTOC) --go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		--go_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
