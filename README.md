# gogen-asyncapi

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/gogen-asyncapi.svg)](https://pkg.go.dev/github.com/psyb0t/gogen-asyncapi)
[![CI](https://github.com/psyb0t/gogen-asyncapi/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/gogen-asyncapi/actions/workflows/pipeline.yml)
[![coverage](https://raw.githubusercontent.com/psyb0t/gogen-asyncapi/badges/coverage.svg)](https://github.com/psyb0t/gogen-asyncapi/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/gogen-asyncapi/badges/version.svg)](https://github.com/psyb0t/gogen-asyncapi/tags)
[![license](https://raw.githubusercontent.com/psyb0t/gogen-asyncapi/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/gogen-asyncapi/badges/importers.svg)](https://github.com/psyb0t/gogen-asyncapi/blob/badges/importers.md)

_your spec, my types, zero guessing_

Feed it an AsyncAPI 3 document and it spits out Go types, strict JSON codecs and a client for your WebSocket API. Run it from `go generate` with a small config file and the path to your spec, and the generated file lands next to your code.

Decoding is strict. A frame either matches its schema or you get an error that names every message it tried and why each one said no.

## What it generates

Structs, where required properties are values, optional ones are `omitempty` pointers and nullable ones are pointers that keep `null`. Enums become named types with constants, `allOf` merges into one struct, and every schema in `components.schemas` gets a type.

`UnmarshalJSON` methods that check required properties, unknown properties under `additionalProperties: false`, `const` values and every constraint the schema sets: lengths, patterns, numeric bounds, `multipleOf`, item counts, `uniqueItems`, property counts, `patternProperties`, `propertyNames`, `dependentRequired`, enums. Errors wrap `commerr.ErrValidationFailed`, `commerr.ErrRequiredFieldNotSet` or `commerr.ErrCouldNotDecode`, so `errors.Is` finds them however deep they happen.

Union types for `oneOf` and `anyOf`, with `FromX`, `AsX` and `Value`. The variant comes from the `discriminator`, from a property every variant pins with `const`, or from the one variant that decodes.

A client: a typed method for every send operation, `DecodeReceivedMessage`, a `Handler` interface with an `UnimplementedHandler` to embed, `Dispatch` and `Client.Run`. It talks to an `asyncapirt.Transport`. `asyncapirt/gorillaws` is a ready one on gorilla/websocket, with header, bearer token, query string and subprotocol auth.

Constants and helpers for the rest of the spec: server URLs with their variables, channel addresses with their parameters, message names, header types and correlation ID locations. Traits, `defaultContentType`, multi-format schemas and `$ref`s across files are resolved first.

Every object and keyword, one row each, is in [docs/spec-coverage.md](docs/spec-coverage.md).

## Install

As a tool in your module, so `go tool gogen-asyncapi` runs the version your `go.mod` pins:

```sh
go get -tool github.com/psyb0t/gogen-asyncapi/cmd/gogen-asyncapi@latest
```

Or as a binary:

```sh
go install github.com/psyb0t/gogen-asyncapi/cmd/gogen-asyncapi@latest
```

## Use

Drop a `config.yaml` and a `gen.go` where the code should go:

```yaml
# config.yaml
package: chat
output: chat.gen.go
generate:
  models: true
  client: true
output-options:
  message-matchers:
    Notice:
      /messages/0/type: notice
```

```go
// gen.go
package chat

//go:generate go tool gogen-asyncapi --config=config.yaml ../../api/asyncapi.yaml
```

Run `go generate ./...`. The spec is the last argument, and paths are relative to the directory `go generate` runs in. `-package` and `-o` override the config, or stand in for it:

```sh
gogen-asyncapi -package=chat -o=chat.gen.go asyncapi.yaml
```

| Config key | Meaning |
|---|---|
| `package` | Go package name. Required. |
| `output` | Generated file. Defaults to `asyncapi.gen.go`. |
| `generate.models` | Types, codecs and validation. On by default. |
| `generate.client` | Message decoding, `Handler`, `Dispatch` and `Client`. Needs models. On by default. |
| `output-options.message-matchers` | Per message, JSON pointers into the payload and the value each must hold. |

### Which message is this?

`DecodeReceivedMessage` works through the messages the client receives:

1. Messages whose matchers all hold. Matchers are your `message-matchers` plus the `const` values on the payload's top-level properties.
2. The remaining messages, decoded strictly. Exactly one has to decode.
3. Messages with an opaque payload, like an Avro schema kept as raw JSON.

Nothing matching gives `asyncapirt.ErrUnknownMessage` with every candidate's reason joined in. Two matches give `asyncapirt.ErrAmbiguousMessage`. Messages that only differ by a value the spec lists under `examples` need a matcher.

### Running the client

```go
conn, err := gorillaws.Dial(ctx, chat.ServerProductionURL, gorillaws.WithBearerToken(token))
if err != nil {
	return err
}

client := chat.NewClient(conn)
defer client.Close()

if err := client.JoinRoom(ctx, chat.JoinRequest{ID: id, Room: "lobby"}); err != nil {
	return err
}

return client.Run(ctx, handler) // handler embeds chat.UnimplementedHandler
```

## Examples

- [examples/chat](examples/chat): the whole client, tested end to end over an in-memory transport.
- [examples/shapes](examples/shapes): schema edge cases, plus an OpenAPI 3.0 payload with `nullable` and a discriminator mapping.
- [examples/kitchensink](examples/kitchensink): every AsyncAPI 3.0 object in one document, valid against the official JSON Schema.
- [examples/official](examples/official): the 19 examples from the AsyncAPI spec repository, generated as they are. CI builds them and checks they match the generator.

## Development

Everything runs in the `Dockerfile.dev` image. You need Docker and Make.

```sh
make help           # list targets
make generate       # regenerate the examples
make lint           # shellcheck, shfmt, go fix, golangci-lint
make test-coverage  # go test -race with a 90% coverage gate
make sec            # gitleaks, govulncheck, semgrep
```

## License

[MIT](LICENSE). The specs in [examples/official](examples/official) come from the AsyncAPI specification repository under Apache 2.0.
