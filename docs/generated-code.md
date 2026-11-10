---
title: "What gets generated"
description: "Each part of an AsyncAPI 3 spec and what gen-asyncapi turns it into, in Go and in TypeScript, from schemas and messages to servers and channels."
nav_order: 7
has_children: true
---

# What gets generated

One spec in, one file out. The Go and TypeScript files come from the same model, so they hold the same pieces under each language's names.

| Spec | Go | TypeScript |
|---|---|---|
| object schema | struct with strict `UnmarshalJSON`, `Validate`, and `MarshalJSON` when it has `const` properties | `interface` plus `validate<Type>` |
| `enum` | named type plus constants | literal union plus `<Type>Values` |
| `oneOf` / `anyOf` | union type with `From<X>`, `As<X>`, `Value`, `Discriminator` | union type plus `validate<Type>` |
| `allOf` | one merged struct | one merged interface |
| `contentSchema` string | `<Field>Value()` and `Set<Field>Value()` | `<type><Field>Value()` and `set<Type><Field>Value()` |
| message | `Message<Name>` constant, headers type, correlation ID location | `Message<Name>` constant |
| `send` operation | `Client` method | `Client` method |
| `receive` operation, reply | `Handler` method, `DecodeReceivedMessage`, `Dispatch`, `Client.Run` | optional `Handler` method, `decodeReceivedMessage`, `dispatch`, `Client.receive` |
| server | `Server<Name>URL`, plus `URLTemplate` and `URLWith` for variables | `Server<Name>URL`, plus `server<Name>URLWith` |
| channel | `<Channel>AddressTemplate`, plus `<Channel>Address(...)` for parameters | `<Channel>AddressTemplate`, plus `<channel>Address(...)` |

`generate: {models: true}` in the [config](configuration.md) keeps the types, codecs, validation and message constants, and leaves out message decoding, the `Handler`, `Dispatch` and the `Client`.

## Go details

- Required fields are values, optional and nullable ones are pointers.
- Decoding checks required properties, `additionalProperties: false`, string length and pattern, numeric bounds, `multipleOf`, item counts, `uniqueItems`, property counts and `dependentRequired`. Failures wrap `commerr.ErrValidationFailed` or `commerr.ErrRequiredFieldNotSet` from [ctxerrors](https://github.com/psyb0t/ctxerrors).
- Constraints of an optional array, map or raw JSON field are skipped when the field is absent, and of a nullable one when it is null.
- `UnimplementedHandler` ignores every message. Its `HandleUnknownMessage` logs a warning and keeps reading.
- Keywords the generated code does not enforce are listed in the type's doc comment, under "Not enforced by the generated code".

## TypeScript details

- Optional properties are `?:`, nullable ones add `| null`, `const` properties are literal types.
- Validators throw `ValidationError` with a `path` such as `"JoinRequest.room"`.
- `Transport` is anything with `send(data: string)` and `close()`. A browser `WebSocket` fits.
- Runtime helpers are copied into the file only when the generated code calls them, so a small spec gives a small file.

Every AsyncAPI object and schema keyword, one row each, is on the [AsyncAPI 3 coverage](spec-coverage.md) page.
