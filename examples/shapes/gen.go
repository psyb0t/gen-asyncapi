// Package shapes is generated from shapes.yaml, which covers the schema forms
// the chat example does not: unions without a discriminator keyword, primitive
// variants, typed maps, nullable fields, recursive schemas and numeric enums.
package shapes

//go:generate go tool gen-asyncapi --config=config.yaml shapes.yaml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=shapes.gen.ts shapes.yaml
