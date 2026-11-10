---
title: "Message matching"
description: "How the generated decoder picks the message an incoming frame is: const matchers, strict decoding, opaque payloads, and message-matchers in the config."
nav_order: 6
---

# Message matching

A WebSocket frame carries no message name, so the generated decoder has to work out which message a frame is. `DecodeReceivedMessage` in Go and `decodeReceivedMessage` in TypeScript do it the same way, and the client's `Run` and `receive` call them for every frame.

## The order

The decoder works through the messages the client receives, from `receive` operations and from replies:

1. Messages whose matchers all hold. Matchers are the `const` values on the payload's top-level properties, plus your `message-matchers`.
2. The remaining messages, decoded strictly. Exactly one has to decode.
3. Messages with an opaque payload, like an Avro schema kept as raw JSON.

Nothing matching gives `asyncapirt.ErrUnknownMessage` (TypeScript: `UnknownMessageError`) with every candidate's reason joined in. Two matches give `asyncapirt.ErrAmbiguousMessage` (`AmbiguousMessageError`). The Go side lives in [`asyncapirt/match.go`](https://github.com/psyb0t/gen-asyncapi/blob/main/asyncapirt/match.go).

## Two messages with the same shape

When two messages have the same shape and only a value tells them apart, strict decoding accepts both and the frame is ambiguous. Point a matcher at that value with a JSON pointer in the [config](configuration.md#message-matchers):

```yaml
output-options:
  message-matchers:
    Notice:
      /messages/0/type: notice
    Alert:
      /messages/0/type: alert
```

A list value matches when the pointer holds any of its entries, so `/messages/0/type: [notice, info]` takes both. An empty list fails generation, and so does a key that names no message in the spec.

[examples/chat](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/chat) does exactly this in its [`config.yaml`](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/chat/config.yaml): its `Notice` and `Alert` messages differ only in the type of their first item.

## Unions inside a payload

The same idea applies one level down. A `oneOf` or `anyOf` union picks its variant from the `discriminator`, from a property every variant pins with `const`, or from the one variant that decodes. A value that fits none fails with `asyncapirt.ErrNoVariantMatched`, a value that fits several with `asyncapirt.ErrAmbiguousVariant`, and a discriminator value that names no variant with `asyncapirt.ErrUnknownDiscriminator`.
