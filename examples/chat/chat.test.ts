import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  AmbiguousMessageError,
  Client,
  type Handler,
  type MessagePosted,
  type ServerEvent,
  type Transport,
  UnknownMessageError,
  ValidationError,
  attachmentMetaValue,
  decodeReceivedMessage,
  MessageAck,
  MessageEvent,
  roomAddress,
  ServerProductionURL,
  setAttachmentMetaValue,
  validateJoinRequest,
  validateServerEvent,
} from "./chat.gen.ts";

const postedAt = "2026-01-02T03:04:05Z";

class RecordingTransport implements Transport {
  readonly sent: string[] = [];
  closed = false;

  send(data: string): void {
    this.sent.push(data);
  }

  close(): void {
    this.closed = true;
  }
}

function frame(value: unknown): string {
  return JSON.stringify(value);
}

function posted(
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    type: "message.posted",
    at: postedAt,
    sequence: 1,
    text: "hi",
    author: { id: "u1", role: "member" },
    ...overrides,
  };
}

function failsAt(path: string): (err: unknown) => boolean {
  return (err: unknown) => err instanceof ValidationError && err.path === path;
}

describe("sending", () => {
  it("checks the payload and writes it as JSON", async () => {
    const transport = new RecordingTransport();
    const client = new Client(transport);

    await client.joinRoom({ type: "join", id: "1", room: "lobby" });

    assert.deepEqual(transport.sent, [
      frame({ type: "join", id: "1", room: "lobby" }),
    ]);
  });

  it("sends nothing when the payload breaks the schema", async () => {
    const transport = new RecordingTransport();
    const client = new Client(transport);

    await assert.rejects(
      client.joinRoom({ type: "join", id: "1", room: "Lobby!" }),
      failsAt("join.room"),
    );
    assert.deepEqual(transport.sent, []);
  });

  it("closes the transport", () => {
    const transport = new RecordingTransport();

    new Client(transport).close();

    assert.equal(transport.closed, true);
  });
});

describe("validation", () => {
  const cases: { name: string; value: unknown; path: string }[] = [
    {
      name: "a missing required property",
      value: { type: "join", id: "1" },
      path: "JoinRequest",
    },
    {
      name: "an unknown property under additionalProperties false",
      value: { type: "join", id: "1", room: "x", extra: 1 },
      path: "JoinRequest",
    },
    {
      name: "a wrong const",
      value: { type: "say", id: "1", room: "x" },
      path: "JoinRequest.type",
    },
    {
      name: "a string one character over maxLength",
      value: { type: "join", id: "1", room: "a".repeat(65) },
      path: "JoinRequest.room",
    },
    {
      name: "an empty string under minLength 1",
      value: { type: "join", id: "1", room: "" },
      path: "JoinRequest.room",
    },
    {
      name: "a number where a string belongs",
      value: { type: "join", id: 1, room: "x" },
      path: "JoinRequest.id",
    },
    {
      name: "null for a property that is not nullable",
      value: { type: "join", id: "1", room: "x", nickname: null },
      path: "JoinRequest.nickname",
    },
    { name: "an array", value: [], path: "JoinRequest" },
    { name: "null", value: null, path: "JoinRequest" },
  ];

  for (const tc of cases) {
    it(`rejects ${tc.name}`, () => {
      assert.throws(() => validateJoinRequest(tc.value), failsAt(tc.path));
    });
  }

  it("accepts exactly maxLength characters, counted as code points", () => {
    validateJoinRequest({ type: "join", id: "1", room: "a".repeat(64) });
    validateServerEvent(posted({ text: "\u{1F600}" }));
  });

  it("narrows the type of a valid value", () => {
    const value: unknown = { type: "join", id: "1", room: "x", nickname: "n" };

    validateJoinRequest(value);

    assert.equal(value.room, "x");
  });

  it("picks the union variant by discriminator", () => {
    const value: unknown = posted();

    validateServerEvent(value);

    assert.equal(value.type, "message.posted");
  });

  it("rejects an unknown discriminator value", () => {
    assert.throws(
      () => validateServerEvent({ type: "nope" }),
      /unknown "type" "nope"/,
    );
  });

  it("checks nested values inside the picked variant", () => {
    assert.throws(
      () => validateServerEvent(posted({ author: { id: "u1", role: "x" } })),
      failsAt("ServerEvent.author.role"),
    );
  });

  it("checks array items", () => {
    const attachments = [{ kind: "file", meta: "{}" }, { kind: "x", meta: "" }];

    assert.throws(
      () => validateServerEvent(posted({ attachments })),
      failsAt("ServerEvent.attachments[1].kind"),
    );
  });

  it("rejects a malformed date-time", () => {
    assert.throws(
      () => validateServerEvent(posted({ at: "yesterday" })),
      /RFC 3339/,
    );
  });

  it("rejects a fractional integer and a number below minimum", () => {
    assert.throws(
      () => validateServerEvent(posted({ sequence: 1.5 })),
      failsAt("ServerEvent.sequence"),
    );
    assert.throws(
      () => validateServerEvent(posted({ sequence: 0 })),
      failsAt("ServerEvent.sequence"),
    );
  });
});

describe("receiving", () => {
  it("decodes a message by its const property", () => {
    const message = decodeReceivedMessage(
      frame({ type: "ack", id: "1", ok: true }),
    );

    assert.equal(message.name, MessageAck);
  });

  it("decodes a message that only strict decoding tells apart", () => {
    assert.equal(decodeReceivedMessage(frame(posted())).name, MessageEvent);
  });

  it("decodes messages told apart only by configured matchers", () => {
    const notice = decodeReceivedMessage(
      frame({ messages: [{ type: "notice", text: "a" }] }),
    );
    const alert = decodeReceivedMessage(
      frame({ messages: [{ type: "alert", text: "b" }] }),
    );

    assert.equal(notice.name, "notice");
    assert.equal(alert.name, "alert");
  });

  it("reports why each candidate turned a message down", () => {
    assert.throws(
      () => decodeReceivedMessage(frame({ hello: "world" })),
      (err: unknown) =>
        err instanceof UnknownMessageError && err.reasons.length > 0,
    );
  });

  it("rejects a pinned message whose payload is invalid", () => {
    assert.throws(
      () => decodeReceivedMessage(frame({ type: "ack", id: "1" })),
      UnknownMessageError,
    );
  });

  it("rejects text that is not JSON", () => {
    assert.throws(() => decodeReceivedMessage("{nope"), UnknownMessageError);
  });

  it("passes each message to its typed handler method", async () => {
    const events: ServerEvent[] = [];
    const acks: boolean[] = [];
    const handler: Handler = {
      handleEvent: (event) => {
        events.push(event);
      },
      handleAck: (ack) => {
        acks.push(ack.ok);
      },
    };
    const client = new Client(new RecordingTransport());

    await client.receive(frame(posted()), handler);
    await client.receive(frame({ type: "ack", id: "1", ok: false }), handler);

    const event = events[0];
    assert.equal(events.length, 1);
    assert.ok(event !== undefined && event.type === "message.posted");

    const message: MessagePosted = event;
    assert.equal(message.text, "hi");
    assert.deepEqual(acks, [false]);
  });

  it("ignores a message whose handler method is missing", async () => {
    const client = new Client(new RecordingTransport());

    await client.receive(frame({ type: "ack", id: "1", ok: true }), {});
  });

  it("gives unmatched frames to handleUnknownMessage", async () => {
    const unmatched: (UnknownMessageError | AmbiguousMessageError)[] = [];
    const client = new Client(new RecordingTransport());

    await client.receive(frame({ hello: "world" }), {
      handleUnknownMessage: (_data, error) => {
        unmatched.push(error);
      },
    });

    assert.equal(unmatched.length, 1);
    assert.ok(unmatched[0] instanceof UnknownMessageError);
  });

  it("throws unmatched frames when nothing handles them", async () => {
    const client = new Client(new RecordingTransport());

    await assert.rejects(
      client.receive(frame({ hello: "world" }), {}),
      UnknownMessageError,
    );
  });

  it("passes handler errors through", async () => {
    const client = new Client(new RecordingTransport());
    const failure = new Error("handler failed");

    await assert.rejects(
      client.receive(frame({ type: "ack", id: "1", ok: true }), {
        handleAck: () => {
          throw failure;
        },
      }),
      (err: unknown) => err === failure,
    );
  });
});

describe("JSON inside a string", () => {
  it("encodes and decodes the content schema", () => {
    const attachment = { kind: "file" as const, meta: "" };

    setAttachmentMetaValue(attachment, { size: 42, mime: "text/plain" });

    assert.deepEqual(attachmentMetaValue(attachment), {
      size: 42,
      mime: "text/plain",
    });
  });

  it("checks the decoded document", () => {
    assert.throws(
      () => attachmentMetaValue({ kind: "file", meta: '{"size":-1}' }),
      failsAt("Attachment.meta.size"),
    );
  });

  it("rejects a string that is not JSON", () => {
    assert.throws(
      () => attachmentMetaValue({ kind: "file", meta: "{" }),
      /not JSON/,
    );
  });
});

describe("servers and channels", () => {
  it("fills channel parameters", () => {
    assert.equal(roomAddress("lobby"), "rooms/lobby");
    assert.equal(roomAddress("{roomId}"), "rooms/{roomId}");
    assert.equal(ServerProductionURL, "wss://chat.example.com/ws");
  });
});
