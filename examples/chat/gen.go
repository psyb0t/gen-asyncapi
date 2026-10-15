// Package chat is generated from chat.yaml, a spec that uses every feature
// gen-asyncapi supports. Its tests check the generated code end to end.
package chat

//go:generate go tool gen-asyncapi --config=config.yaml chat.yaml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=chat.gen.ts chat.yaml
