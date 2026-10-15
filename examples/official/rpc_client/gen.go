// Package rpc_client is generated from the official AsyncAPI example
// rpc-client-asyncapi.yml.
package rpc_client

//go:generate go tool gen-asyncapi --config=config.yaml spec.yml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=rpc_client.gen.ts spec.yml
