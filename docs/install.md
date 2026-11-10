---
title: "Install and run"
description: "Install gen-asyncapi as a Go tool, a binary or a go run line, then run it from go generate or the command line, with its flags and logging settings."
nav_order: 2
---

# Install and run

## Requirements

- Go 1.26.9 or newer to run the generator. That is the `go` line in the module's [`go.mod`](https://github.com/psyb0t/gen-asyncapi/blob/main/go.mod).
- Generated Go code imports [`asyncapirt`](https://github.com/psyb0t/gen-asyncapi/tree/main/asyncapirt) from this module, `asyncapirt/gorillaws` when you use that transport, and `github.com/psyb0t/ctxerrors` and `github.com/psyb0t/ctxscope`. `go mod tidy` picks them up after the first generation.
- Generated TypeScript has no imports and sticks to ES2020, so it needs nothing beyond a browser or Node.js.

## Install

As a Go tool in your module, so `go tool gen-asyncapi` runs the version your `go.mod` pins:

```sh
go get -tool github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest
```

As a binary on your `PATH`:

```sh
go install github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest
```

A TypeScript project needs no Go module, just Go installed. Run the generator straight from the module proxy:

```sh
go run github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest \
  -language=typescript \
  -o=src/chat.gen.ts \
  asyncapi.yaml
```

Pin a version instead of `@latest` and put that line in a `package.json` script, so every run writes the same code.

## Run it from go generate

Put a `config.yaml` and a `gen.go` in the directory where the generated code should live:

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

Run `go generate ./...` and `chat.gen.go` appears. `go generate` runs each line in the directory of the file that holds it, so the spec path and the paths in the config are relative to that directory.

A Go server and its web client can generate from one spec in the same `gen.go`:

```go
//go:generate go tool gen-asyncapi --config=config.yaml chat.yaml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=chat.gen.ts chat.yaml
```

Commit the generated files next to the spec, and regenerate whenever the spec or the config changes. Never edit a `*.gen.go` or `*.gen.ts` file by hand: the next run overwrites it.

## Command line

```
gen-asyncapi [--config=config.yaml] [-language=go|typescript] [-package=name] [-o=output.go] asyncapi.yaml
```

The spec path is the one positional argument. Anything else is an error that prints the usage line.

| Flag | What it does |
| --- | --- |
| `--config` | Path to the [config file](configuration.md). Optional. |
| `-language` | `go` or `typescript`. Overrides `language` in the config. |
| `-package` | Go package name. Overrides `package` in the config. |
| `-o` | Output file. Overrides `output` in the config. |

The flags are applied on top of the config before it is validated, so they can stand in for the config file entirely:

```sh
gen-asyncapi -package=chat -o=chat.gen.go asyncapi.yaml
```

The command writes one file with mode `0644`, replacing any file already at that path, and exits with status 1 on any failure.

## Logging

The command logs through Go's `slog`, configured from the environment by [slogging](https://github.com/psyb0t/slogging). A successful run logs one `generated` line with the spec, the output path and `size_bytes`. A failed run logs `generation failed` with the error.

| Variable | Values | Default |
| --- | --- | --- |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` |
| `LOG_FORMAT` | `text`, `json` | `text` |
| `LOG_ADD_SOURCE` | `true`, `false` | `false` |

`LOG_LEVEL=warn` keeps `go generate` quiet unless something fails.

## Upgrade

Bump the version in `go.mod` (or the `@version` in your `go run` line), then run `go generate ./...` again and commit the new output. Some fixes only reach your code through regeneration, because the generated file holds the validation and matching code. The [CHANGELOG](https://github.com/psyb0t/gen-asyncapi/blob/main/CHANGELOG.md) says when a release needs it, and lists the migration for every breaking change, such as the v0.2.0 rename from `gogen-asyncapi`.

To remove the generator, run `go get -tool github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@none`, or delete the binary `go install` put in `$(go env GOPATH)/bin`. The generated files keep working on their own, since they only import the runtime packages.
