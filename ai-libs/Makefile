GO ?= go
PROTOC ?= protoc
TOOLS_DIR := $(CURDIR)/.bin
PROTOS := api/lattice/catalog/v1/catalog.proto api/lattice/control/v1/control.proto

.PHONY: tools proto check-proto test check

# protoc 36.2 is installed separately. Consumers only need Go.
tools:
	GOBIN="$(TOOLS_DIR)" $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	GOBIN="$(TOOLS_DIR)" $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2

proto:
	@test "$$($(PROTOC) --version)" = "libprotoc 36.2" || { echo 'protoc 36.2 is required'; exit 1; }
	@test "$$($(TOOLS_DIR)/protoc-gen-go --version)" = "protoc-gen-go v1.36.12"
	@test "$$($(TOOLS_DIR)/protoc-gen-go-grpc --version)" = "protoc-gen-go-grpc 1.6.2"
	PATH="$(TOOLS_DIR):$$PATH" $(PROTOC) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $(PROTOS)

check-proto: proto
	git diff --exit-code -- api

test:
	$(GO) test -race ./...

check: check-proto
	$(GO) vet ./...
	$(GO) test -race ./...
