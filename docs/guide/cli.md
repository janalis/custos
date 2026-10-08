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

Without paths, custos uses `paths` from `custos.json`, or else the project
root: the directory holding `custos.json` or `composer.json`. `vendor`,
`node_modules`, `.git`, `.idea`, `.custos` and `var/cache` are always
skipped. To skip more, use `exclude` in `custos.json` or `--exclude a,b`.
Directories are searched for `.php` files (plus the few files some rules
read, such as `composer.json`); a file named on the command line is analysed
whatever its extension.

## Common flags

These apply to both `analyse` and `fix`:

| Flag | Meaning |
|---|---|
| `--php 8.1` | Target PHP version; rules are gated on it. Default: `custos.json`, then composer `config.platform.php`, then the lowest version `require.php` allows, then 8.4. |
| `--rule A,B` | Run only these rules (IDs or PhpStorm inspection names). |
| `--all` | Enable every rule, including the ones off by default. |
| `--comparison-style regular\|yoda` | Preferred operand order for comparison rules and their fixes. |
| `--config DIR` | Where to look for `custos.json` (default: the first path). |
| `--exclude a,b` | Extra directories to skip. |

## analyse

| Flag | Meaning |
|---|---|
| `--format text\|json\|checkstyle\|github\|sarif` | Output format (default `text`). |
| `--fail-on info\|warning\|error\|never` | Exit 1 when a finding has at least this severity (default `warning`). |
| `--baseline FILE` | Ignore the findings recorded in this baseline (default: `baseline` in `custos.json`). |
| `--generate-baseline FILE` | Record all current findings in FILE and exit 0. |
| `--stats` | Print timing, file and rule counts on stderr. |

Output formats:

- `text`: `file:line:col: severity: message [Rule] (fixable)` and a summary line.
- `json`: `{"files": N, "findings": [...]}`, each finding with `path`, `line`,
  `column`, `endLine`, `endColumn`, `rule`, `severity`, `message` and `fixable`.
- `checkstyle`: Checkstyle XML, read by many CI dashboards.
- `github`: GitHub Actions workflow commands, shown as annotations on the
  pull request diff.
- `sarif`: SARIF 2.1.0, for GitHub code scanning and other SARIF viewers.

## fix

| Flag | Meaning |
|---|---|
| `--dry-run` | Do not write files. |
| `--diff` | Print a unified diff of the changes. |

Fixes are applied repeatedly until nothing changes (at most 10 rounds), since
one fix can reveal another finding. A file that cannot be read or written is
reported and the other files are still fixed; `fix` then exits 2. custos
never writes through symlinks.

## explain and rules

```sh
custos explain UnnecessarySemicolon     # summary, options, suppression name
custos rules                            # ✓ marks implemented rules; group, severity, default
custos rules --json                     # the full catalogue, for tooling
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success; no finding at or above `--fail-on`. |
| `1` | Findings at or above `--fail-on`. |
| `2` | Usage or configuration error (unknown flag value, bad `custos.json`…), or files `fix` could not process. |
