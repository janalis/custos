# Contributing to custos

Thanks for your interest! The full contributor guide is at
**https://janalis.github.io/custos/contributing/** (sources in
[`docs/contributing/`](docs/contributing/)).

## Reporting a problem

Open an issue and pick a template (*Wrong finding or fix* for rule problems,
*Bug report* for the CLI, the language server or installation). Include:

- a minimal PHP snippet;
- the target PHP version (`custos analyse --stats` prints it);
- what custos reports, or the fix it applies, and what you expected.

False positives and fixes that change behaviour or produce invalid PHP are
treated as bugs, even when the upstream plugin behaves the same way.

## Clean-room rule (mandatory)

custos is an MIT-licensed, independent clean-room implementation of a rule
catalogue modelled on Php Inspections (EA Extended), which is LGPL-2.1.

- Never copy, translate or closely paraphrase its code, messages,
  descriptions, documentation or test fixtures.
- Rule behaviour is specified in our own words in `specs/<ID>.md`, then
  implemented from the spec only, without opening the upstream sources.

Details: [Clean-room process](https://janalis.github.io/custos/contributing/clean-room).

## Before opening a pull request

```sh
make verify         # lint, tests, own fixtures, 100 % coverage, clean-room scan
make rules-doc      # if you changed a spec or a rule's fixtures
```

- Every new statement is covered by tests (`make coverage`).
- Rule changes come with fixtures in `testdata/rules/<ID>/`.
- User-visible changes get a line under `## [Unreleased]` in `CHANGELOG.md`.
