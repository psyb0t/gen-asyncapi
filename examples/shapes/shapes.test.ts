import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  decodeReceivedMessage,
  feedAddress,
  MessageFrame,
  MessageReading,
  MessageTree,
  ValidationError,
  validateBatch,
  validateMeasurement,
  validateNode,
  validateSensorReading,
} from "./shapes.gen.ts";

function failsAt(path: string): (err: unknown) => boolean {
  return (err: unknown) => err instanceof ValidationError && err.path === path;
}

function reading(
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    value: 1,
    unit: null,
    tags: ["t"],
    level: -1,
    ratio: 0,
    labels: { a: 1 },
    ...overrides,
  };
}

describe("unions", () => {
  it("accepts each variant of an anyOf", () => {
    validateMeasurement(1.5);
    validateMeasurement("12.5cm");
    validateMeasurement({ x: 1, y: 2 });
  });

  it("rejects a value no variant matches", () => {
    assert.throws(() => validateMeasurement("cm"), /no variant matched/);
    assert.throws(
      () => validateMeasurement({ x: 1, y: 2, z: 3 }),
      /no variant matched/,
    );
  });

  it("checks array and map variants", () => {
    validateBatch(["a"]);
    validateBatch({ a: 1 });
    assert.throws(() => validateBatch({ a: "x" }), /no variant matched/);
  });

  it("picks variants through an OpenAPI discriminator mapping", () => {
    validateSensorReading(reading({ source: { kind: "alpha", alpha: 1 } }));
    validateSensorReading(reading({ source: { kind: "SensorC", gamma: true } }));
    assert.throws(
      () => validateSensorReading(reading({ source: { kind: "alpha", beta: "b" } })),
      failsAt("SensorReading.source"),
    );
  });
});

describe("OpenAPI 3.0 keywords", () => {
  it("keeps a nullable required property required", () => {
    const { unit: _unit, ...withoutUnit } = reading();

    assert.throws(
      () => validateSensorReading(withoutUnit),
      failsAt("SensorReading"),
    );
  });

  it("applies boolean exclusive bounds", () => {
    assert.throws(
      () => validateSensorReading(reading({ ratio: 1 })),
      failsAt("SensorReading.ratio"),
    );
    assert.throws(
      () => validateSensorReading(reading({ labels: { a: 0 } })),
      failsAt("SensorReading.labels.a"),
    );
  });

  it("rejects an unknown integer enum value", () => {
    assert.throws(
      () => validateSensorReading(reading({ level: 3 })),
      failsAt("SensorReading.level"),
    );
  });
});

describe("recursive schemas", () => {
  it("checks every level", () => {
    validateNode({ name: "a", children: [{ name: "b", children: [] }] });
    assert.throws(
      () => validateNode({ name: "a", children: [{ children: [] }] }),
      failsAt("Node.children[0]"),
    );
  });
});

describe("receiving", () => {
  it("decodes the one message a frame fits", () => {
    const frame = decodeReceivedMessage('{"type":"run.started","sequence":1}');

    assert.equal(frame.name, MessageFrame);
    assert.equal(decodeReceivedMessage(JSON.stringify(reading())).name, MessageReading);
  });

  it("decodes a frame only strict decoding tells apart", () => {
    assert.equal(decodeReceivedMessage('{"name":"a"}').name, MessageTree);
  });
});

describe("channels", () => {
  it("fills parameters whose names are reserved words", () => {
    assert.equal(feedAddress("news", "u1"), "feed/news/u1");
  });
});
