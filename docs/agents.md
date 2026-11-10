---
title: "Agent integrations"
description: "Install the gen-asyncapi agent skill in Claude Code, Codex or OpenClaw, so a coding agent can set up the generator and use the Go and TypeScript it writes."
nav_order: 9
---

# Agent integrations

The [skill](https://github.com/psyb0t/gen-asyncapi/tree/main/.agents/skills/gen-asyncapi) teaches a coding agent how to set up the generator and use the code it writes, in Go and TypeScript. It works in any agent that reads `.agents/skills/`, and installs natively in the clients below. Its [`references/setup.md`](https://github.com/psyb0t/gen-asyncapi/blob/main/.agents/skills/gen-asyncapi/references/setup.md) holds the install routes, the config keys and the generated API the agent works from.

## Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install gen-asyncapi@psyb0t
```

## Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add gen-asyncapi@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repository.

## OpenClaw

The skill is published to ClawHub on every release:

```bash
openclaw skills install @psyb0t/gen-asyncapi
```

## What the skill may do

The skill declares what it touches. It reads the spec, the files the spec references and the generator config, and writes the one output file. It runs the Go toolchain (`go get -tool`, `go generate`, `go run`, `go build`, `go test`). Network access is limited to fetching the module from your Go module proxy; generation itself is offline. It tells the agent never to hand-edit a generated file and to keep tokens out of the spec and the config.
