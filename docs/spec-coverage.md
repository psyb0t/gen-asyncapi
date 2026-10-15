# AsyncAPI 3 coverage

What the generator does with each part of an AsyncAPI 3.0 document. [examples/kitchensink](../examples/kitchensink/kitchensink.yaml) uses every row and validates against the official 3.0 JSON Schema.

## Document objects

| Object | Generated output |
|---|---|
| Info | Title and version in the file header. |
| `defaultContentType` | Content type of every message that sets none, shown in the message docs. |
| Server | `Server<Name>URL`, with variables at their defaults. |
| Server Variable | `Server<Name>URLTemplate` and `Server<Name>URLWith(...)`. |
| Channel | `<Channel>AddressTemplate`, plus `<Channel>Address(...)` when the channel has parameters. |
| Parameter | One string argument of `<Channel>Address`. Go keywords get a `Value` suffix. |
| Operation, `action: send` | One `Client` method per message. |
| Operation, `action: receive` | Its messages go into `DecodeReceivedMessage`, `Handler`, `Dispatch` and `Client.Run`. |
| Operation Reply | Reply messages go into the same receive path. |
| Message | A `Message<Name>` constant whose doc names the payload type, headers type and content type, and marks deprecation. |
| Message `headers` | A typed headers struct. |
| Message `correlationId` | `Message<Name>CorrelationIDLocation`, the runtime expression. |
| Message and Operation Traits | Merged with JSON Merge Patch in list order, keeping every property the target already sets. |
| Multi Format Schema Object | AsyncAPI, JSON Schema and OpenAPI schemas generate types. Avro, RAML and Protobuf payloads arrive as `json.RawMessage`. |
| Reference Object | Local and relative-file `$ref`s, chains included. |
| Components | Every schema and message gets generated. Servers, channels, operations, replies, reply addresses, parameters, correlation IDs, server variables and traits resolve through `$ref`. |
| Bindings | Merged with their traits and exposed by the spec loader. |

## Schema keywords

| Keyword | Handling |
|---|---|
| `type` | `string`, `integer`, `number`, `boolean`, `array`, `object`. `null` in a type list makes the field nullable. A list of several types becomes `json.RawMessage`. |
| `format` | `date-time` is `time.Time`, `int32` is `int32`, `float` is `float32`, `int64` and `double` are the defaults. Every other format appears in the field doc. |
| `properties`, `required` | Struct fields. Required ones must be present when decoding. |
| `additionalProperties` | `false` rejects undeclared properties on decode. A schema value with no `properties` makes a typed map. |
| `patternProperties` | One pattern makes a typed map whose keys must match it. |
| `propertyNames` | Its `pattern` is checked on map keys. |
| `minProperties`, `maxProperties` | Checked on decode for structs, in `Validate` for maps. |
| `dependentRequired`, `dependencies` (list form) | Checked on decode. |
| `items` | The element type. The tuple form makes `[]json.RawMessage`. |
| `minItems`, `maxItems`, `uniqueItems` | Checked. Uniqueness compares JSON encodings. |
| `minLength`, `maxLength`, `pattern` | Checked. Lengths count characters. |
| `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `multipleOf` | Checked. Exclusive bounds work as numbers and as draft 4 booleans. |
| `enum` | A named type with constants. A single value acts as `const`. |
| `const` | Filled in by `MarshalJSON`, checked on decode, and used to pick union variants and incoming messages. |
| `nullable` (OpenAPI 3.0) | Same as `null` in a type list. |
| `allOf` | One merged struct, constraints included. |
| `oneOf`, `anyOf` | A union type. The variant comes from `discriminator` (a string, or the OpenAPI object with `mapping`), from a property every variant pins with `const`, or from the one variant that decodes. |
| `contentMediaType` + `contentSchema` | `<Field>Value()` and `Set<Field>Value()` decode and encode the JSON carried in the string. |
| `title` | Names the type of an inline schema. |
| `description`, `default`, `examples`, `deprecated` | Doc comments. `deprecated` uses the `Deprecated:` form editors and linters recognize. |
| `$defs`, `$ref` | Resolved, recursive schemas included. |
