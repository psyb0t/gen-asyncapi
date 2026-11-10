---
title: "Go output"
description: "What the generated Go looks like: structs with strict decoding, enums, unions, JSON inside strings, the client, the Handler, servers and channels."
nav_order: 4
---

# Go output

Every snippet below comes from [examples/chat](https://github.com/psyb0t/gen-asyncapi/tree/main/examples/chat): the spec is [chat.yaml](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/chat/chat.yaml), the output is [chat.gen.go](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/chat/chat.gen.go), and [chat_test.go](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/chat/chat_test.go) runs all of it.

## Structs and strict decoding

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

## Enums

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

## Unions

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

## JSON inside a string

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

## Sending

Every message of a `send` operation becomes a client method. It validates the payload, encodes it and writes it to the connection:

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

`gorillaws.Dial` also takes `WithHeader`, `WithQueryParam`, `WithSubprotocols`, `WithBinaryFrames`, `WithReadLimit` and `WithDialer`. Any type with `WriteMessage`, `ReadMessage` and `Close` works as a transport: that is the `asyncapirt.Transport` interface in [`asyncapirt/transport.go`](https://github.com/psyb0t/gen-asyncapi/blob/main/asyncapirt/transport.go). A transport must allow one goroutine reading while others write.

## Receiving

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

Frames that match nothing go to `HandleUnknownMessage` with the reason. The default logs a warning and keeps reading. Without the client, `chat.DecodeReceivedMessage(frame)` and `chat.Dispatch(ctx, h, msg)` do the same two steps by hand. [Message matching](message-matching.md) explains how a frame is matched to a message.

## Servers and channels

From [examples/kitchensink](https://github.com/psyb0t/gen-asyncapi/blob/main/examples/kitchensink/kitchensink.yaml):

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
