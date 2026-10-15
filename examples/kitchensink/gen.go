// Package kitchensink is generated from kitchensink.yaml, which uses every
// AsyncAPI 3.0 object and every schema keyword the generator reads.
package kitchensink

//go:generate go tool gen-asyncapi --config=config.yaml kitchensink.yaml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=kitchensink.gen.ts kitchensink.yaml
