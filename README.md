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

```sh
brew install janalis/tap/custos           # macOS / Linux (Homebrew cask)
composer require --dev janalis/custos     # per project → vendor/bin/custos
```

The Composer package is a small PHP launcher: on first run it downloads the
binary of its version for your platform from the GitHub release (checked
against `SHA256SUMS`) and caches it in the package directory. Set
`CUSTOS_DOWNLOAD_URL` to a mirror (URL or directory holding the release
assets), or `CUSTOS_BINARY` to an existing binary to skip the download.

Or download the archive for your platform from the
[releases page](https://github.com/janalis/custos/releases) (Linux, macOS and
Windows, amd64/arm64; checksums in `SHA256SUMS`), extract `custos` and put it
on your `PATH`. Or build from source (Go 1.27):

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

Write the changes under `## [Unreleased]` in `CHANGELOG.md` as you go, then
run **Actions → release → Run workflow** on `main` with a version (`X.Y.Z`,
or `patch` / `minor` / `major` to bump the latest tag; tick *dry run* to build
without publishing). The workflow:

1. runs `make verify`;
2. `tools/relprep`: turns `[Unreleased]` into `[X.Y.Z] - <date>`, pins the
   version in the Composer launcher (`composer/custos`), extracts the release
   notes; commits `Release vX.Y.Z`, tags and pushes;
3. goreleaser (`.goreleaser.yaml`): archives + bare binaries + `SHA256SUMS`
   on the GitHub release, and the cask in `janalis/homebrew-tap`;
4. Packagist picks up the tag (and is pinged when its secrets are set).

A failed publish can be retried with *Re-run failed jobs* (the tag is kept).
Locally: `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean --skip=publish`.

One-time setup:
- create the empty `janalis/homebrew-tap` repository and the
  `HOMEBREW_TAP_TOKEN` secret (fine-grained token, *Contents: read and write*
  on that repository only);
- submit `https://github.com/janalis/custos` on packagist.org (its GitHub
  hook updates on every tag); optionally add `PACKAGIST_USERNAME` and
  `PACKAGIST_TOKEN` secrets;
- if `main` is protected, let GitHub Actions bypass it so the release commit
  can be pushed.

## License

MIT. Builtin symbol data from JetBrains phpstorm-stubs (Apache-2.0).
