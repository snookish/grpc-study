MODULE      := github.com/snookish/grpc
PROTO_DIR   := proto
GEN_DIR     := gen
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')

.PHONY: all proto-gen proto-clean fmt vet test build run-server run-client run-streaming-client clean

all: proto-gen build

proto-gen:
	@echo "Generating protobuf files..."
	@rm -rf $(GEN_DIR)
	@mkdir -p $(GEN_DIR)
	@protoc \
		-I $(PROTO_DIR) \
		--go_out=$(GEN_DIR) \
		--go_opt=module=$(MODULE)/$(GEN_DIR) \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=module=$(MODULE)/$(GEN_DIR) \
		$(PROTO_FILES)
	@echo "Generated protobuf files in $(GEN_DIR)/"

proto-clean:
	@rm -rf $(GEN_DIR)
	@echo "Removed generated protobuf files"

fmt:
	@gofmt -w $$(find . -name '*.go' -not -path './$(GEN_DIR)/*')
	@echo "Formatted Go files"

vet:
	@go vet ./...

test:
	@go test ./...

build:
	@mkdir -p bin
	@go build -o bin/server ./cmd/server
	@go build -o bin/client ./cmd/client
	@go build -o bin/streaming-client ./cmd/streaming-client
	@echo "Built binaries in bin/"

run-server:
	@go run ./cmd/server

run-client:
	@go run ./cmd/client

run-streaming-client:
	@go run ./cmd/streaming-client

clean: proto-clean
	@rm -rf bin
	@echo "Cleaned build artifacts"