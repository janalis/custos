## What and why

<!-- What does this change, and what problem does it solve? Link the issue: Fixes #123 -->

## How it was tested

<!-- Fixtures added or changed, commands run, real projects checked. -->

## Checklist

- [ ] `make rules-doc` was run if a spec, a rule's fixtures or rule status changed
- [ ] `make verify` passes (architecture, lint, tests, fixtures, 100% coverage, clean-room scan)
- [ ] Rule behavior changes: the spec (`specs/<ID>.md`) is updated, and fixtures in `internal/inspection/rules/<lowercase-ID>/testdata/` cover the change
- [ ] User-visible change: a line under `## [Unreleased]` in `CHANGELOG.md`
- [ ] Clean-room: nothing is copied, translated or closely paraphrased from Php Inspections (EA Extended), and rule code was written from the spec only ([clean-room process](https://janalis.github.io/custos/contributing/clean-room))
