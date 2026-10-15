# Testing

Refines `~/.claude/rule-details/go/writing-tests.md`.

- **TST1** Go table tests use `testCases` and `tc`, call `t.Parallel()` in the test and every subtest, and assert errors with `require.ErrorIs` against the sentinel.
- **TST2** A generator change ships with `make generate` output for every example in both languages. `TestGeneratedExamplesAreCurrent` enforces it.
- **TST3** New generated behavior gets a test against an example's committed output: Go tests in `examples/<name>/<name>_test.go`, TypeScript tests in `examples/<name>/<name>.test.ts` on `node:test`.
- **TST4** TypeScript tests import the generated file with its `.ts` extension and must pass `tsc -p tsconfig.test.json`. The generated files must pass `tsc -p tsconfig.json`, which has no Node.js types.
- **TST5** `make test-coverage` keeps Go statement coverage of `asyncapirt`, `internal` and `cmd` at 90% or above, with `-race`, and runs the TypeScript suite after it.
