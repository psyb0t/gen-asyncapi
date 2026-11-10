---
title: "Development"
description: "Change gen-asyncapi and prove it: the Docker dev image, every make target, example regeneration, the CI pipeline, the docs build and releases."
nav_order: 11
---

# Development

Everything runs in the [`Dockerfile.dev`](https://github.com/psyb0t/gen-asyncapi/blob/main/Dockerfile.dev) image. You need Docker and Make. The image pins Go 1.26.10, Node.js, pnpm, the TypeScript compiler, shellcheck, shfmt, govulncheck, gitleaks and semgrep, so a local run and a CI run use the same tools. Each target builds the image first, then runs as your user with the checkout mounted at `/work`.

## Make targets

```sh
make help           # list targets
make generate       # regenerate every example, Go and TypeScript
make lint           # shellcheck, shfmt, go fix, golangci-lint
make test           # go test -race, then the TypeScript checks
make test-coverage  # go test -race with a 90% coverage gate, then make test-ts
make test-ts        # tsc on the generated TypeScript, then its node:test suites
make sec            # gitleaks, govulncheck, pnpm audit, semgrep into sec.sarif
make dep            # go mod tidy and go mod vendor
make docs           # build this site into .testing/docs-site
```

[`pre-commit.sh`](https://github.com/psyb0t/gen-asyncapi/blob/main/pre-commit.sh) runs `make lint && make test-coverage`. The scripts behind the targets live in [`scripts/`](https://github.com/psyb0t/gen-asyncapi/tree/main/scripts).

The coverage gate counts the hand-written packages only, `asyncapirt`, `internal` and `cmd`, and writes the percentage to `coverage-percent.txt` for the badge. Dependencies are vendored in `vendor/`, and `make dep` refreshes it.

## Changing the generator

- Never hand-edit a `*.gen.go` or `*.gen.ts` file. Change the generator in [`internal/codegen`](https://github.com/psyb0t/gen-asyncapi/tree/main/internal/codegen), then run `make generate`. `TestGeneratedExamplesAreCurrent` fails until the committed examples match.
- Both languages come from one model, so a builder change shows up in both outputs. Check both.
- New generated behavior gets a test against an example's committed output: Go in `examples/<name>/<name>_test.go`, TypeScript in `examples/<name>/<name>.test.ts`.
- The generated TypeScript must pass the strict [`tsconfig.json`](https://github.com/psyb0t/gen-asyncapi/blob/main/tsconfig.json), which checks it without Node.js types. The tests are checked with [`tsconfig.test.json`](https://github.com/psyb0t/gen-asyncapi/blob/main/tsconfig.test.json), which adds them.

## TypeScript tools

The TypeScript compiler and Node.js types are pnpm dev dependencies, pinned in [`package.json`](https://github.com/psyb0t/gen-asyncapi/blob/main/package.json) and installed into the dev image from `pnpm-lock.yaml`. [`pnpm-workspace.yaml`](https://github.com/psyb0t/gen-asyncapi/blob/main/pnpm-workspace.yaml) holds a seven-day release-age gate, and no install runs package scripts. These targets only rewrite `package.json` and the lockfile; the next target that builds the dev image installs the result:

```sh
make pkg-add PKG=name@version   # add an exact dev dependency
make pkg-update PKG=name        # update one to its newest gated release
make pkg-upgrade                # update all of them
make pkg-remove PKG=name        # remove one
make pkg-lock                   # refresh the lockfile under the current pins
```

## CI

[`.github/workflows/pipeline.yml`](https://github.com/psyb0t/gen-asyncapi/blob/main/.github/workflows/pipeline.yml) calls the shared workflows in [psyb0t/reusable-github-workflows](https://github.com/psyb0t/reusable-github-workflows):

- Every branch push runs `make lint`, `make test-coverage`, `make sec` and `make generate`. The last one fails when the regenerated examples differ from the committed ones. The security scan's SARIF goes to the repository's Security tab.
- A push to `main` then refreshes the coverage, version and license badges on the `badges` branch.
- A `v*` tag does not rerun the checks. It waits for the `main` run of the same commit to pass, cuts the GitHub release from the annotated tag message, refreshes the badges with that run's coverage, and publishes the agent skill to ClawHub.

## Docs

This site is `docs/`, published by GitHub Pages from `main` with the [Just the Docs](https://just-the-docs.com) theme. Every page needs front matter with a unique `title` and a `description` of 50 to 170 characters. Links between pages use the relative `.md` path, and a link to a file outside `docs/` uses its full GitHub URL. `make docs` builds the site with the same image GitHub Pages uses.

## Releases

The git tag is the version. A release adds a `## vX.Y.Z, YYYY-MM-DD` section to the top of [`CHANGELOG.md`](https://github.com/psyb0t/gen-asyncapi/blob/main/CHANGELOG.md), moves the `version` in `.agents/.codex-plugin/plugin.json` to match, and is tagged with an annotated `vX.Y.Z` tag whose message is that section. Before 1.0, a breaking change to the command line, the config, the generated API or `asyncapirt` bumps the minor version and comes with a migration note.
