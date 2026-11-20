# Command line

```sh
custos analyse [flags] [paths...]   # report problems
custos fix [flags] [paths...]       # apply quick-fixes
custos format [flags] [paths...]    # format code (--check to only report)
custos explain <rule>               # describe a rule and its options
custos rules [--json]               # list rules
custos test-rules [dir]             # run custom rules over their fixtures
custos lsp                          # language server over stdio
custos mcp [flags]                  # MCP server for AI agents (stdio or HTTP)
custos self-update [--check]        # update to the latest release
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

When your code calls a function or method of a dependency, custos reads the
one file under `vendor/` that declares it and summarizes it, once per run (or
editor session), so request data passed through library wrappers is followed
like data passed through your own. Each dependency file is summarized on its
own; files over 512 KB and callables declared more than once are left out.

The declarations of the files custos indexes without checking them are kept
between runs, under `custos` in the user cache directory (for example
`~/.cache/custos` on Linux, `~/Library/Caches/custos` on macOS), or in
`$CUSTOS_CACHE_DIR` when it is set. A file is parsed again when its size or
modification time changes, and the whole cache is set aside when the custos
binary, the target PHP version or the parse and inference options change, so
results are the same with or without it. `--no-cache` neither reads nor
writes it.

The same cache keeps the cross-file summaries of the checked files. A run
after an edit summarizes again only what the edit reaches: a file whose
source did not change keeps the summaries the code it calls still supports.
The summaries are set aside, and the run summarizes every file, when a
checked file declares something else (a new function, another type or
default; moving code or editing bodies does not count), when the set of
checked files changes, when a file read for its declarations but not
checked changes (a dependency, or a project file outside the checked
paths), when the command runs from another directory, or when such a file
was modified in the last two seconds. One set of summaries is kept per
project and settings: runs over different paths replace each other's.
Findings are the same with or without them.

`analyse` also keeps each checked file's findings. A later run reports them
again, without checking the file, when the file's source is the same and
nothing its findings depend on changed: the checked files' declarations,
the cross-file summaries, the files read for their declarations, Composer's
metadata, the enabled rules and their settings, the baseline, `--parse-once`
and the directory the command runs from. An edit inside a function body that
leaves what the function does for its callers as it was is checked alone;
one that changes it, or changes a declaration, has every file checked again.
`fix` never takes findings from the cache.

Cross-file flow follows dependencies even on a first run. Recursive groups
are analyzed under bounded work limits; summaries that do not settle are
left unresolved.

With `--flow-order`, the cache also keeps the order in which the run
summarized the checked files. The next run uses it as a scheduling hint,
which can avoid repeating summaries during dependency discovery. Setting
`flow.order` to `true` in [`custos.json`](./configuration) has the same
effect, also in the language server.

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
| `--no-cache` | Do not reuse or save the [declarations of dependencies, cross-file summaries and findings](#paths) between runs. |
| `--flow-order` | Summarize cross-file flow in the [order the last run saved](#paths): a scheduling hint for dependency discovery. |
| `--parse-once` | Parse each analyzed file once and keep its syntax tree until it is analyzed, instead of parsing it again in each phase, and keep the types inferred for a file between its cross-file flow summaries. Compared with a run without it: faster (about 1–12% on large projects, most where flow summaries are redone often) at the cost of memory (about 30–65% more at peak). Findings are the same, except possibly where type inference reaches its recursion limit. `analyse` only: `fix` ignores it. |
| `--cpuprofile FILE` | Write a Go CPU profile of the analysis to FILE (for `go tool pprof`). Samples taken while a rule runs carry a `rule` label with its ID: `go tool pprof -tags FILE` lists the time per rule. Data rules share, such as inferred types and flow, is charged to the first rule that needs it in each file. |
| `--memprofile FILE` | Write a Go allocation profile of the analysis to FILE (for `go tool pprof`). |

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

## format

`custos format` lays out the code of the same files as `analyse` (minus the
files `format.exclude` lists) with the project's format configuration. See
[Formatting](./formatting) for presets, settings and ignore comments, and
[Format settings](./format-settings) for every setting.

```sh
custos format                           # rewrite the files that change
custos format --check                   # list what would change; exit 1 if anything would
custos format --dry-run --diff src      # print a unified diff, write nothing
custos format --stdin --stdin-path src/A.php < src/A.php   # standard input to standard output
custos format --show-config src         # each setting, its value and its source
```

| Flag | Meaning |
|---|---|
| `--check` | Write nothing; list each change and exit 1 if a file would change. |
| `--dry-run` | Write nothing; as `--check` for files. With `--stdin`, print the formatted code and exit 1 if it changed. |
| `--diff` | Print a unified diff of the changes instead of the listing. Files are written unless `--check` or `--dry-run` is given. |
| `--report text\|json\|checkstyle\|github\|sarif` | Listing format of the changes (default `text`); not with `--diff`. |
| `--stdin`, or the path `-` | Format standard input to standard output. Not combined with other paths. |
| `--stdin-path FILE` | File standard input stands for: its configuration, `.editorconfig`, `format.exclude` and the name in messages. |
| `--preset NAME` | `auto`, `psr1`, `psr12`, `per`, `symfony`, `laravel` or `drupal`; overrides `format.preset`. |
| `--set id=value` | Set one [setting](./format-settings), over every other source; repeatable. The value is JSON when it parses (`--set line_length=100`), else text; a list takes comma-separated items (`--set trailing_comma=arrays,match`). |
| `--allow-risky` | Apply the preset's [risky settings](./formatting#risky-settings). |
| `--show-config` | Print each setting for the first path (default `.`) with its value and source, then exit. |
| `--php 8.1` | Target PHP version: the formatter parses at it, and version-dependent spellings (trailing commas, short `list()`) follow it. |
| `--config DIR` | Start the search for the [project root](./configuration#project-root) here. |
| `--exclude a,b` | Extra directories to skip, as `exclude` in `custos.json`. |
| `--no-cache` | Neither skip nor remember the files known to be formatted. |

The listing goes to stdout, one line per change under its file, with the
setting behind it (`format.<setting>`); the summary
(`N file(s) reformatted, N unchanged, N skipped`) goes to stderr. Without
`--check` or `--dry-run`, only what formatting cannot fix (long lines, with
`report_long_lines`) is listed. Files with syntax errors or an ignore comment
are skipped; the first ones with a warning.

## explain and rules

```sh
custos explain UnnecessarySemicolon     # summary, options, suppression comment
custos rules                            # ✓ marks implemented rules; group, severity, default
custos rules --json                     # the full catalogue, for tooling
```

`rules --json` lists one object per rule with `id`, `legacyId` (the PhpStorm
name of EA rules), `group`, `severity`, `enabledByDefault`, `hasFix`,
`implemented`, and, when set, `experimental` and `options`. Both list and
explain the project's [custom rules](./custom-rules) too.

## test-rules

```sh
custos test-rules                       # fixtures under custos-tests/ (customRuleTests)
custos test-rules path/to/fixtures      # another directory
```

Runs each rule alone over its fixtures and compares findings and quick-fixes
with what they expect (see [Testing rules](./custom-rules#testing-rules)).
`--config` names the directory to look for `custos.json` from.

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

## self-update

`custos self-update` updates custos the way it was installed. See
[Updating](./installation#updating) for how each install is handled and
for the notice that `analyse` and `fix` print when a release is out.

| Flag | Meaning |
|---|---|
| `--check` | Only report whether a newer release exists; exit 1 when one does. |
| `--version X.Y.Z` | Install this release instead of the latest, including an older one. Release binaries only. |

## Memory

`analyse` and `fix` run Go's garbage collector less often than the default
(`GOGC=200` instead of `100`): most of their heap stays live until the run
ends, so collecting less saves time at the cost of a larger peak heap. To
keep that peak within the machine, they also set a soft memory limit of three
quarters of the physical memory (on Linux, of the cgroup memory limit when
the process's group or one of its ancestors sets a lower one), near which the
collector works harder again. Where the memory cannot be read (on Windows and
the BSDs, or without `/proc` on Linux), they keep the runtime defaults.

Setting `GOGC` in the environment, to any value including `off`, leaves the
collector entirely to it: neither setting is changed, for example `GOGC=100`
to trade speed for a smaller peak. Setting only `GOMEMLIMIT` replaces the
soft limit and keeps `GOGC=200`. `lsp` and `mcp` keep the runtime defaults.

The soft limit is per process: several runs at once (`make -j`, a CI matrix
on one machine) may each grow toward three quarters of the memory. For
parallel runs, set `GOMEMLIMIT` to each process's share, or `GOGC=100`.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success; no finding at or above `--fail-on`. |
| `1` | Findings at or above `--fail-on`; with `format --check` or `--dry-run`, a file would change; with `self-update --check`, a newer release exists; with `test-rules`, a fixture failed. |
| `2` | Usage error (including no or an unknown command), [configuration](./configuration#validation) error, files `fix` could not process, files `format` could not format (a read or write error, or a result its safety check refused), or a failed update. |

`--fail-on` only applies to `analyse`.
