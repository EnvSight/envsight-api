export PATH := $(PATH):$(shell go env GOPATH)/bin

.PHONY: all generate clean verify install-tools

all: generate

install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate: install-tools
	mkdir -p pb
	protoc --proto_path=proto \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		proto/envsight.proto

# verify checks that the committed generated code is up to date with the proto
# source. Run this in CI to catch drift before it lands.
verify: generate
	@git diff --exit-code -- pb/ || ( \
		echo ""; \
		echo "ERROR: generated pb/ files are out of date with proto/."; \
		echo "Run 'make generate' and commit the result."; \
		exit 1; \
	)

clean:
	rm -rf pb/*
