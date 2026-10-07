# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- 178 PHP inspections (probable bugs, performance, security, code style,
  control flow, language-level migration 5.3 → 8.5, PHPUnit, …) with 111
  quick-fixes; rule IDs compatible with Php Inspections (EA Extended), so
  existing `@noinspection` comments keep working. See `docs/rules-reference.md`.
- Hand-written, version-aware PHP 5.3–8.5 parser (lossless, error tolerant,
  adaptive fallback to the newest grammar).
- Symbol index with embedded JetBrains phpstorm-stubs, type inference with
  narrowing, PHPDoc templates/aliases/shapes.
- CLI: `analyse` (text, json, checkstyle, github, sarif; `--fail-on`;
  baselines), `fix` (`--dry-run`, `--diff`), `rules`, `explain`, `lsp`.
- LSP server: push diagnostics, quick-fixes with lazy resolve, fix-all,
  `custos.fixFile` / `custos.fixRule`, background project index kept current
  through file watching.
- Configuration via `custos.json`; PHP target from composer.json.

### Notes
- custos intentionally diverges from upstream where upstream behaviour is
  wrong (fixes that change semantics or produce invalid PHP, false positives);
  see `docs/decisions.md`.
