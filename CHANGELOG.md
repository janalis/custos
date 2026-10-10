# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- 200 native PHP inspections, independently documented with own fixtures and
  quick-fixes where a safe correction can be determined.
- An additional 100 rules cover references, execution state, arrays, JSON,
  cryptography, HTTP, cURL, databases, processes and extension contracts;
  intent-sensitive policies remain opt-in.
- Safe fixes for NaN comparisons, ZIP opening, OpenSSL verification, literal
  regex replacements, normalized query keys and sodium nonce lengths.
- Separate native rule metadata that survives upstream fact extraction; native
  rules use custos IDs and have no upstream conformance requirement.
- Bounded flow analysis with project function and method summaries, resource
  state tracking and context-specific validation; editor summaries include
  unsaved document changes.

## [0.1.2] - 2026-10-10

### Fixed

- More accurate PHP parsing and recovery from invalid syntax.
- Better type inference for generics, arrays, callbacks, generators and
  nullsafe expressions.
- More accurate declaration and property resolution, reducing false positives.
- GitHub releases now include the release notes from this changelog.

## [0.1.1] - 2026-10-08

### Changed

- A suppression comment right before a call argument or an array item now
  applies to that argument or item, so one line of a multi-line call or
  array can be suppressed without silencing the whole statement.

### Fixed

- Parser: in permissive mode, a property or promoted-parameter default
  followed by hooks is no longer read as a legacy `$a{0}` offset.
- JsonEncodingApiUsage fixes keep named arguments (appending `flags:` or
  `associative:` by name), extend a named `flags:` value instead of giving
  up, parenthesise low-precedence flags (`$p ? A : B` no longer becomes
  `(JSON_THROW_ON_ERROR | $p) ? A : B`), and write `\JSON_THROW_ON_ERROR`
  when the file qualifies its global constants.
- MultiAssignmentUsage: consecutive offset reads from a string
  (`$a = $s[0]; $b = $s[1];`) are no longer told to destructure, which
  would assign `null` from a string.
- NotOptimalIfConditions: filesystem calls (`is_file()`, `is_dir()`,
  `file_exists()`, …) are no longer considered cheaper than in-memory checks;
  an operand is no longer moved ahead of a neighbour that narrows one of
  its variables (null/false comparison, `instanceof`, `is_*()`); property
  reads that run code (a PHP 8.4 `get` hook, a virtual, abstract or
  interface property, `__get()`) cost like a method call and are never
  reordered; reads on an unknown receiver are left out of the comparison.
- NotOptimalRegularExpressions: an escaped backslash before `.` or `$`
  (`\\\\.` in a single-quoted pattern) no longer hides the metacharacter,
  so `/s` and `/D` are not called pointless when they matter.
- NullPointerException: a named argument is checked against the parameter
  of that name; a nullsafe check followed by an early exit, a `match (true)`
  arm guarded by a nullsafe check, and a loop condition re-checked after a
  `continue` now narrow the variable; a property write no longer undoes a
  null check.
- PhpUnitTests: `assertTrue(is_resource($h))` / `assertFalse(is_resource($h))`
  are no longer rewritten to `assertIsResource()` / `assertIsNotResource()`,
  which treat a closed resource as a resource; `empty()` checks become
  `assertEmpty()`/`assertNotEmpty()` only when the operand is known to be a
  scalar, an array or null (PHPUnit counts `Countable` objects), and
  `assertTrue(!empty($x))` names the assertion its fix writes.
- PropertyInitializationFlaws: a property default equal to the default of
  the constructor parameter assigned to it is kept (objects built without
  the constructor, such as old serialized messages, still see it).
- ProperNullCoalescingOperatorUsage: a scalar fallback of another scalar
  type (`$id ?? 'new'`) and an array fallback for an iterable object
  (`$node->attributes ?? []`, a Doctrine `Collection`) are no longer
  reported.
- SlowArrayOperationsInLoop: a `for` condition measuring a value the loop
  body visibly writes (element assignment, `[] =`, `array_push()`, `unset()`,
  reassignment) is no longer reported: its length changes on purpose.
- StaticInvocationViaThis: with `EXCEPT_PHPUNIT_ASSERTIONS`, an abstract
  restatement of a PHPUnit assertion (a trait's
  `abstract public static function assertIsResource(…)`) called through
  `$this` is exempt like PHPUnit's own assertions.
- TraitsPropertiesConflicts: a trait property that re-declares a parent
  property only to attach attributes is no longer reported; a hooked
  property composed with a same-named trait property is an error, while
  the parent's hooks are ignored when a trait re-declares its property.
- UnnecessaryAssertion: a value reached through a nullsafe chain
  (`$r->find()?->sidebar()`) is no longer said to be guaranteed by the last
  call's declared return type: the chain can yield `null`.
- UnnecessaryCasting: a `(string)` cast in a concatenation is reported
  only when its operand is known to be a string, int or float, not on
  nullable, `bool`, `mixed` or `__toString()` operands.
- VariableFunctionsUsage: `call_user_func('name', …)` is kept when a
  same-named namespace function or a `use function` shadows the global
  function at the call (a wrapper reaching the builtin).

## [0.1.0] - 2026-10-08

First public release.

### Added

- 178 PHP inspections in 12 groups (probable bugs, performance, security,
  control flow, code style, unused code, PHPUnit, language-level migration
  5.3 → 8.5, …), 108 of them with quick-fixes. Rules follow the target PHP
  version. Rule IDs are compatible with Php Inspections (EA Extended), so
  existing `@noinspection` comments keep working; `@custos-ignore` works
  too. See the [rule reference](https://janalis.github.io/custos/rules/).
- CLI: `analyse` (text, json, checkstyle, github and sarif output;
  `--fail-on`; baselines for legacy code), `fix` (`--dry-run`, `--diff`,
  parallel), `rules`, `explain`, `lsp`, `version`.
- Language server: diagnostics, quick-fixes, fix-all, "suppress for this
  statement" code action, and a background project index kept current
  through file watching. Setup guides for Neovim, Helix, VS Code, PhpStorm,
  Sublime Text and Emacs.
- Configuration through an optional `custos.json` (paths, rules and their
  options, baseline); the PHP target falls back to `composer.json`.
- Own PHP 5.3–8.5 parser (lossless, error tolerant) and semantic layer: a
  symbol index with embedded JetBrains phpstorm-stubs, type inference with
  narrowing, and PHPDoc support (generics and `@template`, type aliases,
  array shapes, conditional return types, `@param-out`, phpstan/psalm
  assertions).
- Distribution: release archives for Linux, macOS and Windows (amd64,
  arm64), a Homebrew cask (`brew install janalis/tap/custos`) and a Composer
  package (`composer require --dev janalis/custos`) whose launcher downloads
  the checksum-verified binary.
- Documentation site: <https://janalis.github.io/custos/> (user guide, one
  page per rule with examples and quick-fix results, contributor guide).

### Security

- Hostile input cannot crash, hang or exhaust the analyser: nesting depth,
  file size (10 MB), syntax errors and findings per file are capped, PHPDoc
  types are parsed in linear time, and known quadratic paths were removed.
- Only regular files are read, with size caps; `paths` and `baseline` must
  stay inside the project; `custos fix` never writes through symlinks; the
  language server rejects malformed or oversized messages.

### Notes

- custos is a clean-room implementation and intentionally differs from
  upstream where upstream behaviour is wrong (false positives, fixes that
  change semantics or produce invalid PHP).
