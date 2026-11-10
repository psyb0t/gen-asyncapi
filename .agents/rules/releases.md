# Releases and docs

Refines `~/.claude/rule-details/versioning-a-project.md` and `~/.claude/rule-details/writing-docs.md`.

- **REL1** The git tag is the version. Before 1.0 a breaking change to the CLI, the config, the generated API or `asyncapirt` bumps the minor version and says so in the CHANGELOG with the migration; everything else is a patch.
- **REL2** Every release prepends a `## vX.Y.Z, YYYY-MM-DD` section to `CHANGELOG.md`. The annotated tag message is `vX.Y.Z`, a blank line, then that section.
- **REL3** `.agents/.codex-plugin/plugin.json` carries `version` and moves to the release version in the same commit. `.agents/.claude-plugin/plugin.json` has no `version`.
- **REL4** A change to the CLI, the config or the generated API updates the matching `docs/` pages (and `docs/spec-coverage.md` for spec handling), the README when its quick start or install lines change, and the agent skill in `.agents/skills/gen-asyncapi/` in the same commit.
- **REL5** Commits, tags and pushes go through `git-update.sh` only. A tag push waits for the `main` run of the tagged commit, cuts the GitHub release from the annotated tag message, then publishes the skill to ClawHub through the `publish-to-clawhub` job in `pipeline.yml`. Push the commit to `main` and let its run pass before tagging it.
