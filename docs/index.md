---
title: "Home"
description: "gen-asyncapi turns an AsyncAPI 3 document into Go or TypeScript types, strict JSON decoding and a WebSocket client, run from go generate."
nav_order: 1
permalink: /
seo:
  type: WebSite
---

# gen-asyncapi

Feed it an AsyncAPI 3 document and it spits out Go or TypeScript types, strict JSON decoding and a client for your WebSocket API. Run it from `go generate` with a small config file and the path to your spec, and the generated file lands next to your code.

Decoding is strict. A frame either matches its schema or you get an error that names every message it tried and why each one said no.

Spec loading comes from [go-asyncapi](https://github.com/psyb0t/go-asyncapi). Grab that one directly if you want to read AsyncAPI 3 documents in Go without generating anything.

## Install

```sh
go get -tool github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest
```

That adds the generator to your module as a Go tool, so `go tool gen-asyncapi` runs the version your `go.mod` pins. [Install and run](install.md) has the other routes, including a TypeScript project with no Go module.

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

Run `go generate ./...` and `chat.gen.go` appears. Then dial the server and send:

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

## Where to go next

| Page | What it covers |
| --- | --- |
| [Install and run](install.md) | Install routes, the command line, `go generate`, logging, upgrades |
| [Configuration](configuration.md) | Every config key, its default and what fails validation |
| [Go output](go.md) | Structs, enums, unions, sending and receiving in Go |
| [TypeScript output](typescript.md) | Interfaces, validators and the browser client |
| [Message matching](message-matching.md) | How incoming frames are told apart, and `message-matchers` |
| [What gets generated](generated-code.md) | Each spec part and what it becomes in both languages |
| [AsyncAPI 3 coverage](spec-coverage.md) | Every AsyncAPI object and schema keyword, one row each |
| [Example packages](examples.md) | The committed examples and what each one proves |
| [Agent integrations](agents.md) | The agent skill for Claude Code, Codex and OpenClaw |
| [Troubleshooting](troubleshooting.md) | Generator errors and runtime decoding errors, with fixes |
| [Development](development.md) | Make targets, the dev image, CI and releases |
