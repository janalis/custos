# Command line

```sh
custos analyse [flags] [paths...]   # report problems
custos fix [flags] [paths...]       # apply quick-fixes
custos explain <rule>               # describe a rule and its options
custos rules [--json]               # list rules
custos lsp                          # language server over stdio
custos mcp [flags]                  # MCP server for AI agents (stdio or HTTP)
custos version
```

`analyze` is accepted as an alias of `analyse`, `--version` of `version`, and
`help`, `-h` or `--help` print the list above. Run `custos <command> -h` for
the flags of a command. Flags can be written `--php 8.1` or `--php=8.1`.

## Paths

Paths on the command line are relative to the current directory. Without
paths, custos uses `paths` from `custos.json`, or else the
[project root](./configuration#project-root). Some directories are
[always skipped](./configuration#excluded-directories). Directories are
searched for files ending in `.php`, lowercase (plus the few files some rules
read, such as `composer.json`); `.phtml` or `.inc` files are only analyzed
when named on the command line, which works whatever the extension. A symlink
found in a directory is analyzed only when it leads to a file inside that
directory.

Before checking files one by one, custos indexes the symbols declared under
`paths` from `custos.json` (even when the command line names fewer files) and
under `vendor/` (or Composer's `config.vendor-dir`), so rules know the
classes and functions your code uses, whichever files a run checks. In
`vendor/`, test directories, `exclude` entries and the Composer-installed
`janalis/custos` package are left out. Run `composer install` first: without
`vendor/`, types from dependencies are unknown and rules that need them stay
silent.

The declarations of the files custos indexes without checking them are kept
between runs, under `custos` in the user cache directory (for example
`~/.cache/custos` on Linux, `~/Library/Caches/custos` on macOS), or in
`$CUSTOS_CACHE_DIR` when it is set. A file is parsed again when its size or
modification time changes, and the whole cache is set aside when the custos
binary, the target PHP version or the parse and inference options change, so
results are the same with or without it. `--no-cache` neither reads nor
writes it.

With `--flow-order`, the cache also keeps the order in which the run
summarized the cross-file flow of the checked files, and the next run with
the flag summarizes each file after the ones it calls, instead of again once
they settle. A repeated run over symfony/src takes about 12% less time.
Results stay the same, except that a chain of calls across files too long
for a run without the flag, whose end it reports as unresolved, may be
followed to the end. Setting `flow.order` to `true` in
[`custos.json`](./configuration) has the same effect, also in the language
server.

## Common flags

These apply to both `analyse` and `fix`:

| Flag | Meaning |
|---|---|
| `--php 8.1` | Target PHP version; overrides the [usual sources](./configuration#target-php-version). |
| `--rule A,B` | Run exactly these rules (rule IDs or PhpStorm names), even ones off by default or disabled in `custos.json`. |
| `--all` | Enable every rule, including the ones off by default, except those `custos.json` sets to `"enabled": false`. |
| `--comparison-style regular\|yoda` | Preferred operand order for comparison rules and their quick-fixes. |
| `--config DIR` | Start the search for the [project root](./configuration#project-root) here. |
| `--exclude a,b` | Extra directories to skip, as `exclude` in `custos.json`. |
| `--no-cache` | Do not reuse or save the [declarations of dependencies](#paths) between runs. |
| `--flow-order` | Summarize cross-file flow in the [order the last run saved](#paths): faster, and may follow longer call chains. |
| `--cpuprofile FILE` | Write a Go CPU profile of the analysis to FILE (for `go tool pprof`). |

## analyse

| Flag | Meaning |
|---|---|
| `--format text\|json\|checkstyle\|github\|sarif` | Output format (default `text`). |
| `--fail-on info\|warning\|error\|never` | Exit 1 when a finding has at least this severity (default `warning`). |
| `--baseline FILE` | Ignore the findings recorded in FILE, relative to the current directory. The file must exist. |
| `--generate-baseline FILE` | Record all current findings in FILE, relative to the current directory, and exit 0. |
| `--stats` | Print file and rule counts, the target PHP version and its source, and the time taken on stderr, plus the number of findings the baseline suppressed. |

Without `--baseline`, `analyse` uses `baseline` from `custos.json`. See
[Baseline](./suppressing#baseline).

Output formats:

- `text`: `file:line:col: severity: message [Rule] (fixable)` and a summary line.
- `json`: `{"files": N, "findings": [...], "analysis": {...}}`, each finding
  with `path`, `line`, `column`, `endLine`, `endColumn`, `rule`, `severity`,
  `message` and `fixable`.
- `checkstyle`: Checkstyle XML, read by many CI dashboards. Each error's
  `source` is `custos.<Rule>`.
- `github`: GitHub Actions workflow commands, shown as annotations on the
  pull request diff, titled with the rule ID. `info` findings become notices.
- `sarif`: SARIF 2.1.0, for GitHub code scanning and other SARIF viewers.
  `info` findings have the `note` level.

## Tool findings

Three findings come from custos itself rather than from a rule, always at
`error` severity:

| Rule | Meaning |
|---|---|
| `syntax` | The file does not parse; findings of other rules may be missing. |
| `io` | The file could not be read. |
| `internal` | A rule failed on the file; the other rules still ran. |

They count for `--fail-on` like any other error, so a syntax error fails the
default run, and they are never recorded in a [baseline](./suppressing#baseline).

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
| `internal-failure` | custos hit an unexpected error on the file; the checks it affected were skipped. |
| `status-limit` | Too many gaps; the list was cut short. |

## fix

| Flag | Meaning |
|---|---|
| `--dry-run` | Do not write files. |
| `--diff` | Print a unified diff of the changes. |

Quick-fixes are applied repeatedly until nothing changes (at most 10 rounds),
since one fix can reveal another finding. A file that cannot be read or
written is reported and the other files are still fixed; `fix` then exits 2.
custos never writes through symlinks. `fix` has no `--fail-on`: it exits 0
whether or not findings remain.

## explain and rules

```sh
custos explain UnnecessarySemicolon     # summary, options, suppression comment
custos rules                            # ✓ marks implemented rules; group, severity, default
custos rules --json                     # the full catalogue, for tooling
```

`rules --json` lists one object per rule with `id`, `legacyId` (the PhpStorm
name of EA rules), `group`, `severity`, `enabledByDefault`, `hasFix`,
`implemented`, and, when set, `experimental` and `options`.

## lsp

`custos lsp` serves the Language Server Protocol over stdin and stdout, the
only transport; `--stdio` is accepted for clients that always pass it. See
[Editors](./editors).

## mcp

`custos mcp` serves the Model Context Protocol to AI agents. See
[MCP server](./mcp) for setup and tools.

| Flag | Meaning |
|---|---|
| `--stdio` | Use stdin and stdout (the default). Cannot be combined with `--http`. |
| `--root DIR` | Workspace root for stdio. Cannot be combined with `--workspaces`. |
| `--workspaces FILE` | Private JSON file listing workspaces, identities and transport settings. |
| `--storage DIR` | Private directory for checkpoints, outside every workspace. |
| `--http ADDR` | Serve Streamable HTTP at ADDR instead of stdio. |
| `--tls-cert FILE`, `--tls-key FILE` | HTTPS certificate and key for `--http`. |

`mcp` takes flags only, no paths.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success; no finding at or above `--fail-on`. |
| `1` | Findings at or above `--fail-on`. |
| `2` | Usage error (including no or an unknown command), [configuration](./configuration#validation) error, or files `fix` could not process. |

`--fail-on` only applies to `analyse`.
