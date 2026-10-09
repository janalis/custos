# Architecture audit and migration evidence

The baseline is the repository HEAD preceding this migration. The original
`go test ./...` passed. Package imports were acyclic; the problem was ownership
and extension cost rather than an existing import cycle. Implementation used
Custos sources and existing clean-room specs; upstream sources were accessed
only through the existing conformance and clean-room commands.

## Findings and decisions

| Baseline evidence | Consequence | Implemented decision |
|---|---|---|
| `cmd/custos/main.go`: `prepare`, `cmdAnalyse`, `cmdFix`, `fixAll`, `readFixable` | Argument handling also owns indexing and parallel fixing. | CLI command files invoke reusable project operations; adapters retain rendering and writes. |
| `internal/lsp/server.go`: `configure`, `buildIndex`, `analyzeNow`, `codeActions` | Transport, settings, workspace and actions require one large file of context. | Separate settings, dispatch, workspace, documents, diagnostics and actions in the editor feature. |
| `internal/analysis/analysis.go`: rule/config contracts, engine and context | Dispatch and per-file semantic state are hard to navigate independently. | Separate contracts, engine, context and existing semantic/suppression files. |
| `internal/analysis/semantic.go`: `SetIndex` | Published engines can be reconfigured in place. | `WithIndex` returns a copy; LSP publishes it under the existing mutex. |
| Group `registered` slices, `register`, per-rule `init()` and root `rules.All` | Enumeration depends on global mutable initialization. | 178 private implementation packages with `New()` and an explicit ordered catalogue. |
| `internal/report/report.go`: `Item`, `Items`; `baseline` imports reporting | Baseline identity depends on an output adapter. | Neutral diagnostic contracts own positions; baseline and output consume them directly. |
| `internal/fix/fix.go`: concrete `*analysis.Engine` parameter | Edit application depends on inspection orchestration. | Fixing uses a parsed-file `Analyzer` interface and diagnostic edits. |
| `internal/analysis/util`: lexical calls, reaching writes, resolved hierarchy and values | Helper consumers must understand unrelated lexical and semantic semantics. | `astquery`, `flowquery`, `semanticquery`; PHPUnit support remains inspection-owned. |
| Rule companions such as `nore_*` and `php_unit_tests_*` | A category package shares unrelated implementation details. | Companions and private tests live with the inspection that uses them. |
| `infer.go`, `narrow.go`, `index/extract.go` | Related algorithms share caches but require oversized files. | Split cohesive implementation files within their existing semantic packages, retaining algorithms. |

No runtime dependency, rule, PHP threshold, configuration option or protocol
feature was added. Specifications, fixture paths, messages, serialized contracts,
execution order and externally visible enumeration remain stable.

## Original production package inventory

Each production package from the baseline is covered below. Dependency counts
include only direct Custos imports. Group package ownership is now individual
inspection packages; the complete per-ID inventory is linked below.

| Original package | Go files | Largest production file (lines) | Direct Custos imports | Current ownership |
|---|---:|---|---:|---|
| `cmd/custos` | 1 | `main.go` (498) | 13 | `cmd/custos` entry; `internal/cli` adapter |
| `internal/analysis` | 5 | `analysis.go` (397) | 8 | `internal/inspection/analysis` |
| `internal/analysis/util` | 33 | `values.go` (405) | 8 | `inspection/astquery`, `flowquery`, `semanticquery` |
| `internal/baseline` | 1 | `baseline.go` (150) | 3 | `internal/project/baseline` |
| `internal/config` | 1 | `config.go` (261) | 4 | `internal/project/config` |
| `internal/conformance` | 2 | `runner.go` (205) | 1 | `internal/testing/conformance` |
| `internal/diff` | 1 | `diff.go` (179) | 0 | `internal/output/diff` |
| `internal/fix` | 1 | `fix.go` (120) | 2 | `internal/fixing` |
| `internal/index` | 11 | `extract.go` (1125) | 5 | `internal/semantic/index` |
| `internal/infer` | 31 | `infer.go` (2257) | 7 | `internal/semantic/infer` |
| `internal/lsp` | 3 | `server.go` (821) | 8 | `internal/editor/lsp` |
| `internal/meta` | 2 | `meta.go` (76) | 0 | `internal/inspection/meta` |
| `internal/names` | 1 | `names.go` (285) | 1 | `internal/semantic/names` |
| `internal/phpdoc` | 3 | `phpdoc.go` (409) | 0 | `internal/php/phpdoc` |
| `internal/phpver` | 1 | `phpver.go` (72) | 0 | `internal/php/version` |
| `internal/report` | 1 | `report.go` (205) | 3 | `internal/output/report` |
| `internal/rules` | 1 | `rules.go` (31) | 13 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/architecture` | 10 | `callable_parameter_use_case_in_type_context.go` (547) | 10 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/codestyle` | 31 | `is_empty_function_usage.go` (308) | 8 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/compatibility` | 5 | `deprecated_ini_options.go` (129) | 5 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/confusing` | 6 | `referencing_objects.go` (252) | 8 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/controlflow` | 26 | `foreach_invariants.go` (967) | 8 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/langmigration` | 27 | `return_type_can_be_declared.go` (599) | 9 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/performance` | 21 | `cascade_string_replacement.go` (696) | 6 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/phpunit` | 9 | `php_unit_tests_asserts.go` (664) | 6 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/probablebugs` | 39 | `null_pointer_exception.go` (858) | 10 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/security` | 13 | `security_advisories.go` (366) | 5 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/typecompat` | 4 | `type_unsafe_comparison.go` (316) | 5 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/rules/unused` | 10 | `only_writes_on_parameter.go` (549) | 8 | `inspection/catalogue` and `inspection/rules/<lowercase-ID>` |
| `internal/runner` | 3 | `runner.go` (152) | 7 | `internal/project` |
| `internal/safeio` | 2 | `safeio.go` (54) | 0 | `internal/platform/safeio` |
| `internal/stubs` | 2 | `legacy.go` (81) | 2 | `internal/semantic/stubs` |
| `internal/syntax` | 17 | `lexer.go` (1153) | 1 | `internal/php/syntax` |
| `internal/testbudget` | 1 | `testbudget.go` (49) | 0 | `internal/testing/testbudget` |
| `internal/types` | 7 | `parse.go` (585) | 2 | `internal/semantic/types` |
| `tools/cleanroom` | 1 | `main.go` (121) | 0 | `tools/cleanroom` |
| `tools/covercheck` | 1 | `main.go` (156) | 0 | `tools/covercheck` |
| `tools/extract` | 1 | `main.go` (585) | 1 | `tools/extract` |
| `tools/genexplain` | 1 | `main.go` (79) | 1 | `tools/genexplain` |
| `tools/genkinds` | 1 | `main.go` (56) | 0 | `tools/genkinds` |
| `tools/genstubs` | 1 | `main.go` (155) | 4 | `tools/genstubs` |
| `tools/internal/specmd` | 1 | `specmd.go` (66) | 0 | `tools/internal/specmd` |
| `tools/relprep` | 1 | `main.go` (192) | 0 | `tools/relprep` |
| `tools/rulesdoc` | 1 | `main.go` (122) | 1 | `tools/rulesdoc` |
| `tools/rulesref` | 1 | `main.go` (266) | 3 | `tools/rulesref` |

## Final ownership map

The maintained [architecture guide](../contributing/architecture.md) contains
the dependency diagram, contracts and extension instructions. All 178 packages
under `inspection/rules` correspond to lowercase IDs from independent metadata.
[Rule status](rules.md) supplies the complete specification, fixture and
conformance inventory; `catalogue/catalogue.go` supplies the actual execution
order. Each package has exactly one explicit catalogue constructor.

| Current production package | Direct Custos dependencies |
|---|---|
| `cmd/custos` | `internal/cli` |
| `internal/cli` | `internal/diagnostic`, `internal/editor/lsp`, `internal/inspection/analysis`, `internal/inspection/catalogue`, `internal/inspection/meta`, `internal/output/diff`, `internal/output/report`, `internal/php/syntax`, `internal/php/version`, `internal/project`, `internal/project/baseline`, `internal/project/config` |
| `internal/diagnostic` | `internal/php/syntax` |
| `internal/editor/lsp` | `internal/diagnostic`, `internal/fixing`, `internal/inspection/analysis`, `internal/inspection/catalogue`, `internal/inspection/meta`, `internal/php/syntax`, `internal/project`, `internal/project/config`, `internal/semantic/index` |
| `internal/fixing` | `internal/diagnostic`, `internal/php/syntax` |
| `internal/inspection/analysis` | `internal/diagnostic`, `internal/inspection/meta`, `internal/php/syntax`, `internal/php/version`, `internal/semantic/index`, `internal/semantic/infer`, `internal/semantic/names`, `internal/semantic/stubs`, `internal/semantic/types` |
| `internal/inspection/astquery` | `internal/diagnostic`, `internal/php/syntax` |
| `internal/inspection/catalogue` | `inspection/analysis` plus all 178 inspection packages |
| `internal/inspection/flowquery` | `internal/inspection/astquery`, `internal/php/syntax` |
| `internal/inspection/meta` | `internal/diagnostic` |
| `internal/inspection/phpunit` | `internal/inspection/analysis`, `internal/inspection/astquery`, `internal/php/syntax` |
| `internal/inspection/semanticquery` | `internal/diagnostic`, `internal/inspection/analysis`, `internal/inspection/astquery`, `internal/inspection/flowquery`, `internal/php/phpdoc`, `internal/php/syntax`, `internal/php/version`, `internal/semantic/index`, `internal/semantic/infer`, `internal/semantic/names` |
| `internal/output/diff` | Standard library only |
| `internal/output/report` | `internal/diagnostic`, `internal/inspection/meta` |
| `internal/php/phpdoc` | Standard library only |
| `internal/php/syntax` | `internal/php/version` |
| `internal/php/version` | Standard library only |
| `internal/platform/safeio` | Standard library only |
| `internal/project` | `internal/diagnostic`, `internal/fixing`, `internal/inspection/analysis`, `internal/php/syntax`, `internal/platform/safeio`, `internal/project/config`, `internal/semantic/index`, `internal/semantic/infer`, `internal/semantic/stubs` |
| `internal/project/baseline` | `internal/diagnostic`, `internal/php/syntax`, `internal/platform/safeio` |
| `internal/project/config` | `internal/diagnostic`, `internal/inspection/analysis`, `internal/inspection/meta`, `internal/php/version`, `internal/platform/safeio` |
| `internal/semantic/index` | `internal/php/phpdoc`, `internal/php/syntax`, `internal/php/version`, `internal/semantic/names`, `internal/semantic/types` |
| `internal/semantic/infer` | `internal/php/phpdoc`, `internal/php/syntax`, `internal/php/version`, `internal/semantic/index`, `internal/semantic/names`, `internal/semantic/stubs`, `internal/semantic/types` |
| `internal/semantic/names` | `internal/php/syntax` |
| `internal/semantic/stubs` | `internal/php/version`, `internal/semantic/index` |
| `internal/semantic/types` | `internal/php/syntax`, `internal/semantic/names` |
| `internal/testing/conformance` | `internal/diagnostic` |
| `internal/testing/testbudget` | Standard library only |
| `tools/architecture` | Standard library only |
| `tools/cleanroom` | Standard library only |
| `tools/covercheck` | Standard library only |
| `tools/extract` | `internal/diagnostic`, `internal/inspection/meta` |
| `tools/genexplain` | `tools/internal/specmd` |
| `tools/genkinds` | Standard library only |
| `tools/genstubs` | `internal/php/syntax`, `internal/php/version`, `internal/semantic/index`, `internal/semantic/types` |
| `tools/internal/specmd` | Standard library only |
| `tools/relprep` | Standard library only |
| `tools/rulesdoc` | `internal/inspection/meta` |
| `tools/rulesref` | `internal/diagnostic`, `internal/inspection/meta`, `internal/testing/conformance`, `tools/internal/specmd` |

## Shared helper consumers

This inventory counts production inspection packages, excluding tests. Query
ownership follows behavior rather than the former category. Helpers whose
only consumer is another query remain internal building blocks; exported
state needed by value discovery remains within inspection query packages.

| Query package | Production inspection consumers | Exported function calls from inspection packages |
|---|---:|---:|
| `astquery` | 144 | 542 |
| `flowquery` | 25 | 40 |
| `semanticquery` | 84 | 152 |
| `phpunit` | 4 | 20 |

The old `Terminates` and `perfPlainArgs` forwarding functions were removed;
consumers call `syntax.Terminates` and `astquery.ArgValues`. Duplicate name-span
logic uses `astquery.NamePartSpan`. Single-inspection companions and helpers
were assigned by declaration references rather than filename prefixes alone.
Lexical statement shapes, argument extraction, whitespace equivalence and exact
source-edit builders belong to `astquery`; they require no semantic context.

## Execution paths

- Analyse: CLI resolves config and overrides → `project.Open` → pattern-aware
  discovery → conditional vendor/project indexing with retained source bytes →
  report-mode parallel analysis → positioned diagnostics → baseline filter →
  output adapter and severity exit decision.
- Fix: the same resolved project → optional indexing → bounded regular-file
  reads → analyzer/fix loop → ordered prepared outcomes → CLI diff/write decision.
  Symlinks remain skipped, and no project operation writes files.
- Editor: transport dispatch → settings/configured engine → workspace index
  replacement → document version snapshot → project buffer analysis → stale
  version rejection → UTF-16 diagnostics and lazy/inline actions. Suppression and
  fix-all continue to use exact diagnostic byte edits.
- Rule documentation: independent metadata + explicit catalogue + existing
  specs and own fixtures → generated descriptions/reference/status pages.

## Extension scenarios A–G

| Scenario | Owned edits and practical validation |
|---|---|
| A — Add an inspection | One ID package, one catalogue entry, its spec and fixture directory. All 178 constructors and correspondence are checked by `make architecture`; constructor statements are exercised by fixture coverage. No category registry or adapter change is needed. |
| B — Improve inference | Edit the appropriate semantic file and colocated tests. Package boundaries forbid inference from importing adapters, orchestration or inspection implementations; existing inference tests run unchanged. |
| C — Add a quick-fix | Generation belongs to its inspection; edits are diagnostic contracts. Generic application belongs to fixing, file preparation to project, writes to CLI. Conflict, iteration and fix-all tests retain behavior. |
| D — Add an output format | Edit `output/report` and its adapter tests; consume positioned diagnostic items. Baselines and analysis have no report dependency. All five existing formats were compared with the original binary. |
| E — Extend PHP syntax | Edit `php/syntax`, regenerate kinds and add parser/version tests. No semantic or inspection import is permitted from PHP packages. Existing PHP/Unicode tests validate downstream positions. |
| F — Fix a false positive | Locate the lowercase ID package and stable spec/fixtures by the same ID. Single-rule fixtures and conformance commands still work. This migration changes ownership, not rule behavior. |
| G — Parallel development | Independent inspections have private types, helpers and companion tests. Shared edits are limited to catalogue insertion, deliberate query/semantic changes and generated status. Registration is explicit and deterministic. No parallel agents were used during this migration. |

## Verification and performance

Exact command outcomes and benchmark comparisons are recorded in
[verification results](architecture-verification.md). Corpus-dependent skips are
reported separately from executed checks. The initial broad benchmark command
completed inference measurements but its runner stage failed after packages
moved; it is not counted as a completed baseline benchmark run. A separate
baseline checkout is used for final repeated comparisons.

## Remaining design limits

- Query APIs expose some reusable value-walk state to share lexical/local-flow
  collection with resolved discovery. Moving that state into semantic inference
  would incorrectly make PHP/semantic services depend on inspections.
- Project options, engine rules and metadata must be treated as immutable during
  analysis, as before. `WithIndex` eliminates index installation mutation; it does
  not promise a deep copy of all configuration maps.
- The explicit catalogue is a shared edit for inspection additions. Its order is
  externally significant, so generation or unordered registration is avoided.
- The architecture checker validates production imports and static ID/constructor
  correspondence. The compiler, fixtures, coverage, race tests and corpus checks
  supply behavioral guarantees.
- Original implementation algorithms and caches are preserved. Further algorithm
  changes need their own measured correctness/performance evidence.
