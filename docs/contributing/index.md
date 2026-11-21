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
documentation site and the Markdown linter). PHP 8.5 is optional: the syntax
tests compare custos' lexer and parser with it when it is installed, and
skip that comparison otherwise; CI installs it.

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
| `make examples` | Vet and build the example programs that use the public Go API (`examples/`, their own modules). |
| `make verify` | **Definition of done**, run by CI: `architecture`, `lint`, `test`, `examples`, `fixtures`, `coverage` and `cleanroom`. |
| `make architecture` | Check production dependencies, inspection catalogue completeness, rule package comments and fixture names. |
| `make build` | Build `bin/custos`. |
| `make test` | `go test ./...` |
| `make lint` | golangci-lint, actionlint, editorconfig-checker and markdownlint (see [Code style](#code-style)). |
| `make fmt` | Apply the Go formatters and markdownlint's automatic fixes. |
| `make fmt-check` / `make vet` | Shortcuts for the formatter diff and `govet`; both are part of `make lint`. |
| `make fixtures RULE=<ID>` | Run the own fixtures of every rule package, optionally for one rule. |
| `make conformance RULE=<ID>` | Compare against the upstream fixtures of a local Php Inspections checkout (`EA_PATH`, default `~/Sites/phpinspectionsea`). Skipped when absent. |
| `make cleanroom` | Scan the repository for forbidden upstream tokens and, when the local upstream checkout exists, for text found verbatim in it. See [Clean-room process](./clean-room#checking-against-upstream-without-copying-it). |
| `make extract` | Regenerate the EA-modeled rule facts and the local upstream case index from `EA_PATH`, then `make rules-doc`. |
| `make coverage` | 100% statement coverage gates (see [Ground rules](#ground-rules)). |
| `make clean` | Remove `bin/`, `dist/` and the docs build output. |
| `make rules-doc` | Regenerate the inspection catalogue (`internal/inspection/catalogue/catalogue.go`), the rule catalogue docs (`docs/rules/`, `docs/internals/rules.md`, `custos explain` texts) and the format settings reference (`docs/guide/format-settings.md`). |
| `make docs` / `make docs-dev` | Build or serve this site. |
| `make bench` / `make fuzz` | Benchmarks (`-benchmem`), with the language server's edit-latency budget enforced, and a short fuzz run of every fuzz target. |
| `make compare` / `make compare-doc` | Compare custos with other PHP tools on the projects of `bench/competitors.json` (needs PHP with `ast` and `pcntl`, Composer and git; `TOOL=` / `CORPUS=` filter), or only regenerate `docs/guide/benchmark.md` and the README comparison from `bench/results.json` and `bench/features.json`. The weekly `benchmark` workflow bumps versions, reruns it and opens a pull request. |
| `make corpus-baseline` / `make corpus-diff` | Record the findings of every rule on the pinned benchmark corpora, then compare a later build with the record: added and removed findings are listed by rule and written to `.cache/corpus-diff/`. `CORPORA=` filters. |
| `make typecov` | Measure the share of expressions whose inferred type is unknown on the benchmark corpora, by kind and cause (`.cache/typecov/run.summary`; `TYPECOV=` takes `root=srcdir` pairs separated by `;`). |
| `make fixcheck` | Apply every quick-fix on the projects in `CUSTOS_CORPUS` and require the result to still parse, to need no newer PHP version than the target, and each fix to remove its finding. `FIXCHECK_PHP=5.6` sets the target (default 8.5); with `FIXCHECK=php`, `PHP_BIN` names a php binary of that version to lint the results with. |
| `make formatcheck` | Format a copy of each benchmark corpus with each preset and run the style checker the preset follows on it (see [Formatter compliance](#formatter-compliance)). Needs PHP and Composer; `CORPORA=` and `PRESETS=` filter. Not part of `make verify`. |
| `make stubs` | Rebuild the embedded PHP builtin symbol index from phpstorm-stubs. |

AST node kinds are generated (by `tools/genkinds`): run
`go generate ./internal/php/syntax` after changing
`internal/php/syntax/ast.go`.

Some tests read optional environment variables:

| Variable | Effect |
|---|---|
| `CUSTOS_CORPUS` | Real PHP projects: `:`-separated roots for `make fixcheck` and the no-crash test of every rule; the parser oracle test reads a single root. |
| `CUSTOS_CORPUS_MAX` | Cap on the files the parser oracle test takes from the corpus. |
| `CUSTOS_FIXCHECK` | Set by `make fixcheck`; `php` also lints samples with `php -l`. |
| `CUSTOS_FIXCHECK_PHP`, `CUSTOS_PHP` | Set by `make fixcheck` from `FIXCHECK_PHP` and `PHP_BIN`: the target version and the php binary, which must be of that version when the target is set. Below 8.0 every fix result is linted, not a sample. |
| `CUSTOS_FORMAT_CORPUS` | A directory of real PHP code for the formatter's corpus test (`go test -run TestCorpus ./internal/format`): every preset must pass the safety check and be idempotent. `CUSTOS_FORMAT_PRESET` limits it to one preset, `CUSTOS_FORMAT_RISKY=1` allows risky settings. |
| `CUSTOS_BENCH_FILE` | A PHP file for the parser benchmark on real input. |
| `CUSTOS_PERF_TRACE` | `1` adds raw timing for each edit, parse/analysis phases, CPU time, GC count and pause deltas to `TestEditLatency`; GC snapshots and trace logging occur outside the measured edit. |
| `CUSTOS_PERF` | `1` (set by `make bench`) turns logged performance budgets into failures. |
| `CUSTOS_TYPECOV`, `CUSTOS_TYPECOV_OUT` | `root=srcdir;…` roots to measure expressions with unknown inferred types, and where to write the report. |

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
- Prose uses US English (`behavior`, `analyze`); the `custos analyse` command
  and identifiers from PHP and the LSP protocol keep their own spelling.
- Terms: "custos" is lowercase, even at the start of a sentence; user docs
  say "rule" (not "inspection") and "quick-fix"; the upstream project is
  "Php Inspections (EA Extended)", and rules derived from it are
  "EA-modeled".
- Generated files (`docs/rules/`, `docs/internals/rules.md`, the `explain`
  texts) are fixed in their generator or their spec, then `make rules-doc`.
- Editors: enable EditorConfig support, and gofumpt in gopls
  (`"gopls": {"formatting.gofumpt": true}`).

## Adding a format setting

The formatter (`internal/format`) has its own workflow; formatting is not a
rule, so there is no spec or clean-room step.

1. Add the setting to the registry in `internal/format/settings.go`: a field
   of `Settings`, then an entry with its ID, group, kind, allowed values,
   description (user-facing, our own words) and `Risky: true` when it
   changes more than layout. Give it its neutral value (usually `keep` or
   `false`) in `keepAll` in `presets.go`, and the value each style guide asks
   for in the presets (in `risky` for a risky setting).
2. Add the message of its changes to `messages.go`, short and imperative.
3. Implement it in the engine, then add a golden case
   `internal/format/testdata/cases/<name>.php` whose first line names what
   to run: `<?php // presets: per symfony; set: line_length=60, wrap=long; risky`
   (list values separated by `|`). Run `go test ./internal/format -update`
   to write `<name>.<preset>.php`, and review every output by hand: the
   golden test then guards them, and checks that the changes replay to the
   output and that formatting it again changes nothing.
4. Fuzz it (`go test -fuzz FuzzFormat ./internal/format`) and run it over
   real code (`CUSTOS_FORMAT_CORPUS=/path/to/project go test -run TestCorpus ./internal/format`),
   then compare the presets with their style checkers (`make formatcheck`).
5. If another tool's configuration maps onto it unambiguously, read it in
   `internal/project/config/formatimport.go`.
6. Run `make rules-doc` to regenerate
   [Format settings](/guide/format-settings), and record the setting in
   `CHANGELOG.md`.

## Formatter compliance

`make formatcheck` checks the presets against the tools that define the
style guides they follow. For each benchmark corpus (`CORPORA=`, default
all), it copies the corpus's source into the system temporary directory,
formats a copy with each preset (`PRESETS=`, default all), and runs the
preset's reference checker on the copy before and after formatting:

| Preset | Checker |
|---|---|
| `psr12` | PHP_CodeSniffer, `--standard=PSR12` |
| `per` | PHP-CS-Fixer, `--rules=@PER-CS` |
| `symfony` | PHP-CS-Fixer, `--rules=@Symfony` |
| `laravel` | Laravel Pint, `--preset laravel` |
| `drupal` | PHP_CodeSniffer, `--standard=Drupal` (drupal/coder) |

The checkers are the benchmark's pinned packages, installed by the target
under `.cache/bench`; drupal/coder is pinned in `tools/formatcheck` and
installed under `.cache/formatcheck`. The corpora are never modified.

The command prints one row per corpus and preset:

- `before` and `after`: the checker's findings on the copy before and after
  formatting (for PHP-CS-Fixer and Pint, one per fixer and file);
- `divergent`: the findings after formatting that
  `internal/format/testdata/divergences.json` explains: reference rules
  custos deliberately does not satisfy, such as docblock content, risky
  rewrites it only applies on request, or a line that cannot be shortened;
- `remaining`: the findings left that are not divergences, listed below
  the table by rule with examples (`file:line: message`; PHP-CS-Fixer gives
  the first change of the file, as it does not say which fixer made which
  change).

PHP-CS-Fixer and Pint name every fixer that touched a file, including the
fixers that only tidy up after a divergent one (the blank lines left by a
removed import, say), and Pint names some whose changes cancel out. For
their presets, `remaining` therefore comes from a second run with the
divergent fixers turned off (patterns such as `phpdoc_*` are expanded to
the installed PHP-CS-Fixer's rules): its findings for PHP-CS-Fixer, and for
Pint the files whose code, docblocks aside, its fix still changes. There,
`after` minus `divergent` minus `remaining` counts those follow-on reports.

The command exits 1 when a finding remains. Each one is either a formatter
bug, or a deliberate choice to record with its reason in
`divergences.json` (a rule name or a pattern such as `phpdoc_*`). `-json`
prints every count and remaining finding for scripts:
`go run ./tools/formatcheck -json -corpora guzzle`.

## Ground rules

- **Coverage is 100%.** Every statement of `internal/inspection/rules` must be
  covered by the rules' own fixtures and tests and the catalogue tests, and
  every statement of `cmd/`, `internal/` and `tools/` by the whole suite. Remove unreachable code rather
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
  status changes, then `make verify`. The status table's conformance column
  comes from `.cache/ea/conformance.json`, which only a full
  `make conformance` (without `RULE=`) rewrites; filtered runs keep the
  previous status of the other rules.
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
