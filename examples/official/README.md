# Official AsyncAPI examples

Each `spec.yml` here is an unmodified copy of an example from the [AsyncAPI specification repository](https://github.com/asyncapi/spec/tree/1dd65fd2c1ed13f06365c1e870c61cdc82d8a981/examples), taken at commit `1dd65fd2c1ed13f06365c1e870c61cdc82d8a981`. The `gen.go` doc comment in each directory names the original file.

The specs are licensed under the Apache License 2.0 by the AsyncAPI Initiative. See [LICENSE](LICENSE) and [NOTICE](NOTICE). The rest of this repository, including the generator and the `config.yaml`, `gen.go` and generated `*.gen.go` files around the specs, is MIT licensed.

`make generate` regenerates every package from its spec. CI builds them all and fails when the committed output no longer matches what the generator produces.
