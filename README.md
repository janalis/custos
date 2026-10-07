# custos

A fast PHP inspector and fixer written in Go: 178 inspections with quick-fixes
covering probable bugs, performance, security, code style, control flow,
language-level migration (PHP 5.3 → 8.5), PHPUnit usage and more. Single
static binary, no PHP runtime required. Usable from the command line / CI and
as an LSP server for any editor.

custos is an independent clean-room implementation whose rule catalogue is
modelled on *Php Inspections (EA Extended)*; rule IDs are compatible, so
existing `@noinspection XxxInspection` comments keep working. See `NOTICE`.

## Install

Download the archive for your platform from the releases page (Linux, macOS
and Windows, amd64/arm64; checksums in `SHA256SUMS`), extract `custos` and put
it on your `PATH`. Or build from source (Go 1.27):

```sh
make build          # → bin/custos
```

## Usage

```sh
custos analyse [paths...]                 # report problems (exit 1 on warnings by default)
custos analyse --format=checkstyle src    # text | json | checkstyle | github
custos analyse --all --php 7.4 src        # every rule, target PHP 7.4
custos analyse --generate-baseline custos-baseline.json src   # accept current findings
custos analyse --baseline custos-baseline.json src            # report only new ones
custos fix --dry-run --diff src           # preview quick-fixes
custos fix --rule NestedNotOperators src  # apply one rule's fixes
custos rules                              # list rules (✓ = implemented)
custos explain UnnecessarySemicolon       # describe a rule and its options
custos lsp                                # language server over stdio
```

Suppress a finding with a comment before the statement or declaration:

```php
/** @noinspection UnnecessarySemicolonInspection */
// @custos-ignore UnnecessarySemicolon
```

## Configuration

`custos.json` at the project root (all keys optional):

```json
{
  "php": "8.3",
  "comparisonStyle": "yoda",
  "paths": ["src", "tests"],
  "exclude": ["var"],
  "baseline": "custos-baseline.json",
  "rules": {
    "OneTimeUseVariables": { "enabled": true, "severity": "warning",
                             "options": { "ALLOW_LONG_STATEMENTS": false } },
    "MultipleReturnStatements": { "enabled": false }
  }
}
```

Without `php`, the target version is taken from `composer.json`
(`config.platform.php`, else the lowest version allowed by `require.php`).
The same object is accepted as LSP `initializationOptions`.

## Editor integration

`custos lsp` publishes diagnostics as you type and offers quick-fixes
(`quickfix`, lazily resolved), a `source.fixAll.custos` action and the
`custos.fixFile` / `custos.fixRule` commands; run it beside your main PHP
language server. Setup for common editors: `docs/usage.md`.

## Development

```sh
make verify         # lint + tests + own fixtures + clean-room scan
make conformance    # compare against a local EA checkout (EA_PATH=…)
make bench          # parser / rule benchmarks
make stubs          # rebuild the embedded PHP stubs index
```

Docs: `docs/usage.md` (CLI, configuration, CI, editors), `docs/rules-reference.md` (every rule, its options and defaults),
`docs/migration.md` (plan & status), `docs/rules.md` (per-rule status),
`docs/decisions.md`, `specs/` (one behavioural spec per rule).

## Releasing

Update `CHANGELOG.md`, then tag: `git tag vX.Y.Z && git push --tags`. The
`release` workflow runs `make verify` and goreleaser (`.goreleaser.yaml`).
Locally: `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean`.

## License

MIT. Builtin symbol data from JetBrains phpstorm-stubs (Apache-2.0).
