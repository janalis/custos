# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Security
- Untrusted code can no longer crash, hang or exhaust memory: syntax
  nesting capped (4,000 levels); PHPDoc types, aliases and generic chains
  parsed in linear time with size caps; files over 10 MB skipped; at most
  1,000 syntax errors and 10,000 findings per file.
- Only regular files are read, with size caps (sources, `custos.json`,
  `composer.json`, baselines); `paths` and `baseline` must stay inside the
  project; `custos fix` never writes through symlinks.
- LSP: oversized or malformed `Content-Length` headers and negative fix
  indexes return errors instead of crashing the server.
- Quadratic paths removed (suppression lookup, long-line columns,
  reachability, value discovery, class hierarchy walks, narrowing, several
  rules) so large legal files stay fast.

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
- Method and function `@template` binding from call arguments
  (`getRepository(Foo::class)` → `Foo`), `class-string<Foo>` for
  `Foo::class`.
- `@phpstan-assert` / `@psalm-assert` (incl. `-if-true` / `-if-false`)
  narrowing.
- Types for private untyped properties, inferred from the class's writes.

### Fixed
- False positives in ReturnTypeCanBeDeclared, UnnecessaryCasting,
  CallableParameterUseCaseInTypeContext, OffsetOperations,
  NullPointerException, MagicMethodsValidity, MockingMethodsCorrectness,
  LoopWhichDoesNotLoop, NotOptimalRegularExpressions,
  ClassOverridesFieldOfSuperClass, PropertyInitializationFlaws.
- Unsafe quick-fixes: nullable return types where values can be null,
  preload/autoload requires kept, `@mkdir` kept, typed property defaults
  kept; regex-to-string-function rewrites only when exactly equivalent
  (`$` and trailing newlines, `\s` vs `trim()`, modifiers).
- Parser: `if ($a) ?>html` (close tag as an empty body), `readonly(` as a
  call on PHP 8.2+, `new` without parentheses, `break N` levels; about 50
  further rule, engine, CLI and LSP bugs found by the 100% coverage work
  (see `docs/decisions.md`).
- CLI: unknown `--fail-on` / `--format` values are usage errors (a typo
  no longer disables the CI gate); `fix` continues past unwritable files.
- LSP: `custos.json` `paths`/`exclude`/`baseline` honoured when the editor
  sends settings; crashing requests return an error instead of success.

### Notes
- custos intentionally diverges from upstream where upstream behaviour is
  wrong (fixes that change semantics or produce invalid PHP, false positives);
  see `docs/decisions.md`.
