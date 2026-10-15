// Package rpc_server is generated from the official AsyncAPI example
// rpc-server-asyncapi.yml.
package rpc_server

//go:generate go tool gen-asyncapi --config=config.yaml spec.yml
//go:generate go tool gen-asyncapi --config=config.yaml -language=typescript -o=rpc_server.gen.ts spec.yml
