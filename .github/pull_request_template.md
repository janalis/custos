## What and why

<!-- What does this change, and what problem does it solve? Link the issue: Fixes #123 -->

## How it was tested

<!-- Fixtures added or changed, commands run, real projects checked. -->

## Checklist

- [ ] `make verify` passes (lint, tests, fixtures, 100 % coverage, clean-room scan)
- [ ] Rule behaviour changes: the spec (`specs/<ID>.md`) is updated, and fixtures in `testdata/rules/<ID>/` cover the change
- [ ] `make rules-doc` was run if a spec, a rule's fixtures or rule status changed
- [ ] User-visible change: a line under `## [Unreleased]` in `CHANGELOG.md`
- [ ] Clean-room: nothing in this PR is copied, translated or closely paraphrased from Php Inspections (EA Extended) code, messages, docs or fixtures, and rule code was written from the spec only
