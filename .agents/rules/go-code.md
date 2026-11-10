# Go code

- **GO1** Runtime sentinels the generated code exposes (`asyncapirt.ErrUnknownMessage` and friends) live in `asyncapirt/errors.go`.
- **GO2** Only the CLI and the generated `UnimplementedHandler` log. `internal/codegen` and `asyncapirt` return errors and never log.
- **GO3** Generated output is shaped by the templates and the renderers, never by post-editing the file. Go output goes through `imports.Process`; TypeScript output is indented by `tsIndent`.
- **GO4** Every name the generator emits is claimed through a `nameSet`, so schema types never collide with each other or with the reserved names in `reservedNames`.
- **GO5** Lines are at most 80 columns, the `.golangci.yml` limit.
