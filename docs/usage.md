# Using custos

## Command line

```sh
custos analyse [flags] [paths...]
custos fix [flags] [paths...]
custos explain <rule>
custos rules [--json]
custos lsp
```

Paths default to `paths` from `custos.json`, else the project root (the
directory holding `custos.json` or `composer.json`). `vendor`, `node_modules`,
`.git`, `.idea`, `.custos` and `var/cache` are skipped; add more with
`exclude` in `custos.json` or `--exclude a,b`.

Common flags (`analyse` and `fix`):

| Flag | Meaning |
|---|---|
| `--php 8.1` | Target PHP version (rules are gated on it). Default: `custos.json` → composer `config.platform.php` → lowest `require.php` → 8.4. |
| `--rule A,B` | Run only these rules (IDs or PhpStorm inspection names). |
| `--all` | Enable every rule, including the ones off by default. |
| `--comparison-style regular\|yoda` | Preferred operand order for comparison rules and fixes. |
| `--config DIR` | Where to look for `custos.json`. |
| `--exclude a,b` | Extra directories to skip. |

`analyse` only:

| Flag | Meaning |
|---|---|
| `--format text\|json\|checkstyle\|github\|sarif` | Output format. |
| `--fail-on info\|warning\|error\|never` | Exit 1 when a finding has at least this severity (default `warning`). |
| `--baseline FILE` / `--generate-baseline FILE` | Suppress known findings / record current ones. |
| `--stats` | Timing and rule count on stderr. |

`fix` only: `--dry-run` (do not write) and `--diff` (print a unified diff).
Fixes are applied repeatedly until nothing changes (at most 10 rounds).
Symlinks and other non-regular files are skipped (with a note on stderr);
a file that cannot be read or written is reported and the other files are
still fixed, then `fix` exits 2.

Exit codes: `0` success / no finding at or above `--fail-on`; `1` findings at
or above `--fail-on`; `2` usage or configuration error (including an unknown
`--format` or `--fail-on` value), or files `fix` could not process.

## Configuration (`custos.json`)

```json
{
  "php": "8.2",
  "comparisonStyle": "yoda",
  "shortOpenTag": false,
  "paths": ["src", "tests"],
  "exclude": ["var", "public/build"],
  "baseline": "custos-baseline.json",
  "rules": {
    "MultipleReturnStatements": { "enabled": false },
    "OneTimeUseVariables": { "enabled": true, "severity": "warning",
                             "options": { "ALLOW_LONG_STATEMENTS": false } },
    "ForgottenDebugOutput": { "options": { "configuration": ["\\App\\Debug::dump"] } }
  }
}
```

Rule options and defaults are listed in `docs/rules-reference.md` and by
`custos explain <rule>`. Unknown rule IDs are a configuration error.

## Suppressing findings

- Before a statement or declaration: `/** @noinspection RuleIdInspection */`
  (PhpStorm-compatible, so existing comments keep working) or
  `// @custos-ignore RuleId`. Several names may be listed; `ALL` suppresses
  every rule.
- Before the first statement of a file: applies to the whole file.
- Legacy code: generate a baseline once and commit it
  (`custos analyse --generate-baseline custos-baseline.json`); afterwards only
  new findings are reported. Baseline entries survive code moving around (they
  match on file, rule, message and the flagged line's text).

## CI

### GitHub Actions (with code scanning)

```yaml
# custos installed with `composer require --dev janalis/custos`
- name: custos
  run: vendor/bin/custos analyse --format=sarif > custos.sarif || true
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: custos.sarif
- name: custos (gate)
  run: vendor/bin/custos analyse --format=github --fail-on=warning
```

`--format=github` prints workflow annotations inline on the PR diff.

### GitLab / other CI

```sh
custos analyse --format=checkstyle > custos-checkstyle.xml   # for checkstyle-aware viewers
custos analyse --fail-on=error                               # gate on errors only
```

## Editors (LSP)

`custos lsp` speaks LSP over stdio: diagnostics as you type (push), quick-fixes
(`quickfix`, lazily resolved when the client supports it), a
"Suppress <Rule> for this statement" action (inserts `// @custos-ignore
<Rule>` above the statement; offered only when it silences exactly that
finding, never at file level), a `source.fixAll.custos` action and the
commands `custos.fixFile` / `custos.fixRule`. Saving a file updates the
project index and re-checks the other open files. Settings are read from `custos.json` at the workspace root;
editors may override them with `initializationOptions` (same JSON shape). Run
it beside your main PHP language server.

### Neovim (0.11+)

```lua
vim.lsp.config('custos', {
  cmd = { 'custos', 'lsp' },
  filetypes = { 'php' },
  root_markers = { 'custos.json', 'composer.json', '.git' },
})
vim.lsp.enable('custos')
```

### Helix (`languages.toml`)

```toml
[language-server.custos]
command = "custos"
args = ["lsp"]

[[language]]
name = "php"
language-servers = ["intelephense", "custos"]   # keep your main server first
```

### VS Code

Use any generic LSP client extension and point it at `custos lsp` for the
`php` language, or wrap it in a minimal extension using
`vscode-languageclient` with `command: "custos", args: ["lsp"]`.

### PhpStorm

Through the LSP4IJ plugin: add a language server with command `custos lsp`
mapped to PHP files.
