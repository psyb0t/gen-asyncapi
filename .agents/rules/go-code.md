# Go code

Refines `~/.claude/rule-details/go/wrapping-errors.md`, `~/.claude/rule-details/go/logging.md` and `~/.claude/rule-details/coding-style.md`.

- **GO1** Every new failure wraps a `ctxerrors/commerr` sentinel with `ctxerrors.Wrap` or `ctxerrors.Wrapf`; errors from a call are wrapped on the next line. No `fmt.Errorf`, no `errors.New` outside a sentinel declaration. Runtime sentinels the generated code exposes (`asyncapirt.ErrUnknownMessage` and friends) live in `asyncapirt/errors.go`.
- **GO2** Only the CLI and the generated `UnimplementedHandler` log, through `ctxscope.GetLogger(ctx)` with structured `lowercase_snake` fields. `internal/codegen` and `asyncapirt` return errors and never log.
- **GO3** Generated output is shaped by the templates and the renderers, never by post-editing the file. Go output goes through `imports.Process`; TypeScript output is indented by `tsIndent`.
- **GO4** Every name the generator emits is claimed through a `nameSet`, so schema types never collide with each other or with the reserved names in `reservedNames`.
- **GO5** `make lint` passes with zero issues. Lines are at most 80 columns, the `.golangci.yml` limit; split calls one argument per line rather than adding `//nolint`.
- **GO6** Finite string values get a named type with constants, as `Language` does.
