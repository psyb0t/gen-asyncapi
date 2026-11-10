---
title: "TypeScript output"
description: "Generate one dependency-free TypeScript file from an AsyncAPI 3 spec: interfaces, literal unions, validate functions and a client for a browser WebSocket."
nav_order: 5
---

# TypeScript output

Set `language: typescript`, or pass `-language=typescript`, and the same spec turns into one `.ts` file with zero dependencies. `package` is ignored, and the default output is `asyncapi.gen.ts`.

## Generate

A TypeScript project needs no Go module, just Go installed. Run the generator straight from the module proxy:

```sh
go run github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest \
  -language=typescript \
  -o=src/chat.gen.ts \
  asyncapi.yaml
```

Pin a version instead of `@latest` and put the line in a `package.json` script. A Go server and its web client can generate from one spec in the same `gen.go`:

```go
//go:generate go tool gen-asyncapi --config=config.yaml chat.yaml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=chat.gen.ts chat.yaml
```

The snippets below come from [examples/chat](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/chat): the output is [chat.gen.ts](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/chat/chat.gen.ts), and [chat.test.ts](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/chat/chat.test.ts) runs it.

## Types

Schemas become interfaces and unions. Optional properties are `?:`, nullable ones add `| null`, and `const` properties are literal types, so a `oneOf` with a discriminator narrows in a `switch`:

```ts
export interface JoinRequest {
  type: "join";
  id: string;
  room: string;
  nickname?: string;
}

export type ServerEvent = MessagePosted | MemberJoined | Ping;
export type MemberRole = "owner" | "member" | "guest";
```

Every enum also gets a `<Type>Values` array, such as `MemberRoleValues`. Unlike Go, the caller writes `const` properties out, because they are part of the type.

## Validation

Every type gets a `validate<Type>` function. It throws a `ValidationError` naming the bad value in `path`, and narrows the type when the value passes:

```ts
const value: unknown = JSON.parse(text);

validateJoinRequest(value); // a bad room throws with err.path "JoinRequest.room"
value.room;                 // typed from here on
```

A string field with `contentMediaType: application/json` and a `contentSchema` gets a reader and a writer, such as `attachmentMetaValue(attachment)` and `setAttachmentMetaValue(attachment, meta)`.

## Client

The client takes anything with `send(data: string)` and `close()`, the `Transport` interface in the generated file, so a browser `WebSocket` works as it is. Every handler method is optional:

```ts
const socket = new WebSocket(ServerProductionURL);
const client = new Client(socket);

const handler: Handler = {
  handleEvent: (event) => {
    if (event.type === "message.posted") {
      render(event.author.id, event.text);
    }
  },
  handleUnknownMessage: (_data, error) => {
    report(error);
  },
};

socket.addEventListener("message", (message) => {
  client.receive(message.data, handler).catch(report);
});

await client.joinRoom({ type: "join", id: crypto.randomUUID(), room: "lobby" });
```

Each send method validates the payload before it calls `send`. Without `handleUnknownMessage`, `receive` rejects with the `UnknownMessageError` or `AmbiguousMessageError`. `decodeReceivedMessage(text)` and `dispatch(handler, message)` do the same two steps by hand. [Message matching](message-matching.md) explains how a frame is matched to a message.

## Servers and channels

Servers and channels become constants and functions, named like their Go counterparts with a lowercase first letter on the functions:

```ts
ServerProductionURL                          // "wss://eu.example.com:443/v2/ws"
serverProductionURLWith("us", "8443", "1")   // "wss://us.example.com:8443/v1/ws"
commandsAddress("acme")                      // "tenants/acme/commands"
CommandsAddressTemplate                      // "tenants/{tenantId}/commands"
```

Those come from [kitchensink.gen.ts](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/kitchensink/kitchensink.gen.ts).

## Compiler settings

The file type-checks under `strict` plus `exactOptionalPropertyTypes`, `noUncheckedIndexedAccess`, `noUnusedLocals`, `verbatimModuleSyntax` and `erasableSyntaxOnly`. It sticks to ES2020 and needs no Node.js types, so it runs in browsers and straight from Node.js. CI checks every generated file against the repository's [`tsconfig.json`](https://github.com/psyb0t/gen-asyncapi/blob/main/tsconfig.json).
