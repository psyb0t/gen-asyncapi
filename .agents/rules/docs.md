# Docs

Refines `~/.claude/rule-details/writing-docs.md`.

- **DOC1** The documentation site is `docs/`, published by GitHub Pages from `main` at https://psyb0t.github.io/gen-asyncapi/ with the Just the Docs theme (`remote_theme` pinned in `docs/_config.yml`, colors and sizes in `docs/_sass/`, the shared visitor tracker in `docs/_includes/head_custom.html`). Keep `docs/` flat: a file name is its URL.
- **DOC2** Every page has front matter with a unique `title` and a `description` of 50 to 170 characters, and exactly one `#` heading. A child page sets `parent` to the parent's exact title, as `spec-coverage.md` does with "What gets generated".
- **DOC3** Links between pages use the relative `.md` path. A link to a file outside `docs/` uses the full `https://github.com/psyb0t/gen-asyncapi/blob/main/<path>` URL, or `tree/main/<path>` for a directory. Wrap a code block that contains `{{` or `{%` in `<!-- {% raw %} -->` and `<!-- {% endraw %} -->`.
- **DOC4** The README is the front door: what the generator is, install, a quick start, and a Documentation section with full site URLs. Reference material goes in `docs/`, and each fact has one home. README links to repository files use full GitHub URLs, because the README is also shown on pkg.go.dev.
- **DOC5** After a docs change, run `make docs` and the site checker against `.testing/docs-site`, and fix every error before handing off.
- **DOC6** Error messages quoted in `docs/troubleshooting.md` are copied from a real run of the command. Update them when the wording in the code changes.
- **DOC7** Prose: no em dashes, "directory" never "folder", no AI vocabulary, sentence case headings, one paragraph per source line. Do not mention local scratch directories in committed text.
