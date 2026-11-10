# CI and tooling

Applies to `Dockerfile.dev`, `Makefile`, `scripts/` and `.github/workflows/pipeline.yml`.

- **CI1** Every check runs through a make target inside the `Dockerfile.dev` image. The CI runner has Docker, Make and Bash, not the Go or Node.js toolchain, so a new check gets a make target that uses `$(DEV_RUN)`, never a host tool.
- **CI2** Tool versions are pinned in one place each: the base image by tag and digest, `apk` packages by exact version, Go tools by version in `go.mod` or the `go install` line, semgrep by hash in `requirements-semgrep.txt`, TypeScript tools in `package.json` and `pnpm-lock.yaml`. Change a pin there, never in CI.
- **CI3** `scripts/sec.sh` collects every scanner's exit code and gates on all of them.
- **CI4** `pipeline.yml` runs on every push (`on: [push]`) with top-level `permissions: {}`; each job grants only what its called workflow needs. The `code` job is skipped on tags, which run `wait`, `release`, `badges` and `publish-to-clawhub` instead, against the `main` run of the tagged commit.
- **CI5** Validate workflow edits with `docker run --rm -v "$PWD:/repo" -w /repo rhysd/actionlint:1.7.7`. Leave `collaborators-only.yml`, `issue-pull.yml` and `mirror-and-archive.yml` alone unless the change is about them.
- **CI6** Shell scripts in `scripts/` use the shared `on_error` and `log` helpers.
