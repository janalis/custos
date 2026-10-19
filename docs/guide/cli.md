# Command line

```sh
custos analyse [flags] [paths...]   # report problems
custos fix [flags] [paths...]       # apply quick-fixes
custos explain <rule>               # describe a rule and its options
custos rules [--json]               # list rules
custos lsp                          # language server over stdio
custos version
```

`analyze` is accepted as an alias of `analyse`. Run `custos <command> -h` for
the flags of a command. Flags can be written `--php 8.1` or `--php=8.1`.

## Paths

Paths on the command line are relative to the current directory. Without
paths, custos uses `paths` from `custos.json`, or else the
[project root](./configuration#project-root). Some directories are
[always skipped](./configuration#excluded-directories). Directories are
searched for `.php` files (plus the few files some rules read, such as
`composer.json`); a file named on the command line is analyzed whatever its
extension.

## Common flags

These apply to both `analyse` and `fix`:

| Flag | Meaning |
|---|---|
| `--php 8.1` | Target PHP version; overrides the [usual sources](./configuration#target-php-version). |
| `--rule A,B` | Run only these rules (rule IDs or PhpStorm names). |
| `--all` | Enable every rule, including the ones off by default. |
| `--comparison-style regular\|yoda` | Preferred operand order for comparison rules and their quick-fixes. |
| `--config DIR` | Start the search for the [project root](./configuration#project-root) here. |
| `--exclude a,b` | Extra directories to skip, as `exclude` in `custos.json`. |
| `--cpuprofile FILE` | Write a Go CPU profile of the analysis to FILE (for `go tool pprof`). |

## analyse

| Flag | Meaning |
|---|---|
| `--format text\|json\|checkstyle\|github\|sarif` | Output format (default `text`). |
| `--fail-on info\|warning\|error\|never` | Exit 1 when a finding has at least this severity (default `warning`). |
| `--baseline FILE` | Ignore the findings recorded in FILE, relative to the current directory. The file must exist. |
| `--generate-baseline FILE` | Record all current findings in FILE, relative to the current directory, and exit 0. |
| `--stats` | Print file and rule counts, the target PHP version and its source, and the time taken on stderr, plus the number of findings the baseline suppressed. |

Without `--baseline`, `analyse` uses `baseline` from `custos.json`, relative
to the project root; a missing file there suppresses nothing. See
[Baseline](./suppressing#baseline).

Output formats:

- `text`: `file:line:col: severity: message [Rule] (fixable)` and a summary line.
- `json`: `{"files": N, "findings": [...], "analysis": {...}}`, each finding
  with `path`, `line`, `column`, `endLine`, `endColumn`, `rule`, `severity`,
  `message` and `fixable`.
- `checkstyle`: Checkstyle XML, read by many CI dashboards.
- `github`: GitHub Actions workflow commands, shown as annotations on the
  pull request diff. `info` findings become notices.
- `sarif`: SARIF 2.1.0, for GitHub code scanning and other SARIF viewers.
  `info` findings have the `note` level.

## Analysis status

Besides findings, custos reports whether each file was analyzed completely.
When a rule could not reach a conclusion somewhere, for example because the
code is too dynamic or an analysis budget ran out, the file is *incomplete*:
custos then stays silent there rather than guess, so a clean result is not a
proof that the code is correct.

- The CLI prints `custos: analysis incomplete in N file(s): reasons` on
  stderr. The status never changes the exit code.
- JSON output has `analysis.complete` and `analysis.incomplete` (file, range
  and reason of each gap); SARIF has the same object in the run's
  `properties.analysis`.
- The language server logs a message when a document's status changes; see
  [Editors](./editors).

| Reason | Meaning |
|---|---|
| `dynamic-code` | Code built or evaluated at runtime. |
| `unresolved-dispatch` | A call whose target or behavior is unknown. |
| `unsupported-control-flow` | Control flow the analysis does not model. |
| `query-budget`, `flow-budget`, `argument-budget`, `inference-budget` | An analysis limit was reached. |
| `inference-recursion` | A type depends on itself recursively. |
| `syntax-recovery` | The file has a syntax error; the parser recovered. |
| `finding-limit` | The file has too many findings; the rest are dropped. |
| `rule-failure` | A rule failed on the file. |
| `io-error` | The file could not be read. |
| `internal-failure` | The language server hit an unexpected error. |
| `status-limit` | Too many gaps; the list was cut short. |

## fix

| Flag | Meaning |
|---|---|
| `--dry-run` | Do not write files. |
| `--diff` | Print a unified diff of the changes. |

Quick-fixes are applied repeatedly until nothing changes (at most 10 rounds),
since one fix can reveal another finding. A file that cannot be read or
written is reported and the other files are still fixed; `fix` then exits 2.
custos never writes through symlinks.

## explain and rules

```sh
custos explain UnnecessarySemicolon     # summary, options, suppression comment
custos rules                            # ✓ marks implemented rules; group, severity, default
custos rules --json                     # the full catalogue, for tooling
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success; no finding at or above `--fail-on`. |
| `1` | Findings at or above `--fail-on`. |
| `2` | Usage or [configuration](./configuration#validation) error, or files `fix` could not process. |
