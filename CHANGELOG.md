# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.0, 2026-10-15

- Added the `gogen-asyncapi` command, which generates Go types, strict JSON codecs, validation, message decoding and a client from an AsyncAPI 3 document. It takes a config file and the spec path, and installs with `go get -tool` or `go install`.
- Added the `asyncapirt` runtime package the generated code uses, and `asyncapirt/gorillaws`, a transport over gorilla/websocket with header, bearer token, query and subprotocol options.
- Covered the AsyncAPI 3.0 objects and JSON Schema keywords listed in `docs/spec-coverage.md`.
- Added examples, including one document that uses every AsyncAPI 3.0 object and the 19 official AsyncAPI examples, all generated and tested in CI.
