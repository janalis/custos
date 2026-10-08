# Contributing

Thanks for helping! Bug reports with a minimal PHP snippet are the most
useful contribution: a false positive, a missed case, or a fix that changes
behaviour or produces invalid PHP. Please
[open an issue](https://github.com/janalis/custos/issues/new/choose) with the
snippet, the target PHP version and the output of `custos analyse`. The
*Wrong finding or fix* template asks for all of it. Pull requests get a
checklist that matches the ground rules below.

::: warning Clean-room rule
custos is a clean-room implementation. Never copy, translate or closely
paraphrase code, messages, descriptions, documentation or test fixtures from
Php Inspections (EA Extended), and do not open its sources while you
implement a rule. Read the [clean-room process](./clean-room) before you
touch a rule.
:::

## Setup

You need Go 1.27 or later, `make` and `git`. Node 20+ is only needed for the
documentation site.

```sh
git clone https://github.com/janalis/custos
cd custos
make build          # → bin/custos
make test
```

Binaries go to `bin/`. Do not run a bare `go build ./cmd/...`, which drops a
binary in the working directory; use `go build ./...` to only check that the
code compiles.

## Make targets

| Target | What it does |
|---|---|
| `make verify` | **Definition of done**: lint, tests, own fixtures, 100 % coverage gates and the clean-room scan. CI runs it. |
| `make build` | Build `bin/custos`. |
| `make test` | `go test ./...` |
| `make lint` | `go vet`, `gofmt` check and a pinned `staticcheck` (`STATICCHECK=` skips it offline). |
| `make fixtures RULE=<ID>` | Run the own fixtures in `testdata/rules/`, optionally for one rule. |
| `make conformance RULE=<ID>` | Compare against the upstream fixtures of a local Php Inspections checkout (`EA_PATH`, default `~/Sites/phpinspectionsea`). Skipped when absent. |
| `make coverage` | 100 % statement coverage gates (see below). |
| `make rules-doc` | Regenerate the rule catalogue docs: `docs/rules/`, `docs/internals/rules.md`, `custos explain` texts. |
| `make docs` / `make docs-dev` | Build or serve this site. |
| `make bench` / `make fuzz` | Benchmarks (`-benchmem`) and a short fuzz run of every fuzz target. |
| `make fixcheck` | Apply every quick-fix on the projects in `CUSTOS_CORPUS` and require the result to still parse. |
| `make stubs` | Rebuild the embedded PHP builtin symbol index from phpstorm-stubs. |

## Ground rules

- **Coverage is 100 %.** Every statement of `internal/rules` must be covered
  by the rules' own fixtures and tests, and every statement of `cmd/`,
  `internal/` and `tools/` by the whole suite. Remove unreachable code rather
  than exempting it.
- **Performance is a feature.** Add a benchmark for any hot path you touch,
  and watch allocations (`-benchmem`). Timing limits in tests use
  `testbudget.Of(d)`, which scales with machine load.
- **Correctness beats upstream fidelity.** When the upstream behaviour is
  wrong (false positives, fixes that change semantics or produce invalid PHP),
  diverge. Document it in the spec's *Divergences* section and in
  `testdata/ea-divergences.json`.
- **Standard library first**: no new dependency without a clear reason.
- **Messages** are short, imperative and in our own words.
- Supported PHP versions are 5.3 to 8.5: gate version-dependent rules on the
  target version.
- Write user-visible changes under `## [Unreleased]` in `CHANGELOG.md` as you
  go.

## Where to go next

- [Architecture](./architecture): how a file goes from bytes to findings.
- [Adding a rule](./adding-a-rule): the spec → implementation → fixtures
  workflow.
- [Fixtures](./fixtures): the test file format.
- [Releasing](./releasing): cutting a version.

The working notes of the port (plan, design decisions, divergence candidates,
per-rule status) are in
[`docs/internals/`](https://github.com/janalis/custos/tree/main/docs/internals).
