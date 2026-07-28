export PATH := $(PATH):$(shell go env GOPATH)/bin

# Generator versions are pinned, not floating on @latest, because both plugins
# stamp their own version into the header of every generated file. With
# @latest, `make verify` fails the moment upstream cuts a release — reporting
# "generated code is out of date" for a one-line version banner while the
# actual wire contract is unchanged. That trains everyone to ignore the drift
# check, which is the one thing this module exists to enforce.
#
# PROTOC_GEN_GO_VERSION must match google.golang.org/protobuf in go.mod.
# PROTOC_VERSION must match the `protoc` line in the pb/*.pb.go headers.
# Bump these deliberately, run `make generate`, and commit the result.
PROTOC_GEN_GO_VERSION      := v1.36.11
PROTOC_GEN_GO_GRPC_VERSION := v1.6.2
PROTOC_VERSION             := 35.1

.PHONY: all generate clean verify install-tools protoc-version print-protoc-version

all: generate

# Single source of truth for CI, which must install this exact protoc.
print-protoc-version:
	@echo $(PROTOC_VERSION)

install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

# protoc itself is not go-installable, so it cannot be pinned the same way.
# Fail loudly on a mismatch instead of silently regenerating files that differ
# only by a version banner.
protoc-version:
	@have="$$(protoc --version | awk '{print $$2}')"; \
	if [ "$$have" != "$(PROTOC_VERSION)" ]; then \
		echo "ERROR: protoc $(PROTOC_VERSION) required, found $$have."; \
		echo "Install the matching release from"; \
		echo "  https://github.com/protocolbuffers/protobuf/releases/tag/v$(PROTOC_VERSION)"; \
		echo "or bump PROTOC_VERSION in this Makefile if the upgrade is intentional."; \
		exit 1; \
	fi

generate: protoc-version install-tools
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
