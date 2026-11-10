---
title: "Example packages"
description: "The committed gen-asyncapi examples: chat, shapes, kitchensink and the 19 official AsyncAPI specs, each generated in Go and TypeScript and checked in CI."
nav_order: 8
---

# Example packages

Every example is a directory with a spec, a `config.yaml`, a `gen.go` holding the `go:generate` lines, and the committed Go and TypeScript output. `make generate` rewrites them all, and `TestGeneratedExamplesAreCurrent` in [`internal/codegen/codegen_test.go`](https://github.com/psyb0t/gen-asyncapi/blob/main/internal/codegen/codegen_test.go) fails when the committed output no longer matches what the generator writes.

| Directory | What it shows |
| --- | --- |
| [examples/chat](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/chat) | The whole client in Go and TypeScript, tested end to end over an in-memory transport. Its config uses `message-matchers`. |
| [examples/shapes](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/shapes) | Schema edge cases, plus an OpenAPI 3.0 payload with `nullable` and a discriminator mapping. |
| [examples/kitchensink](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/kitchensink) | Every AsyncAPI 3.0 object in one document, valid against the official JSON Schema, with an external file pulled in through `$ref`. |
| [examples/official](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/official) | The 19 examples from the AsyncAPI specification repository, generated as they are in both languages. |

## Tests

The chat, shapes and kitchensink directories each carry a Go test (`<name>_test.go`) and a TypeScript test (`<name>.test.ts`, on `node:test`) that run against the committed output. The official examples have no tests of their own: `make lint` type-checks their Go, since golangci-lint runs over `./...`, the same test checks they match the generator, and `tsc` type-checks their TypeScript under the strict [`tsconfig.json`](https://github.com/psyb0t/gen-asyncapi/blob/main/tsconfig.json).

## License of the official specs

The specs in [examples/official](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/official) are unmodified copies from the AsyncAPI specification repository, under the Apache License 2.0. The [README](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/official/README.md) in that directory names the upstream commit, and its `LICENSE` and `NOTICE` files carry the terms. Everything around the specs is MIT, like the rest of the repository.
