# gen-asyncapi

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/gen-asyncapi.svg)](https://pkg.go.dev/github.com/psyb0t/gen-asyncapi)
[![CI](https://github.com/psyb0t/gen-asyncapi/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/gen-asyncapi/actions/workflows/pipeline.yml)
[![coverage](https://raw.githubusercontent.com/psyb0t/gen-asyncapi/badges/coverage.svg)](https://github.com/psyb0t/gen-asyncapi/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/gen-asyncapi/badges/version.svg)](https://github.com/psyb0t/gen-asyncapi/tags)
[![license](https://raw.githubusercontent.com/psyb0t/gen-asyncapi/badges/license.svg)](LICENSE)

Feed it an AsyncAPI 3 document and it spits out Go or TypeScript types, strict JSON decoding and a client for your WebSocket API. Run it from `go generate` with a small config file and the path to your spec, and the generated file lands next to your code.

Decoding is strict. A frame either matches its schema or you get an error that names every message it tried and why each one said no.

Spec loading comes from [go-asyncapi](https://github.com/psyb0t/go-asyncapi). Grab that one directly if you want to read AsyncAPI 3 documents in Go without generating anything.

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [Config](#config)
- [Examples](#examples)
  - [Structs and strict decoding](#structs-and-strict-decoding)
  - [Enums](#enums)
  - [Unions](#unions)
  - [JSON inside a string](#json-inside-a-string)
  - [Sending](#sending)
  - [Receiving](#receiving)
  - [Telling messages apart](#telling-messages-apart)
  - [Servers and channels](#servers-and-channels)
- [TypeScript](#typescript)
- [What gets generated](#what-gets-generated)
- [Example packages](#example-packages)
- [Development](#development)
- [License](#license)

## Install

```sh
go install github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest
```

As a Go tool in your module, so `go tool gen-asyncapi` runs the version your `go.mod` pins:

```sh
go get -tool github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest
```

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

Run `go generate ./...` and `chat.gen.go` appears. The spec is the last argument, and paths are relative to the directory `go generate` runs in. `-language`, `-package` and `-o` override the config, or stand in for it:

```sh
gen-asyncapi -package=chat -o=chat.gen.go asyncapi.yaml
```

## Config

```yaml
language: go              # go or typescript, defaults to go
package: chat             # Go package name, required for Go
output: chat.gen.go       # defaults to asyncapi.gen.go, or asyncapi.gen.ts
generate:
  models: true            # types, codecs, validation (on by default)
  client: true            # decoding, Handler, Dispatch, Client (on by default, needs models)
output-options:
  message-matchers:       # see "Telling messages apart"
    Notice:
      /messages/0/type: notice
```

## Examples

Every snippet below comes from [examples/chat](examples/chat): the spec is [chat.yaml](examples/chat/chat.yaml), the output is [chat.gen.go](examples/chat/chat.gen.go), and [chat_test.go](examples/chat/chat_test.go) runs all of it.

### Structs and strict decoding

```yaml
JoinRequest:
  type: object
  additionalProperties: false
  required: [type, id, room]
  properties:
    type: { type: string, const: join }
    id: { type: string, format: uuid }
    room: { type: string, minLength: 1, maxLength: 64, pattern: '^[a-z0-9-]+$' }
    nickname: { type: string }
```

```go
type JoinRequest struct {
	Type     string  `json:"type"`
	ID       string  `json:"id"`
	Room     string  `json:"room"`
	Nickname *string `json:"nickname,omitempty"`
}
```

Required fields are values, optional ones are pointers. The `const` fills itself in when you marshal:

```go
data, _ := json.Marshal(chat.JoinRequest{ID: "1", Room: "lobby"})
// {"type":"join","id":"1","room":"lobby"}
```

Decoding checks everything the schema says, and the errors wrap `ctxerrors/commerr` sentinels:

```go
var req chat.JoinRequest

err := json.Unmarshal([]byte(`{"type":"join","id":"1"}`), &req)
errors.Is(err, commerr.ErrRequiredFieldNotSet) // true, room is missing

err = json.Unmarshal([]byte(`{"type":"join","id":"1","room":"Lobby!"}`), &req)
errors.Is(err, commerr.ErrValidationFailed) // true, pattern

err = json.Unmarshal([]byte(`{"type":"join","id":"1","room":"x","extra":1}`), &req)
errors.Is(err, commerr.ErrValidationFailed) // true, additionalProperties: false
```

Every type also has `Validate()`, so you can check a value you built yourself before sending it.

### Enums

```yaml
role:
  type: string
  enum: [owner, member, guest]
```

```go
type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleMember MemberRole = "member"
	MemberRoleGuest  MemberRole = "guest"
)
```

A value outside the list fails decoding with `commerr.ErrValidationFailed`.

### Unions

```yaml
ServerEvent:
  oneOf:
    - $ref: '#/components/schemas/MessagePosted'   # type: { const: message.posted }
    - $ref: '#/components/schemas/MemberJoined'    # type: { const: member.joined }
    - $ref: '#/components/schemas/Ping'            # type: { const: ping }
  discriminator: type
```

The union gets `FromX`, `AsX`, `Discriminator` and `Value`. `Value` decodes whichever variant the `type` field picks:

```go
var event chat.ServerEvent
if err := json.Unmarshal(frame, &event); err != nil {
	return err // unknown type, or the picked variant is invalid
}

value, err := event.Value()
if err != nil {
	return err
}

switch e := value.(type) {
case chat.MessagePosted:
	fmt.Println(e.Author.ID, "said", e.Text)
case chat.MemberJoined:
	fmt.Println(e.Member.ID, "joined")
case chat.Ping:
	// keepalive
}
```

Building one goes the other way:

```go
var event chat.ServerEvent
err := event.FromPing(chat.Ping{}) // marshals to {"type":"ping"}
```

Without a `discriminator`, a property every variant pins with `const` does the same job. With neither, exactly one variant has to decode.

### JSON inside a string

Some APIs send a JSON document as a string field. `contentMediaType` plus `contentSchema` says what is inside:

```yaml
meta:
  type: string
  contentMediaType: application/json
  contentSchema:
    $ref: '#/components/schemas/AttachmentMeta'
```

```go
meta, err := attachment.MetaValue() // decodes and validates the string as AttachmentMeta

err = attachment.SetMetaValue(chat.AttachmentMeta{Size: 42}) // encodes it back into the string
```

### Sending

Every message of a `send` operation becomes a client method. It validates the payload, encodes it and writes it to the connection:

```go
conn, err := gorillaws.Dial(ctx, chat.ServerProductionURL, gorillaws.WithBearerToken(token))
if err != nil {
	return err
}

client := chat.NewClient(conn)
defer client.Close()

err = client.JoinRoom(ctx, chat.JoinRequest{ID: id, Room: "lobby"})
```

`gorillaws.Dial` also takes `WithHeader`, `WithQueryParam`, `WithSubprotocols`, `WithBinaryFrames`, `WithReadLimit` and `WithDialer`. Any type with `WriteMessage`, `ReadMessage` and `Close` works as a transport.

### Receiving

Messages from `receive` operations and from replies land in a `Handler`. Embed `UnimplementedHandler` and write only the methods you care about:

```go
type handler struct {
	chat.UnimplementedHandler
}

func (handler) HandleEvent(ctx context.Context, event chat.ServerEvent) error {
	value, err := event.Value()
	if err != nil {
		return err
	}

	log.Printf("event: %+v", value)

	return nil
}

func (handler) HandleAck(ctx context.Context, ack chat.Ack) error {
	if !ack.Ok {
		log.Printf("request %s failed", ack.ID)
	}

	return nil
}

err := client.Run(ctx, handler{}) // reads until ctx ends or the connection drops
```

Frames that match nothing go to `HandleUnknownMessage` with the reason. The default logs a warning and keeps reading. Without the client, `chat.DecodeReceivedMessage(frame)` and `chat.Dispatch(ctx, h, msg)` do the same two steps by hand.

### Telling messages apart

`DecodeReceivedMessage` works through the messages the client receives:

1. Messages whose matchers all hold. Matchers are the `const` values on the payload's top-level properties, plus your `message-matchers`.
2. The remaining messages, decoded strictly. Exactly one has to decode.
3. Messages with an opaque payload, like an Avro schema kept as raw JSON.

Nothing matching gives `asyncapirt.ErrUnknownMessage` with every candidate's reason joined in. Two matches give `asyncapirt.ErrAmbiguousMessage`.

When two messages have the same shape and only a value tells them apart, point a matcher at that value with a JSON pointer:

```yaml
output-options:
  message-matchers:
    Notice:
      /messages/0/type: notice
    Alert:
      /messages/0/type: alert
```

### Servers and channels

From [examples/kitchensink](examples/kitchensink/kitchensink.yaml):

```yaml
servers:
  production:
    host: '{region}.example.com:{port}'
    pathname: /v{version}/ws
    protocol: wss
    variables:
      region: { enum: [eu, us], default: eu }
      port: { enum: ['443', '8443'], default: '443' }
      version: { default: '2' }
channels:
  commands:
    address: tenants/{tenantId}/commands
```

```go
kitchensink.ServerProductionURL                         // "wss://eu.example.com:443/v2/ws"
kitchensink.ServerProductionURLWith("us", "8443", "1")  // "wss://us.example.com:8443/v1/ws"
kitchensink.CommandsAddress("acme")                     // "tenants/acme/commands"
kitchensink.CommandsAddressTemplate                     // "tenants/{tenantId}/commands"
```

## TypeScript

Set `language: typescript`, or pass `-language=typescript`, and the same spec turns into one `.ts` file with zero dependencies. A TypeScript project needs no Go module, just Go installed. Run it straight from the repo:

```sh
go run github.com/psyb0t/gen-asyncapi/cmd/gen-asyncapi@latest -language=typescript -o=src/chat.gen.ts asyncapi.yaml
```

Pin a version instead of `@latest` and put the line in a `package.json` script. A Go server and its web client can generate from one spec in the same `gen.go`:

```go
//go:generate go tool gen-asyncapi --config=config.yaml chat.yaml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=chat.gen.ts chat.yaml
```

Schemas become interfaces and unions. Const properties are literal types, so a `oneOf` with a discriminator narrows in a `switch`:

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

Every type gets a `validate<Type>` function. It throws a `ValidationError` naming the bad value, and narrows the type when the value passes:

```ts
const value: unknown = JSON.parse(text);

validateJoinRequest(value); // a bad room throws with err.path "JoinRequest.room"
value.room;                 // typed from here on
```

The client takes anything with `send` and `close`, so a browser `WebSocket` works as it is. Every handler method is optional:

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

Without `handleUnknownMessage`, `receive` rejects with the `UnknownMessageError` or `AmbiguousMessageError`. `decodeReceivedMessage(text)` and `dispatch(handler, message)` do the same two steps by hand.

The file type-checks under `strict` plus `exactOptionalPropertyTypes`, `noUncheckedIndexedAccess`, `noUnusedLocals`, `verbatimModuleSyntax` and `erasableSyntaxOnly`. It sticks to ES2020, so it runs in browsers and straight from Node.js.

## What gets generated

| Spec | Go |
|---|---|
| object schema | struct with strict `UnmarshalJSON`, `Validate`, and `MarshalJSON` when it has consts |
| `enum` | named type plus constants |
| `oneOf` / `anyOf` | union type with `FromX`, `AsX`, `Value` |
| `allOf` | one merged struct |
| `contentSchema` string | `<Field>Value()` and `Set<Field>Value()` |
| message | `Message<Name>` constant, headers type, correlation ID location |
| `send` operation | `Client` method |
| `receive` operation, reply | `Handler` method, `DecodeReceivedMessage`, `Dispatch`, `Client.Run` |
| server | `Server<Name>URL`, plus `URLTemplate` and `URLWith` for variables |
| channel | `<Channel>AddressTemplate`, plus `<Channel>Address(...)` for parameters |

The TypeScript file has the same pieces under TypeScript names: `validate<Type>`, `<Type>Values` for enums, `server<Name>URLWith(...)`, `<channel>Address(...)`, `decodeReceivedMessage`, `dispatch` and `Client.receive`.

Every object and keyword, one row each, is in [docs/spec-coverage.md](docs/spec-coverage.md).

## Example packages

- [examples/chat](examples/chat): the whole client in Go and TypeScript, tested end to end over an in-memory transport.
- [examples/shapes](examples/shapes): schema edge cases, plus an OpenAPI 3.0 payload with `nullable` and a discriminator mapping.
- [examples/kitchensink](examples/kitchensink): every AsyncAPI 3.0 object in one document, valid against the official JSON Schema.
- [examples/official](examples/official): the 19 examples from the AsyncAPI spec repository, generated as they are in both languages. CI builds and type-checks them and checks they match the generator.

## Development

Everything runs in the `Dockerfile.dev` image. You need Docker and Make.

```sh
make help           # list targets
make generate       # regenerate the examples
make lint           # shellcheck, shfmt, go fix, golangci-lint
make test-coverage  # go test -race with a 90% coverage gate, then make test-ts
make test-ts        # tsc on the generated TypeScript, then its node:test suites
make sec            # gitleaks, govulncheck, pnpm audit, semgrep
```

The TypeScript checks run `tsc` and Node.js types from `pnpm-lock.yaml`, installed in the dev image. The `pkg-*` targets, such as `make pkg-add PKG=name@version`, change them.

## License

[MIT](LICENSE). The specs in [examples/official](examples/official) come from the AsyncAPI specification repository under Apache 2.0.
