---
title: "Troubleshooting"
description: "gen-asyncapi error messages from the command line, the config and the spec, plus the runtime decoding errors, each with its cause and fix."
nav_order: 10
---

# Troubleshooting

The command logs one `generation failed` line with the whole wrapped error, then exits with status 1. Each layer of the error ends with the file, line and function that added it, so the text in brackets points at the check that refused. The messages below leave those brackets out.

## Generator errors

| Message contains | Cause | Fix |
| --- | --- | --- |
| `want one spec path, got 0 arguments: invalid argument` | No spec path, or more than one positional argument. The usage line is printed first. | Pass exactly one spec path, after the flags. |
| `package: required config value is not set` | Go output with no `package` in the config and no `-package` flag. | Set `package`, or pass `-package=name`. |
| `field ouput not found in type codegen.Config: parse failed` | An unknown key in the config, here a typo of `output`. | Fix the key. Only the keys on the [Configuration](configuration.md) page are accepted. |
| `language "rust", want go or typescript: invalid value` | An unsupported `language` or `-language`. | Use `go` or `typescript`. |
| `generate: nothing selected, enable models or client: invalid value` | A `generate` block with both keys false or missing. | Turn on `models`, or drop the block to get everything. |
| `generate: client needs models: invalid value` | `client: true` without `models: true`. A key missing from the block is false. | Add `models: true`. |
| `message-matchers key "Nope" is not a message in the spec: not found` | A `message-matchers` key that is not a message key. | Use the key under `components/messages`, or `<channel>_<message>` for an inline message. |
| `message Notice match /messages/0/type: empty list: invalid value` | A matcher whose value is an empty list. | Give the list at least one value, or use a single value. |
| `asyncapi version "2.6.0" is not supported, only 3.x is: invalid value` | The spec is AsyncAPI 2.x. | Convert the spec to AsyncAPI 3, for example with the AsyncAPI CLI's `convert` command. |
| `read spec ...: no such file or directory` | The spec path is wrong. | Paths are relative to the directory `go generate` runs in, which is the directory of the `gen.go` file. |
| `schema at ... allows nothing: invalid value` | A schema that is the boolean `false`, so no value can ever match it. | Remove the field, or give it a real schema. |
| `const ... is not a scalar: invalid value` | A `const` holding an object or a list. The generator only handles scalar consts. | Use a string, number or boolean `const`, or drop it. |

## Runtime errors in Go

The generated code wraps sentinel errors, so check them with `errors.Is`:

| Sentinel | Meaning |
| --- | --- |
| `commerr.ErrRequiredFieldNotSet` | A required property is missing from the JSON. |
| `commerr.ErrValidationFailed` | A value broke a schema rule: pattern, length, bounds, enum, `const`, `additionalProperties: false` and the rest. |
| `asyncapirt.ErrUnknownMessage` | An incoming frame matched no message the client receives. The error holds every candidate's reason. |
| `asyncapirt.ErrAmbiguousMessage` | A frame decoded as more than one message. Add `message-matchers`, see [Message matching](message-matching.md). |
| `asyncapirt.ErrNoVariantMatched` | A union value fit none of its variants. |
| `asyncapirt.ErrAmbiguousVariant` | A union value without a discriminator fit several variants. |
| `asyncapirt.ErrUnknownDiscriminator` | A union's discriminator value names no variant. |

`commerr` is [`github.com/psyb0t/ctxerrors/commerr`](https://github.com/psyb0t/ctxerrors), and the `asyncapirt` sentinels are declared in [`asyncapirt/errors.go`](https://github.com/psyb0t/gen-asyncapi/blob/main/asyncapirt/errors.go).

## Runtime errors in TypeScript

- `ValidationError`: a value broke its schema. `path` names the field, such as `"JoinRequest.room"`.
- `UnknownMessageError` and `AmbiguousMessageError`: the TypeScript names for the two message errors above. `Client.receive` passes them to `handleUnknownMessage`, or rejects with them when the handler has no such method.

## Generated code is out of date

When `TestGeneratedExamplesAreCurrent` or the pipeline's generated-code check fails in this repository, the generator changed and the committed examples did not. Run `make generate` and commit the result. In your own project, the same happens after an upgrade: run `go generate ./...` again.
