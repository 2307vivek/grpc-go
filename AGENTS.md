# AGENTS.md

gRPC-Go (`google.golang.org/grpc`). Go implementation of gRPC. Before submitting code, review the enforced conventions in `scripts/vet.sh` and `CONTRIBUTING.md`.

## Commands

- Full suite: `go test -cpu 1,4 -timeout 7m ./...` (also with `-race`). See `Makefile` (`make all` = vet + test + testrace).
- Run `./scripts/vet.sh -install` once, then `./scripts/vet.sh` for the full static check suite. It installs pinned tools from the `test/tools` module (goimports, staticcheck, revive, misspell).
- To vet/test one package: `go test ./internal/transport/` etc.

## Critical gotchas

- **`scripts/vet.sh` requires a clean git working tree.** It fails if `git status --porcelain` is non-empty, and its `trap cleanup EXIT` runs `git reset --hard HEAD`. Run it only after committing/stashing your work; uncommitted changes will be destroyed.
- **Multi-module repo.** The root `go test ./...` does NOT cover nested modules: `security/advancedtls`, `security/advancedtls/examples`, `cmd/protoc-gen-go-grpc`, `gcp/observability`, `stats/opencensus`, `test/tools`, `interop/xds`, `interop/observability`, `examples`. Run tests from each module dir when a change touches them. `make testsubmodule` covers the two `security/` modules.
- CI runs on Go 1.25 and 1.26; `go.mod` requires `go 1.25.0`. Only the two latest Go majors are supported.

## Enforced repo conventions (vet.sh fails on these)

- Every non-generated `.go` file needs the Apache-2.0 header containing `Copyright <4+ digits> gRPC authors` (copy from an existing file; update year). Files marked `DO NOT EDIT` are exempt.
- No `interface{}` — use `any`. No direct `grpclog.I/W/E/F/V` calls — use `grpclog.Component(name).Logger(...)`.
- No `time.After(` outside `*_test.go`; no `net.ParseIP` — use `netip.ParseAddr`; no `regexp.Compile(` in xDS production code — use `CompileSafeRegex`.
- No trailing whitespace; no tabs in `.md`; terminating newline on every file.
- Do not import `github.com/golang/protobuf`; never add dependencies to the core `grpc` package without maintainer discussion.
- Always alias `grpc_testing` (and generated proto/service pairs) imports: `*pb` suffix for message code, `*grpc` suffix for service code.

## Test conventions

- Most test suites are written as `type s struct { grpctest.Tester }` with sub-tests, run via `grpctest.RunSubTests(t, s{})` (`internal/grpctest`). Teardown performs goroutine/timer leak checks, so new tests should follow this pattern.
- End-to-end tests live in `test/` (not `internal/`); `interop/` holds cross-language interop tests (own module). Heavy/network-dependent suites: `test/` e2e, `xds/` (needs xDS infra, see `xds/test/e2e`).

## Generated code

- `*.pb.go` and `*_grpc.pb.go` files are generated. Do not hand-edit; regenerate with `scripts/regenerate.sh` (needs `protoc`, Go tools, and network to pull `grpc-proto`). `grpc_testing_not_regenerated/*.pb.go` are intentionally never regenerated.
- `scripts/vet-proto.sh` checks generated code freshness against upstream protos. It can legitimately fail when upstream protos changed without a corresponding repo update — this CI check is not a merge blocker.

## Conventions for changes

- Style: Google Go Style Guide (https://google.github.io/styleguide/go/). See `.gemini/styleguide.md` for the import-aliasing rule.
- PR titles start with the component (e.g. `transport:`, `xds:`, `client:`); PR descriptions must end with a `RELEASE NOTES:` section, get a `Type:` label, and a `Release` milestone.
- Versioning: semver with exceptions for experimental/deprecated APIs; `[v1.74.0, v1.74.1]` are retracted in `go.mod` — do not reference them.
