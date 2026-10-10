# custos

Fast PHP inspector + fixer written in Go. Its rule catalogue includes native
inspections and rules modelled on Php Inspections (EA Extended); it is an
**independent clean-room implementation** released under MIT. Also runs as an LSP server (`custos lsp`) for editors.

Architecture and contributor practices: `docs/contributing/`. Per-rule status: `docs/internals/rules.md`.

Documentation describes current behavior and durable contributor practices.
Keep agent plans, progress logs and local verification reports in ignored
`.cache/`; never commit them as documentation. Keep catalogue totals in the
generated rule reference and status table, not in marketing or usage prose.

## Clean-room rules (MANDATORY)

The upstream project (`~/Sites/phpinspectionsea`, LGPL-2.1) must never leak
into this repo.

- NEVER copy, translate or paraphrase closely EA Java code, messages,
  descriptions, docs or fixtures into custos. No EA file is ever copied.
- Allowed facts: rule IDs (EA short name minus `Inspection`), groups, default
  severity/enabled flags, option names/defaults, PHP-version thresholds — all
  already captured in `internal/inspection/meta/rules.json` by `make extract`.
- Per rule, two separate steps done by separate agents/sessions:
  1. `spec-rule` skill: read EA, write `specs/<ID>.md` in our own words with
     new examples.
  2. `implement-rule` skill: implement **from the spec only** — do not open EA
     Java sources or EA fixtures while implementing.
- EA fixtures are only *read at test time* from the local checkout by
  `make conformance` (rule/range/severity + fix result compared; EA message
  text is never compared or stored).

## Layout

```text
cmd/custos/                       executable entry point
internal/php/{syntax,version,phpdoc}/ language, AST, traversal and source positions
internal/semantic/{names,types,index,infer,stubs}/ resolution and semantic computation
internal/inspection/analysis/     engine, context, semantic access and suppressions
internal/inspection/meta/         independent rule facts and generated descriptions
internal/inspection/catalogue/    explicit ordered inspection constructors
internal/inspection/rules/<lowercase-ID>/ private implementation, New(), companion tests
internal/inspection/{astquery,flowquery,semanticquery}/ focused shared inspection queries
internal/inspection/phpunit/      shared PHPUnit call/version behavior
internal/diagnostic/              findings, severity, lazy fixes, edits and positioned results
internal/fixing/                  edit conflicts, application and bounded fix loop
internal/project/                discovery, reading, indexing, analysis and prepared fixes
internal/project/{config,baseline}/ resolved settings and diagnostic baselines
internal/output/{report,diff}/    rendering adapters
internal/cli/                     commands, profiling, output, exit codes and file writes
internal/editor/lsp/              protocol, settings, workspace, documents and actions
internal/platform/safeio/         bounded filesystem primitives
internal/testing/{conformance,testbudget}/ fixture harness and test timing
tools/                          generators, architecture, coverage and clean-room checks
tools/internal/specmd/          spec section/front-matter/example parsing
specs/                           clean-room behavioral specification per stable rule ID
testdata/rules/<ID>/             own PHP fixtures, fixed bytes and option sidecars
docs/                           VitePress guide, contributing docs and generated rules
docs/internals/                 generated rule status
```

## Commands

```text
make build          # bin/custos
make test           # go test ./...
make lint           # golangci-lint + actionlint + editorconfig-checker + markdownlint (pinned; GOLANGCI=/ACTIONLINT=/ECCHECK=/MDLINT= to skip)
make fmt            # gofumpt + goimports + markdownlint --fix (run before committing)
make extract        # regenerate rule facts + local EA index (EA_PATH=~/Sites/phpinspectionsea)
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
- Correctness beats upstream fidelity: when EA behaviour is wrong (false
  positives, fixes changing semantics or producing invalid PHP), diverge,
  document it in the spec's Divergences and in `testdata/ea-divergences.json`.
- No rule lands without own fixtures in `testdata/rules/<ID>/` (positive,
  false-positive cases, `.fixed.php` when it has a fix) covering 100% of its
  statements (`make coverage`; remove unreachable code rather than exempt it) and a green
  `make conformance RULE=<ID>` (or a documented, justified divergence in the
  spec's "Divergences" section).
- Messages: short, imperative, our own wording; no `[EA]` prefix.
- PHP versions 5.3–8.5 supported: version-gate rules via `meta`/spec thresholds.
- New code ships with tests covering every statement (`make coverage`);
  timing limits in tests use `testbudget.Of(d)` (load-scaled), and hard
  performance budgets live in `make bench`.
- Run `make verify` before declaring work done; run `make rules-doc` when a
  spec, a rule's fixtures or rule status changes (CI fails on stale rule pages).
- User docs live in `docs/guide/` (check claims against the real binary);
  never put `<Tag>`-like text or `{{` in prose outside code (Vue compiles
  the pages).

## Architecture boundaries

`make architecture` enforces production dependency direction and catalogue
completeness. Each inspection exposes `New() analysis.Rule`; never import
another inspection or add registration through `init()`. Keep implementation
helpers local until shared; choose `astquery`, `flowquery` or `semanticquery`
by lexical, local-flow or resolved behavior. Keep PHPUnit support inspection-owned.

Project operations receive resolved options and explicit inspection lists.
Adapters own protocol/argument conversion, rendering and file writes. Publish
engine copies from `WithIndex` under the existing LSP synchronization boundary.
Diagnostic contracts are independent of inspections and reporting. See
`docs/contributing/architecture.md`.
