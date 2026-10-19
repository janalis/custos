# Contributing to custos

Thanks for your interest! The full contributor guide is at
**<https://janalis.github.io/custos/contributing/>** (sources in
[`docs/contributing/`](docs/contributing/)).

## Reporting a problem

Open an issue and pick a template (*Wrong finding or fix* for rule problems,
*Bug report* for the CLI, the language server or installation). Include:

- a minimal PHP snippet;
- the target PHP version (`custos analyse --stats` prints it);
- what custos reports, or the fix it applies, and what you expected.

False positives and fixes that change behavior or produce invalid PHP are
treated as bugs, even when the upstream plugin behaves the same way.

## Clean-room rule (mandatory)

custos is an MIT-licensed, independent clean-room implementation. Never copy,
translate or closely paraphrase code, messages, descriptions, documentation or
fixtures from Php Inspections (EA Extended) (LGPL-2.1); rules are specified in
our own words, then implemented from the spec only. Read the
[clean-room process](https://janalis.github.io/custos/contributing/clean-room)
before touching a rule.

## Before opening a pull request

```sh
npm ci --prefix docs  # once: markdownlint (and the docs site)
make fmt              # gofumpt, goimports, markdownlint fixes
make rules-doc        # if you changed a spec, a rule's fixtures or rule status
make verify           # architecture, lint, tests, own fixtures, coverage, clean-room scan
```

- Every new statement is covered by tests (`make coverage`).
- Rule changes come with fixtures in
  `internal/inspection/rules/<lowercase-ID>/testdata/`.
- User-visible changes get a line under `## [Unreleased]` in `CHANGELOG.md`.

See [Make targets](https://janalis.github.io/custos/contributing/#make-targets)
for the full list.
