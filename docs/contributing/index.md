# Contributing

Thanks for helping! Bug reports with a minimal PHP snippet are the most
useful contribution: a false positive, a missed case, or a fix that changes
behavior or produces invalid PHP. Please
[open an issue](https://github.com/janalis/custos/issues/new/choose) with the
snippet, the target PHP version and the output of `custos analyse`. The
*Wrong finding or fix* template asks for all of it. Pull requests get a
checklist that matches the ground rules below.

::: warning Clean-room rule
custos is a clean-room implementation of rules modeled on Php Inspections
(EA Extended). Never copy, translate or closely paraphrase its code or text,
and never open its sources while implementing a rule. Read the
[clean-room process](./clean-room) before you touch a rule.
:::

## Setup

You need Go 1.27 or later, `make`, `git` and Node 20+ (for the
documentation site and the Markdown linter).

```sh
git clone https://github.com/janalis/custos
cd custos
npm ci --prefix docs  # markdownlint, VitePress
make build            # → bin/custos
make test
```

Binaries go to `bin/`. Do not run a bare `go build ./cmd/...`, which drops a
binary in the working directory; use `go build ./...` to only check that the
code compiles.

## Make targets

| Target | What it does |
|---|---|
| `make verify` | **Definition of done**, run by CI: `architecture`, `lint`, `test`, `fixtures`, `coverage` and `cleanroom`. |
| `make architecture` | Check production dependencies and inspection catalogue completeness. |
| `make build` | Build `bin/custos`. |
| `make test` | `go test ./...` |
| `make lint` | golangci-lint, actionlint, editorconfig-checker and markdownlint (see [Code style](#code-style)). |
| `make fmt` | Apply the Go formatters and markdownlint's automatic fixes. |
| `make fmt-check` / `make vet` | Shortcuts for the formatter diff and `govet`; both are part of `make lint`. |
| `make fixtures RULE=<ID>` | Run the own fixtures of every rule package, optionally for one rule. |
| `make conformance RULE=<ID>` | Compare against the upstream fixtures of a local Php Inspections checkout (`EA_PATH`, default `~/Sites/phpinspectionsea`). Skipped when absent. |
| `make cleanroom` | Scan the repository for text found verbatim in the local upstream checkout. Skipped when absent. |
| `make extract` | Regenerate the EA-modeled rule facts and the local upstream case index from `EA_PATH`, then `make rules-doc`. |
| `make coverage` | 100% statement coverage gates (see [Ground rules](#ground-rules)). |
| `make rules-doc` | Regenerate the rule catalogue docs: `docs/rules/`, `docs/internals/rules.md`, `custos explain` texts. |
| `make docs` / `make docs-dev` | Build or serve this site. |
| `make bench` / `make fuzz` | Benchmarks (`-benchmem`) and a short fuzz run of every fuzz target. |
| `make fixcheck` | Apply every quick-fix on the projects in `CUSTOS_CORPUS` and require the result to still parse. |
| `make stubs` | Rebuild the embedded PHP builtin symbol index from phpstorm-stubs. |

AST node kinds are generated: run `go generate ./internal/php/syntax` after
changing `internal/php/syntax/ast.go`.

## Code style

`make lint` (part of `make verify`) checks every file; `make fmt` fixes most
findings for you. Run it before committing. Tool versions are pinned in the
`Makefile` and the Go ones are built with your Go toolchain on first use; set
a tool's variable empty to skip it offline (`make lint MDLINT=`).

| Files | Tool | Rules |
|---|---|---|
| Go | [golangci-lint](https://golangci-lint.run) | `.golangci.yml`: `gofumpt` and `goimports` formatting (standard library, then `custos/...` imports), `govet`, `staticcheck`, `errcheck`, `errorlint`, `revive`, `gocritic`, `misspell` and a few more. |
| GitHub workflows | [actionlint](https://github.com/rhysd/actionlint) | Workflow syntax and expressions; `run:` scripts through `shellcheck` when it is installed. |
| All text files | [editorconfig-checker](https://github.com/editorconfig-checker/editorconfig-checker) | `.editorconfig`: UTF-8, LF, final newline, no trailing whitespace, tabs in Go and the `Makefile`, spaces elsewhere. |
| Markdown | [markdownlint](https://github.com/DavidAnson/markdownlint) | `.markdownlint-cli2.jsonc`: the default rules, without line length, inline HTML and code-span spacing checks. |

- Fix findings rather than silencing them. When a finding is wrong, silence
  that one line with `//nolint:<linter> // <reason>`; a bare or unexplained
  `//nolint` is itself an error.
- PHP fixtures and other `testdata/` files are byte-exact test inputs: no
  formatter or checker touches them.
- Prose uses US English (`behavior`); the `custos analyse` command and
  identifiers from PHP and the LSP protocol keep their own spelling.
- Generated files (`docs/rules/`, `docs/internals/rules.md`, the `explain`
  texts) are fixed in their generator or their spec, then `make rules-doc`.
- Editors: enable EditorConfig support, and gofumpt in gopls
  (`"gopls": {"formatting.gofumpt": true}`).

## Ground rules

- **Coverage is 100%.** Every statement of `internal/inspection/rules` must be
  covered by the rules' own fixtures and tests, and every statement of `cmd/`,
  `internal/` and `tools/` by the whole suite. Remove unreachable code rather
  than exempting it.
- **Performance is a feature.** Add a benchmark for any hot path you touch,
  and watch allocations (`-benchmem`). Timing limits in tests use
  `testbudget.Of(d)`, which scales with machine load, and time compute-bound
  work with `testbudget.CPU()` (process CPU time) rather than the wall clock.
- **Correctness beats upstream fidelity.** When the upstream behavior is
  wrong (false positives, fixes that change semantics or produce invalid PHP),
  diverge and document it (see [Divergences](./clean-room#divergences)).
- **Standard library first**: no new dependency without a clear reason.
- **Messages** are short, imperative and in our own words.
- Supported PHP versions are 5.3 to 8.5: gate version-dependent rules on the
  target version.
- **Order of checks**: `make rules-doc` when a spec, a rule's fixtures or rule
  status changes, then `make verify`.
- **Changelog**: record every user-visible change under `## [Unreleased]` in
  `CHANGELOG.md`, and check the staged diff against it before committing.
- **Documentation** describes current behavior and durable practices. Keep
  session plans, progress logs, corpus comparisons and timing reports in the
  ignored `.cache/` directory.

## Where to go next

- [Architecture](./architecture): package ownership and how a file goes from
  bytes to findings.
- [Adding a rule](./adding-a-rule): the spec → implementation → fixtures
  workflow.
- [Clean-room process](./clean-room): what may be reused from upstream.
- [Fixtures](./fixtures): the test file format.
- [Releasing](./releasing): cutting a version.

The generated
[rule status](https://github.com/janalis/custos/blob/main/docs/internals/rules.md)
tracks specifications and upstream conformance.
