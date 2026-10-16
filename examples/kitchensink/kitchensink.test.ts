import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  Client,
  commandsAddress,
  decodeReceivedMessage,
  MessageAvroReading,
  MessageHeartbeat,
  ServerProductionURL,
  serverProductionURLWith,
  type Transport,
  ValidationError,
  validateCommandPayload,
  validateDimensions,
  validateItem,
  validateRenamed,
} from "./kitchensink.gen.ts";

const int32Max = 2147483647;
const maxStringPayload = 64;

function failsAt(path: string): (err: unknown) => boolean {
  return (err: unknown) => err instanceof ValidationError && err.path === path;
}

function item(
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    sku: "ABC-1234",
    price: 19.99,
    quantity: 1,
    tags: ["ab"],
    dimensions: { width: 1, height: 1 },
    ...overrides,
  };
}

describe("object constraints", () => {
  it("accepts values on every limit", () => {
    validateItem(item({ quantity: 1000, price: 0.01 }));
    validateItem(
      item({
        tags: ["aa", "bb", "cc", "dd", "ee", "ff", "gg", "hh", "ii", "jj"],
      }),
    );
    validateItem(item({ aliases: ["a"] }));
  });

  const cases: {
    name: string;
    overrides: Record<string, unknown>;
    path: string;
  }[] = [
    {
      name: "a price that is not a multiple of 0.01",
      overrides: { price: 0.015 },
      path: "Item.price",
    },
    {
      name: "a zero price under exclusiveMinimum",
      overrides: { price: 0 },
      path: "Item.price",
    },
    {
      name: "a quantity one over maximum",
      overrides: { quantity: 1001 },
      path: "Item.quantity",
    },
    {
      name: "duplicate tags under uniqueItems",
      overrides: { tags: ["ab", "ab"] },
      path: "Item.tags",
    },
    {
      name: "an empty tag list under minItems",
      overrides: { tags: [] },
      path: "Item.tags",
    },
    {
      name: "an empty optional list under minItems",
      overrides: { aliases: [] },
      path: "Item.aliases",
    },
    {
      name: "a map key outside the key pattern",
      overrides: { attributes: { color: "red" } },
      path: "Item.attributes.color",
    },
    {
      name: "a map one entry over maxProperties",
      overrides: { counters: { a: 1, b: 2, c: 3, d: 4 } },
      path: "Item.counters",
    },
    {
      name: "a map value below minimum",
      overrides: { counters: { a: -1 } },
      path: "Item.counters.a",
    },
    {
      name: "an unknown string enum value",
      overrides: { status: "gone" },
      path: "Item.status",
    },
    {
      name: "an unknown numeric enum value",
      overrides: { priority: 4 },
      path: "Item.priority",
    },
  ];

  for (const tc of cases) {
    it(`rejects ${tc.name}`, () => {
      assert.throws(() => validateItem(item(tc.overrides)), failsAt(tc.path));
    });
  }

  it("enforces minProperties and maxProperties", () => {
    const { dimensions: _dimensions, ...four } = item();
    const thirteen = item({
      weight: 1,
      attributes: {},
      counters: {},
      legacyCode: "x",
      password: "12345678",
      status: "active",
      priority: 1,
      ratio: 0.5,
    });

    assert.throws(() => validateItem(four), failsAt("Item"));
    assert.throws(() => validateItem(thirteen), failsAt("Item"));
  });

  it("checks the int32 range", () => {
    validateDimensions({ width: 1, height: int32Max });
    assert.throws(
      () => validateDimensions({ width: 1, height: int32Max + 1 }),
      failsAt("Dimensions.height"),
    );
  });

  it("allows null only where the schema does", () => {
    validateDimensions({ width: 1, height: 1, depth: null });
    assert.throws(
      () => validateDimensions({ width: 1, height: null }),
      failsAt("Dimensions.height"),
    );
  });

  it("enforces dependentRequired", () => {
    const renamed = {
      kind: "renamed",
      id: "1",
      occurredAt: "2026-01-02T03:04:05+02:00",
      to: "b",
    };

    assert.throws(() => validateRenamed(renamed), failsAt("Renamed"));
  });
});

describe("unions without a discriminator", () => {
  it("accepts a value exactly one variant matches", () => {
    validateCommandPayload("go");
    validateCommandPayload(3);
    validateCommandPayload({ width: 1, height: 2 });
  });

  it("rejects a value no variant matches", () => {
    assert.throws(() => validateCommandPayload(true), /no variant matched/);
    assert.throws(
      () => validateCommandPayload("x".repeat(maxStringPayload + 1)),
      /no variant matched/,
    );
  });
});

describe("receiving", () => {
  it("decodes a message pinned by a boolean const", () => {
    const message = decodeReceivedMessage(JSON.stringify({ beat: true }));

    assert.equal(message.name, MessageHeartbeat);
  });

  it("falls back to the raw-payload message when nothing else fits", () => {
    const message = decodeReceivedMessage(JSON.stringify({ anything: [1] }));

    assert.equal(message.name, MessageAvroReading);
    assert.deepEqual(message.payload, { anything: [1] });
  });
});

describe("sending", () => {
  it("awaits an asynchronous transport", async () => {
    const sent: string[] = [];
    const transport: Transport = {
      send: async (data) => {
        sent.push(data);
      },
      close: () => {},
    };
    const client = new Client(transport);

    await client.sendCommand({ commandId: "c1", action: "start" });

    assert.deepEqual(sent, ['{"commandId":"c1","action":"start"}']);
  });
});

describe("servers and channels", () => {
  it("fills server variables and channel parameters", () => {
    assert.equal(ServerProductionURL, "wss://eu.example.com:443/v2/ws");
    assert.equal(
      serverProductionURLWith("us", "8443", "1"),
      "wss://us.example.com:8443/v1/ws",
    );
    assert.equal(commandsAddress("acme"), "tenants/acme/commands");
  });
});
