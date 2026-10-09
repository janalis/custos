# Architecture overhaul verification

Verification used Go 1.27.1 on macOS/arm64. The baseline is commit `eb371a6`,
compiled before implementation; repeated baseline benchmarks used a detached
checkout of that commit. No upstream implementation source was opened during
implementation. Local EA data was read only by the existing conformance and
clean-room tools.

## Behavior and acceptance checks

| Check | Outcome |
|---|---|
| Original `go test ./...` | Passed before migration. |
| `make fmt` | Passed; the untracked architecture plan is excluded from formatting. |
| `make architecture` | Passed: dependency direction, cross-rule imports, constructors, unique/complete catalogue, directory IDs, specs and fixtures. |
| `make verify` | Passed: architecture, lint, tests, own fixtures, both coverage gates and clean-room scan. |
| `make coverage` | Inspection implementations **100%**; all production command, internal and tool statements **100%**, with the existing one-statement command-entry exemption. Coverage subprocess concurrency is bounded at four because instrumented binaries include the complete module. |
| CLI baseline comparison | **206 exact comparisons** of exit status, stdout, stderr and fixed bytes passed. Included all 178 explanations, rule listing, all five output formats, PHP 5.3/5.6/7.0/7.4/8.0/8.4/8.5, suppression aliases, Unicode, configuration precedence, baseline generation/filtering, dry-run diffs and written fixes. |
| `go test -race ./internal/editor/lsp ./internal/project ./internal/inspection/analysis ./internal/fixing` | Passed. Existing editor tests cover settings, queued indexes, declaration changes, stale results, lazy/inline actions, suppression and fix-all. |
| `make conformance` | Passed: **261 case passes**, **54 existing documented divergence skips**, 315 cases total. The separate parse check processed **436 fixtures with zero parse errors**. Skips do not count as successful behavioral comparisons. |
| `CUSTOS_CORPUS=… make fixcheck FIXCHECK=php` | Passed on the local Biblindex `src` corpus: **2,332 fixes on 903 PHP files**, 402 files checked with combined fixes, **zero broken fixes**. PHP lint sampling was enabled. |
| `CUSTOS_BENCH_FILE=… GOMAXPROCS=1 make bench` | Passed all benchmarks and the enforced real-file editor latency budget: p50 **12.02 ms**, p95 **20.27 ms** against 30 ms. Earlier intermittent timing failures remain recorded below. |
| `make fuzz` | All **10 fuzz targets** passed their 30-second smoke runs. |
| `go generate ./internal/php/syntax` | Passed. |
| `make rules-doc` | Passed: 178 rules, 178 specifications and 178 passing own-fixture statuses; rule pages remain current. |
| `make docs` | Passed VitePress production build. |
| `git diff --check` | Passed. |

The first `make fixcheck` without `CUSTOS_CORPUS` skipped; it is not counted as
validation. It was followed by the explicit corpus run above. The default
whole-suite real-file lexer/parser benchmarks and edit-latency test also skip
without `CUSTOS_BENCH_FILE`; explicit real-file runs are recorded below.

`ARCHITECTURE-PLAN.md` remains untracked and byte-identical to the original,
including its absent terminal newline. SHA-256:
`800570f2fc2e2b62a494dfc295fb0db3965b40ecc46f3fd25dbd3414a8ff68d8`.
Markdownlint and EditorConfig checking exclude that preserved file.
`CLAUDE.md` remains a symlink to `AGENTS.md`.

## Performance comparison

Five repeated baseline/final runs used `-benchmem -benchtime=200ms -count=5`
for syntax, inference, inspection dispatch and project operations. Five more
alternating pairs used `GOMAXPROCS=4`, reversing old/new execution order on
successive pairs. There are **49 common benchmark scenarios**. Allocation
counts match in the small deterministic workloads, including buffer analysis
(**985 allocations**, approximately 151,214 bytes) and inspection dispatch
(**811 allocations**, approximately 10,151 bytes). Three large inference cases
varied by one allocation between independent runs; byte counts varied below
0.5%, consistent with existing map/arena behavior.

Wall-clock comparisons were not reliable: unrelated Node/Vitest, browser and
PHP processes ran concurrently. The unchanged baseline itself slowed on a
control rerun. Alternating buffer ratios ranged from 0.76 to 1.47; dispatch,
parser and inference scenarios also changed with run order. These measurements
are retained as evidence, not presented as a speed regression or improvement.
The initial benchmark attempted during migration also failed at its old runner
package path and does not count as a completed benchmark run.

To investigate apparent regressions above 5%, a temporary combined probe
compiled the original and migrated packages into **one process**. Each workload
ran 41 alternating batches, reversing order each batch, with `GOMAXPROCS=1`.
Three complete probe runs measured process CPU time using `getrusage`.
Garbage collection was disabled within batches and run between paired batches,
so collection triggered by one implementation could not be charged to the next.
Independent `-benchmem` results retain allocation and ordinary GC evidence.
The table reports the median of the three median paired ratios; independently
rounded old/new CPU medians need not yield that ratio.

| Matched workload | Original CPU ns/op | Migrated CPU ns/op | Paired change |
|---|---:|---:|---:|
| Lex 116 KB corpus file | 506,906 | 503,250 | −0.35% |
| Parse 116 KB corpus file | 1,316,312 | 1,340,375 | +0.40% |
| PHP clone argument parsing | 3,493 | 3,486 | −0.23% |
| Array builtin inference | 8,438 | 8,460 | +0.40% |
| Inspection dispatch | 11,641 | 11,624 | +0.83% |
| Parse and analyze an editor buffer | 288,297 | 283,023 | −1.30% |
| Inference environment and variable types | 117,023 | 120,051 | −0.63% |
| Full real-file editor analysis (separate probe) | 11,301,000 | 11,226,500 | +0.45% |

No reproducible CPU regression above 5% remains in these matched hot paths.
The loaded host still prevents establishing a tight wall-clock throughput
bound for every scenario. A quiet machine is required to certify that bound;
the paired CPU probe does not substitute for end-to-end latency measurement.

Real-file `BenchmarkLex` and `BenchmarkParse` were also run with `-benchmem`,
500 ms batches and five repetitions, using
`CUSTOS_BENCH_FILE=/Users/pierre/Sites/biblindex-chore-custos-cleanup/src/Repository/QuotationRepository.php`
(116,279 bytes). The same file was used for explicit LSP edit-latency checks.

| Real-file benchmark | Original median ns/op | Migrated median ns/op | Original / migrated bytes | Allocations, both |
|---|---:|---:|---:|---:|
| Parse | 3,915,878 | 2,625,568 | 1,053,965 / 1,053,937 | 2,801 |
| Lex | 1,010,364 | 797,876 | 352,419 / 352,419 | 14 |

These independent wall-clock medians are included for completeness; the paired
CPU comparison above is the meaningful comparison on this host.

The real-file latency budget had **intermittent failures** before the final
complete benchmark run passed.
The first four-thread baseline run measured p50 12.30 ms / p95 16.18 ms;
the migrated run measured 12.55 ms / 36.22 ms and failed the 30 ms budget.
Five single-thread repeats measured baseline p95 20.19–28.07 ms (five passes)
and migrated p95 22.41–37.27 ms (two passes, three failures). Concurrent host
load and GC affect tails. A matched real-file CPU probe was added to investigate
whether the full real-file analysis itself regressed: its median paired CPU
change was +0.45%; the largest increase across the three runs was 0.97%. These timing failures
remain visible. The subsequent complete `make bench` run passed with p50
12.02 ms and p95 20.27 ms on the same file (`GOMAXPROCS=1`); ordinary
correctness verification does not enforce latency
without `CUSTOS_PERF=1`.

## Evidence and reproduction

Local raw logs, comparison scripts and the temporary paired probe are retained
under `.cache/architecture/` (ignored build evidence). The principal filenames
are `verify.log`, `compatibility.log`, `conformance.log`, `race.log`,
`fixcheck-corpus.log`, `fuzz.log`, `docs.log`, `bench-baseline.log`,
`bench-final.log`, `make-bench-final.log`, `paired-comparison.json`,
`matched-realfile.log` and `matched-cpu-nogc.log`.
The probe imports original packages that no longer exist in this checkout;
reproduce it in a detached baseline checkout with the migrated feature trees
copied alongside the original packages. It is measurement tooling, not a
production dependency or a retained compatibility layer.

The ownership map, audit of every original production package, helper consumer
inventory, migration rationale and extension scenarios A–G are in
[the architecture audit](architecture-audit.md). Contributor instructions and
the dependency diagram are in [the architecture guide](../contributing/architecture.md).
