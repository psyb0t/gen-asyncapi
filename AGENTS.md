# gen-asyncapi

Instructions for coding agents (Codex, Claude Code and others) and for people changing this repository. To use the generator, read the documentation site, https://psyb0t.github.io/gen-asyncapi/, or the agent skill in [.agents/skills/gen-asyncapi/SKILL.md](.agents/skills/gen-asyncapi/SKILL.md). This file is about changing the code.

## Agent setup

`CLAUDE.md` is a committed symlink to `AGENTS.md`, and `.claude` is a committed symlink to `.agents`, so Claude Code reads the same instructions, rules and skills as Codex. Claude's Edit and Write tools refuse to write through a symlink, so edit `AGENTS.md` and `.agents/rules/*.md` directly.

## What it is

A Go module with a command, `gen-asyncapi`, that reads an AsyncAPI 3 document and writes one Go or TypeScript file: types, strict JSON decoding with validation, message dispatch and a WebSocket client. Users run it from `go generate` with a small YAML config. The module also ships `asyncapirt`, the runtime the generated Go imports, and `asyncapirt/gorillaws`, a transport over gorilla/websocket. There is no Docker image and no binary release; users install with `go get -tool`, `go install` or `go run`.

## Architecture

- `cmd/gen-asyncapi`: the command. Parses `--config`, `-language`, `-package`, `-o` and the one spec path, applies the flags on top of the config, validates it, calls `codegen.Generate` and writes the output file with mode `0644`. It is the only code that logs.
- `internal/codegen`: the generator. `config.go` loads the YAML config (unknown keys fail) and validates it. `builder.go` and `schema.go` turn schemas into an in-memory model (`model.go`: declarations, type references, checks); `names.go` holds the `nameSet` that hands out collision-free identifiers. `generate.go` builds messages, servers, channels and operations, then picks the renderer. `render.go` plus `templates/file.tmpl` write Go, formatted by `imports.Process`; `typescript.go`, `tscheck.go`, `templates/typescript.tmpl` and `templates/runtime.ts` write TypeScript. `doc.go` renders schema doc comments.
- `asyncapirt`: runtime the generated Go imports. Validation checks (`validate.go`), strict object decoding and incoming message matching (`match.go`), `Send` and the `Transport` interface (`transport.go`), and the exported sentinels (`errors.go`).
- `asyncapirt/gorillaws`: a `Transport` over gorilla/websocket, with `Dial` and its `With*` options.
- Spec loading comes from `github.com/psyb0t/go-asyncapi`, a separate module (vendored).

Data flow: spec path -> `asyncapi.Load` -> `builder` model -> Go template or TypeScript renderer -> one output file.

## Stack

- Go: `go 1.26.9` minimum in `go.mod`; the dev image runs Go 1.26.10. Errors: `ctxerrors` wrapping `commerr` sentinels. Logging (command only): `slog` configured from `LOG_*` by `slogging/slogconf`, logger from `ctxscope.GetLogger(ctx)`.
- Lint: `golangci-lint` as a Go tool (`go tool golangci-lint`), configured in `.golangci.yml`; shellcheck and shfmt for `scripts/`.
- TypeScript checks: `typescript` and `@types/node` pinned in `package.json`, locked in `pnpm-lock.yaml`, installed into the dev image; tests use `node:test`.
- Security scan: gitleaks, govulncheck, `pnpm audit` and semgrep (hash-pinned in `requirements-semgrep.txt`), merged into `sec.sarif`.
- CI: `.github/workflows/pipeline.yml` calling `psyb0t/reusable-github-workflows`.
- Docs: Jekyll with the Just the Docs theme, built by GitHub Pages from `docs/` on `main`.

## Layout

| Path | What it holds |
| --- | --- |
| `cmd/gen-asyncapi/` | The command and its tests |
| `internal/codegen/` | Config, model builder, Go and TypeScript renderers, `templates/` |
| `asyncapirt/`, `asyncapirt/gorillaws/` | The runtime the generated Go imports, and the WebSocket transport |
| `examples/chat/`, `examples/shapes/`, `examples/kitchensink/` | Specs, configs, `gen.go`, committed Go and TypeScript output, and Go and TypeScript tests against it |
| `examples/official/` | The 19 upstream AsyncAPI examples, generated in both languages. The specs are byte-for-byte upstream (Apache 2.0, see its `README.md`, `LICENSE`, `NOTICE`) |
| `docs/` | The documentation site. `spec-coverage.md` lists what happens to each AsyncAPI object and schema keyword |
| `scripts/` | `lint.sh`, `test-coverage.sh`, `test-ts.sh`, `sec.sh`, run by the Makefile inside the dev image |
| `Dockerfile.dev` | The dev image every make target runs in |
| `tsconfig.json`, `tsconfig.test.json` | Strict check of the generated TypeScript, and of the tests with Node.js types |
| `package.json`, `pnpm-lock.yaml`, `pnpm-workspace.yaml` | The TypeScript check tools and the seven-day release-age gate |
| `vendor/` | Committed Go dependencies, refreshed by `make dep` |
| `.agents/skills/gen-asyncapi/` | The agent skill and its `references/setup.md` |
| `.agents/.claude-plugin/`, `.agents/.codex-plugin/` | Claude Code and Codex plugin manifests |
| `.agents/rules/` | Rules for changing this repository |
| `.github/workflows/` | `pipeline.yml` plus repository housekeeping workflows |

## Commands

Docker and Make are enough. Every target builds the dev image and runs inside it.

```bash
make help           # every target
make generate       # regenerate every example (Go and TypeScript)
make lint           # shellcheck, shfmt, go fix -diff, golangci-lint
make test           # go test -race, then make test-ts
make test-coverage  # go test -race with a 90% gate, then make test-ts
make test-ts        # tsc on the generated TypeScript, then node:test suites
make sec            # gitleaks, govulncheck, pnpm audit, semgrep into sec.sarif
make dep            # go mod tidy and go mod vendor
make docs           # build the documentation site into .testing/docs-site
```

`pre-commit.sh` runs `make lint && make test-coverage`. CI runs `make lint`, `make test-coverage`, `make sec` and `make generate` (which fails on drift) on every branch push.

## Project-specific landmines

- Never hand-edit `*.gen.go` or `*.gen.ts`. `TestGeneratedExamplesAreCurrent` regenerates every `go:generate` line and fails on any difference, so a generator change always ships with regenerated examples.
- The generated TypeScript must keep passing the strict `tsconfig.json`, which checks it without Node.js types. Runtime helpers in `templates/runtime.ts` are only emitted when the generated code calls them, through the `//@rt <name>` markers; a helper without a marker never ships.
- Both languages share one model. A builder change shows up in both outputs; check both.
- Alpine drops old package versions, so the exact `apk` pins in `Dockerfile.dev` stop resolving over time and `make dev-image` fails with "unable to select packages". Update the pins to what the pinned base image's Alpine release offers.
- A Go patch release with a standard library fix makes `make sec` fail through govulncheck until the `Dockerfile.dev` base image moves to that release.

## Rules

Rules for every change live in `.agents/rules/`, one file per topic. Read the ones that match what you are touching before editing.
