---
title: "Configuration"
description: "Every gen-asyncapi config key (language, package, output, generate, output-options) with its default and the rules that fail validation."
nav_order: 3
---

# Configuration

The config is a small YAML file passed with `--config`. The [command-line flags](install.md#command-line) override it, and it is optional when the flags say enough. The loader is [`internal/codegen/config.go`](https://github.com/psyb0t/gen-asyncapi/blob/main/internal/codegen/config.go).

```yaml
language: go              # go or typescript, defaults to go
package: chat             # Go package name, required for Go
output: chat.gen.go       # defaults to asyncapi.gen.go, or asyncapi.gen.ts
generate:
  models: true            # types, codecs, validation (on by default)
  client: true            # decoding, Handler, Dispatch, Client (on by default, needs models)
output-options:
  message-matchers:       # see Message matching
    Notice:
      /messages/0/type: [notice, info]
```

## Keys

| Key | Default | What it does |
| --- | --- | --- |
| `language` | `go` | `go` or `typescript`. Any other value fails. Flag: `-language`. |
| `package` | none | Go package name of the generated file. Required for Go, ignored for TypeScript. Flag: `-package`. |
| `output` | `asyncapi.gen.go`, or `asyncapi.gen.ts` for TypeScript | The file to write, relative to the directory the generator runs in. Flag: `-o`. |
| `generate.models` | `true` | Schema types, JSON codecs and validation. |
| `generate.client` | `true` | Message decoding, the `Handler` interface, `Dispatch` and the `Client`. Needs `models`. |
| `output-options.message-matchers` | none | Extra rules that tell incoming messages apart. See [Message matching](message-matching.md). |

## The generate block

Leave `generate` out and you get everything. Once the block is present, a key left out of it is `false`, so `generate: {models: true}` means models only. Turning both off, or turning on `client` without `models`, fails validation.

## message-matchers

The keys under `message-matchers` are message keys from the spec: the name under `components/messages`, or `<channel>_<message>` for a message defined inline in a channel's `messages`. Each one maps JSON pointers ([RFC 6901](https://www.rfc-editor.org/rfc/rfc6901)) into the payload to the value that pointer must hold, or to a list of allowed values:

```yaml
output-options:
  message-matchers:
    Notice:
      /messages/0/type: [notice, info]
    Alert:
      /messages/0/type: alert
```

A key that names no message in the spec fails generation, and so does an empty list. [Message matching](message-matching.md) explains where these rules sit in the decoding order.

## Validation

The config is checked after the flags are applied, before the spec is read:

- Unknown keys fail with `commerr.ErrParseFailed`. A typo such as `ouput:` is caught, not ignored.
- A Go run without `package` fails with `commerr.ErrRequiredConfigValueNotSet`.
- A `language` other than `go` or `typescript` fails with `commerr.ErrInvalidValue`.
- `generate` with nothing selected, or with `client` but not `models`, fails with `commerr.ErrInvalidValue`.

The exact messages are on the [Troubleshooting](troubleshooting.md) page.
