# Agent surface

Applies to `.agents/skills/`, `.agents/.claude-plugin/` and `.agents/.codex-plugin/`.

- **AGT1** The skill (`.agents/skills/gen-asyncapi/SKILL.md` and `references/setup.md`) is for people using the generator, not for changing it. Every flag, config key, generated name and error it mentions must match the code and the docs site; fix both when either changes.
- **AGT2** The `permissions` block in `SKILL.md` describes what the skill reads, writes, runs and fetches. Keep it true when the skill's instructions change.
- **AGT3** `pipeline.yml` publishes exactly the directories named in `skill_names` (`["gen-asyncapi"]`) and `plugin_names` (`[]`). Add a new skill or plugin directory to those arrays.
- **AGT4** `.agents/.codex-plugin/plugin.json` carries the release `version`; `.agents/.claude-plugin/plugin.json` has none (see `releases.md`). Keep the two manifests' `description` and `keywords` equal.
