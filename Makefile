.PHONY: proto proto-tools

GOOGLEAPIS_DIR ?= .cache/googleapis
GOOGLEAPIS_BASE ?= https://raw.githubusercontent.com/googleapis/googleapis/master
GOOGLEAPIS_PROTO_DIR := $(GOOGLEAPIS_DIR)/google/api
GOOGLEAPIS_PROTOS := $(GOOGLEAPIS_PROTO_DIR)/annotations.proto $(GOOGLEAPIS_PROTO_DIR)/http.proto

$(GOOGLEAPIS_PROTOS):
	mkdir -p $(GOOGLEAPIS_PROTO_DIR)
	curl -L -o $(GOOGLEAPIS_PROTO_DIR)/annotations.proto $(GOOGLEAPIS_BASE)/google/api/annotations.proto
	curl -L -o $(GOOGLEAPIS_PROTO_DIR)/http.proto $(GOOGLEAPIS_BASE)/google/api/http.proto

proto: $(GOOGLEAPIS_PROTOS)
	protoc -I proto -I $(GOOGLEAPIS_DIR) \
		--go_out=. --go_opt=module=aeibi \
		--go-grpc_out=. --go-grpc_opt=module=aeibi \
		--grpc-gateway_out=. --grpc-gateway_opt=module=aeibi \
		proto/*.proto

proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
