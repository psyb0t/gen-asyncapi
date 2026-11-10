# gen-asyncapi

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/gen-asyncapi.svg)](https://pkg.go.dev/github.com/psyb0t/gen-asyncapi)
[![CI](https://github.com/psyb0t/gen-asyncapi/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/gen-asyncapi/actions/workflows/pipeline.yml)
[![coverage](https://raw.githubusercontent.com/psyb0t/gen-asyncapi/badges/coverage.svg)](https://github.com/psyb0t/gen-asyncapi/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/gen-asyncapi/badges/version.svg)](https://github.com/psyb0t/gen-asyncapi/tags)
[![license](https://raw.githubusercontent.com/psyb0t/gen-asyncapi/badges/license.svg)](https://github.com/psyb0t/gen-asyncapi/blob/main/LICENSE)

Feed it an AsyncAPI 3 document and it spits out Go or TypeScript types, strict JSON decoding and a client for your WebSocket API. Run it from `go generate` with a small config file and the path to your spec, and the generated file lands next to your code.

Decoding is strict. A frame either matches its schema or you get an error that names every message it tried and why each one said no.

Spec loading comes from [go-asyncapi](https://github.com/psyb0t/go-asyncapi). Grab that one directly if you want to read AsyncAPI 3 documents in Go without generating anything.

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [Documentation](#documentation)
- [Agent integrations](#agent-integrations)
- [Development](#development)
- [Sponsor](#sponsor)
- [License](#license)

## Install

```sh
go install github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest
```

As a Go tool in your module, so `go tool gen-asyncapi` runs the version your `go.mod` pins:

```sh
go get -tool github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest
```

A TypeScript project needs no Go module, just Go installed: [Install and run](https://psyb0t.github.io/gen-asyncapi/install/) has the `go run` line.

## Quick start

Drop a `config.yaml` and a `gen.go` where the code should go:

```yaml
# config.yaml
package: chat
output: chat.gen.go
```

```go
// gen.go
package chat

//go:generate go tool gen-asyncapi --config=config.yaml ../../api/asyncapi.yaml
```

Run `go generate ./...` and `chat.gen.go` appears, with a struct per schema, strict `UnmarshalJSON` and `Validate`, a `Handler` for incoming messages and a `Client` with one method per outgoing one:

```go
conn, err := gorillaws.Dial(
	ctx,
	chat.ServerProductionURL,
	gorillaws.WithBearerToken(token),
)
if err != nil {
	return err
}

client := chat.NewClient(conn)
defer client.Close()

err = client.JoinRoom(ctx, chat.JoinRequest{ID: id, Room: "lobby"})
```

Add `-language=typescript -o=chat.gen.ts` and the same spec turns into one dependency-free `.ts` file whose client takes a browser `WebSocket`.

## Documentation

The full documentation lives at **https://psyb0t.github.io/gen-asyncapi/**.

| Page | What it covers |
| --- | --- |
| [Install and run](https://psyb0t.github.io/gen-asyncapi/install/) | Install routes, the command line, `go generate`, logging, upgrades |
| [Configuration](https://psyb0t.github.io/gen-asyncapi/configuration/) | Every config key, its default and what fails validation |
| [Go output](https://psyb0t.github.io/gen-asyncapi/go/) | Structs, enums, unions, sending and receiving in Go |
| [TypeScript output](https://psyb0t.github.io/gen-asyncapi/typescript/) | Interfaces, validators and the browser client |
| [Message matching](https://psyb0t.github.io/gen-asyncapi/message-matching/) | How incoming frames are told apart, and `message-matchers` |
| [What gets generated](https://psyb0t.github.io/gen-asyncapi/generated-code/) | Each spec part and what it becomes in both languages |
| [AsyncAPI 3 coverage](https://psyb0t.github.io/gen-asyncapi/spec-coverage/) | Every AsyncAPI object and schema keyword, one row each |
| [Example packages](https://psyb0t.github.io/gen-asyncapi/examples/) | The committed examples and what each one proves |
| [Agent integrations](https://psyb0t.github.io/gen-asyncapi/agents/) | The agent skill for Claude Code, Codex and OpenClaw |
| [Troubleshooting](https://psyb0t.github.io/gen-asyncapi/troubleshooting/) | Generator errors and runtime decoding errors, with fixes |
| [Development](https://psyb0t.github.io/gen-asyncapi/development/) | Make targets, the dev image, CI and releases |

## Agent integrations

The [skill](https://github.com/psyb0t/gen-asyncapi/tree/main/.agents/skills/gen-asyncapi) teaches a coding agent how to set up the generator and use the code it writes, in Go and TypeScript:

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install gen-asyncapi@psyb0t
```

Codex and OpenClaw installs are on the [Agent integrations](https://psyb0t.github.io/gen-asyncapi/agents/) page.

## Development

Everything runs in the `Dockerfile.dev` image. You need Docker and Make.

```sh
make help           # list targets
make generate       # regenerate the examples
make lint           # shellcheck, shfmt, go fix, golangci-lint
make test-coverage  # go test -race with a 90% coverage gate, then make test-ts
make docs           # build the documentation site into .testing/docs-site
```

[Development](https://psyb0t.github.io/gen-asyncapi/development/) covers the rest, including CI and releases. Contributor and agent instructions are in [AGENTS.md](https://github.com/psyb0t/gen-asyncapi/blob/main/AGENTS.md).

## Sponsor

<a href="https://www.buymeacoffee.com/psyb0t"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me a Coffee" height="60" width="217"></a>

## License

[MIT](https://github.com/psyb0t/gen-asyncapi/blob/main/LICENSE). The specs in [examples/official](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/official) come from the AsyncAPI specification repository under Apache 2.0.
