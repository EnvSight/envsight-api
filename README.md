# EnvSight API

Shared gRPC/Protobuf contract between the [EnvSight Agent](https://github.com/Moriaty911/envsight-agent) and the [EnvSight Server](https://github.com/Moriaty911/envsight-server).

This repository is the **single source of truth** for the `AgentService` gRPC contract. Both the agent and the server import the generated Go code from this module instead of maintaining their own copies, so the wire format can never silently drift between the two sides.

## Why a separate module?

Previously each side kept its own copy of `envsight.proto` and the generated `*.pb.go`. That led to:

- The server's `go_package` pointing at the agent's module path, papered over with a hand-edited `package pb`.
- The two proto copies and their generated code drifting out of sync (stripped comments, stale annotations).

A dedicated, MIT-licensed module fixes all three at once: one proto, one generator, one import path, and a CI target that rejects drift.

## Structure

```
envsight-api/
├── proto/
│   └── envsight.proto   # The contract source — edit this
├── pb/                  # Generated Go code (committed)
│   ├── envsight.pb.go
│   └── envsight_grpc.pb.go
├── Makefile             # `make generate` / `make verify`
└── go.mod               # module github.com/Moriaty911/envsight-api
```

The generated `pb/` files are committed (not gitignored) so consumers can `go get` without running protoc themselves. This matches the convention used by most Go protobuf modules.

## Prerequisites

- **Go** 1.21+ (the module's `go` directive may track a newer toolchain; consumers on 1.21 are fine)
- **protoc** + `protoc-gen-go` + `protoc-gen-go-grpc` — only if you modify the `.proto` file

## Regenerating after a proto change

```bash
make generate   # regenerates pb/ from proto/
```

Then commit both `proto/envsight.proto` and the updated `pb/*.pb.go`.

## Preventing drift in CI

Run this in CI to fail the build when the committed generated code is stale:

```bash
make verify
```

It regenerates into `pb/` and errors out if `git diff` finds any change — i.e. someone edited the proto without regenerating, or hand-edited the generated files.

## Consuming from the agent or server

Add the dependency:

```bash
go get github.com/Moriaty911/envsight-api
```

Then import the generated package:

```go
import pb "github.com/Moriaty911/envsight-api/pb"
```

## Versioning

This module follows [Go module versioning](https://go.dev/doc/modules/version-numbers). Tag releases with semver (`v1.0.0`, `v1.1.0`, ...). A breaking change to the proto contract requires a major version bump (`v2`), which Go automatically reflects in the import path as `github.com/Moriaty911/envsight-api/v2/pb`.

## License

MIT — see [LICENSE](./LICENSE). This permissive license lets both the MIT-licensed agent and the proprietary-licensed server depend on it cleanly.
