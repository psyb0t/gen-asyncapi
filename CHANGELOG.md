# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.2.0, 2026-10-15

- Renamed the project from `gogen-asyncapi` to `gen-asyncapi`, since it now writes more than Go. Change imports, the `go get -tool` line and every `go:generate` line to the new names, then regenerate:

  | Old | New |
  |---|---|
  | `github.com/psyb0t/gogen-asyncapi/asyncapirt` | `github.com/psyb0t/gen-asyncapi/asyncapirt` |
  | `github.com/psyb0t/gogen-asyncapi/asyncapirt/gorillaws` | `github.com/psyb0t/gen-asyncapi/asyncapirt/gorillaws` |
  | `github.com/psyb0t/gogen-asyncapi/cmd/gogen-asyncapi` | `github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi` |
  | `go tool gogen-asyncapi` | `go tool gen-asyncapi` |
- Added TypeScript output. `language: typescript` in the config, or `-language=typescript` on the command line, writes one dependency-free `.ts` file with interfaces and unions, a `validate<Type>` assert function per type, message decoding, a `Handler` and a `Client` that takes a browser `WebSocket` as its transport. `package` is optional for TypeScript, and the default output is `asyncapi.gen.ts`.
- Added the `-language` flag.
- Moved the spec loader to its own module, [go-asyncapi](https://github.com/psyb0t/go-asyncapi), so other tools can read AsyncAPI 3 documents with it. The generator depends on its v0.1.0.
- Every example, the 19 official ones included, now also generates TypeScript. CI type-checks it under strict settings and runs `node:test` suites for the chat, shapes and kitchensink examples.
- Added Node.js, pnpm and a locked TypeScript toolchain to the dev image, the `make test-ts` and `pkg-*` targets, and `pnpm audit` in `make sec`.

## v0.1.0, 2026-10-15

- Added the `gogen-asyncapi` command, which generates Go types, strict JSON codecs, validation, message decoding and a client from an AsyncAPI 3 document. It takes a config file and the spec path, and installs with `go get -tool` or `go install`.
- Added the `asyncapirt` runtime package the generated code uses, and `asyncapirt/gorillaws`, a transport over gorilla/websocket with header, bearer token, query and subprotocol options.
- Covered the AsyncAPI 3.0 objects and JSON Schema keywords listed in `docs/spec-coverage.md`.
- Added examples, including one document that uses every AsyncAPI 3.0 object and the 19 official AsyncAPI examples, all generated and tested in CI.
