# custos

Fast PHP inspector + fixer written in Go. Its rule catalogue includes native
inspections and rules modeled on Php Inspections (EA Extended); it is an
**independent clean-room implementation** released under MIT. Also runs as an
LSP server (`custos lsp`) for editors.

Contributor docs (canonical, read before changing an area):
`docs/contributing/` — `index.md` (make targets, code style, ground rules),
`architecture.md` (package ownership, contracts, helper placement),
`adding-a-rule.md`, `clean-room.md`, `fixtures.md`, `releasing.md`.
Per-root `internal/*/AGENTS.md` summarize ownership. Per-rule status:
`docs/internals/rules.md`.

Documentation describes current behavior and durable contributor practices.
Keep agent plans, progress logs and local verification reports in ignored
`.cache/`; never commit them as documentation. Keep catalogue totals in the
generated rule reference and status table, not in marketing or usage prose.

## Clean-room rules (MANDATORY)

The upstream project (`~/Sites/phpinspectionsea`, LGPL-2.1) must never leak
into this repo. Full process: `docs/contributing/clean-room.md`.

- NEVER copy, translate or closely paraphrase EA Java code, messages,
  descriptions, docs or fixtures. No EA file is ever copied.
- Allowed facts only: rule IDs (EA short name minus `Inspection`), groups,
  default severity/enabled flags, option names/defaults, PHP-version
  thresholds, all captured in `internal/inspection/meta/rules.json` by
  `make extract`.
- Two separate agents/sessions per rule:
  1. `spec-rule` skill: read EA, write `specs/<ID>.md` in our own words with
     new examples.
  2. `implement-rule` skill: implement **from the spec only**; never open EA
     Java sources or EA fixtures while implementing.
- EA fixtures are only *read at test time* from the local checkout by
  `make conformance` (rule/range/severity + fix result compared; EA message
  text is never compared or stored).

## Layout

Canonical ownership table: `docs/contributing/architecture.md`.

```text
cmd/custos/                         executable entry point
internal/php/                       syntax, version, phpdoc
internal/semantic/                  names, types, index, infer, flow, stubs
internal/inspection/                analysis, meta, catalogue, phpunit, {ast,flow,semantic}query
internal/inspection/rules/<id>/     rule.go (New), implementation, tests, testdata/, optional NOTES.md
internal/diagnostic/ internal/fixing/  findings and edits; fix application
internal/project/                   discovery, indexing, analysis, prepared fixes; config, baseline
internal/output/ internal/cli/ internal/editor/lsp/  adapters
internal/platform/safeio/ internal/testing/  bounded IO; conformance harness, test budgets
tools/                              generators, architecture, coverage, clean-room checks
specs/                              clean-room behavioral spec per rule ID
docs/                               VitePress guide, contributing docs, generated rules
```

## Commands

```text
make build          # bin/custos
make test           # go test ./...
make lint           # golangci-lint + actionlint + editorconfig-checker + markdownlint (pinned; GOLANGCI=/ACTIONLINT=/ECCHECK=/MDLINT= to skip)
make vet / fmt-check # govet only / formatter diff (both part of lint)
make fmt            # gofumpt + goimports + markdownlint --fix (run before committing)
make extract        # regenerate rule facts + local EA index (EA_PATH=~/Sites/phpinspectionsea), then rules-doc
make rules-doc      # regenerate docs/rules/ (site), docs/internals/rules.md, explain texts
make fixtures       # own fixtures (CI gate)            RULE=<ID> to filter
make conformance    # EA fixtures from local checkout   RULE=<ID> to filter
make bench / fuzz
make cleanroom      # scan repo for verbatim EA text (local)
make docs / docs-dev # build / serve the docs site (Node 20+, docs/package.json)
make coverage       # 100% gates: internal/inspection/rules from own fixtures + rule tests; cmd/ + internal/ + tools/ from the whole suite
make architecture   # package boundaries and complete inspection catalogue
make verify         # definition of done: architecture + lint + test + fixtures + coverage + cleanroom
make stubs          # rebuild embedded PHP stubs index
make fixcheck       # apply every quick-fix on CUSTOS_CORPUS, require parsable output (FIXCHECK=php adds php -l samples; ~1-4 min)
go generate ./internal/php/syntax  # regenerate node kinds after editing ast.go
```

## Conventions

- Binaries go to `bin/` at the repository root: build with `make build`
  (→ `bin/custos`) or `go build -o bin/<name> ./cmd/<name>`; never run a bare
  `go build ./cmd/...` (it drops a binary in the working directory). Use
  `go build ./...` or `go vet ./...` to only check compilation.
- Go 1.27, standard library first; no new dependency without a clear reason.
- Performance is a feature: single AST walk per file with kind-indexed rule
  dispatch, arena-allocated nodes, lazy semantic data, parallel files. Add a
  benchmark for any hot path you touch; check `-benchmem` allocations.
- Byte offsets (`uint32`) everywhere; fixes are text edits on exact ranges.
- Rule IDs: EA short name without `Inspection`. `@noinspection <LegacyID>`
  and `@custos-ignore <ID>` must both suppress.
- Severity mapping: EA ERROR→error, WARNING→warning, WEAK WARNING→info.
- Correctness beats upstream fidelity: when EA behavior is wrong (false
  positives, fixes changing semantics or producing invalid PHP), diverge,
  document it in the spec's Divergences and in
  `internal/testing/conformance/testdata/ea-divergences.json`.
- No rule lands without own fixtures in
  `internal/inspection/rules/<id>/testdata/` (positive, false-positive cases,
  `.fixed.php` when it has a fix) covering 100% of its statements
  (`make coverage`; remove unreachable code rather than exempt it) and a green
  `make conformance RULE=<ID>` (or a documented, justified divergence in the
  spec's "Divergences" section).
- Specs are behavior only: no Go identifiers, repo paths, budget mechanics or
  implementer advice. Keep those in an optional
  `internal/inspection/rules/<id>/NOTES.md`, titled
  `# <RuleID> implementation notes`, with free `##` sections in our own
  words; create it only when there is something worth keeping.
- Messages: short, imperative, our own wording; no `[EA]` prefix.
- PHP versions 5.3–8.5 supported: version-gate rules via `meta`/spec thresholds.
- New code ships with tests covering every statement (`make coverage`);
  timing limits in tests use `testbudget.Of(d)` (load-scaled) against
  `testbudget.CPU()` for compute-bound work, and hard
  performance budgets live in `make bench`.
- Order: `make rules-doc` when a spec, a rule's fixtures or rule status
  changes, then `make verify` before declaring work done. CI (`docs.yml`)
  fails when `docs/rules/` or the rules sidebar is stale; the status table
  and explain texts are not checked, so regenerate them yourself.
- Before committing, check the staged diff for user-visible changes and ensure
  they are recorded under `## [Unreleased]` in `CHANGELOG.md`.
- User docs live in `docs/guide/` (check claims against the real binary);
  never put `<Tag>`-like text or `{{` in prose outside code (Vue compiles
  the pages).

## Architecture boundaries

`make architecture` enforces production dependency direction and catalogue
completeness. Each inspection exposes `New() analysis.Rule` from `rule.go`
and is listed explicitly in `internal/inspection/catalogue`; never import
another inspection or register through `init()`. Helper placement
(`astquery` / `flowquery` / `semanticquery`), the LSP `WithIndex` publication
boundary and adapter responsibilities are described once in
`docs/contributing/architecture.md`.
