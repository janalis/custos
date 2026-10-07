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
- Closures and arrow functions carry their return type (declared, or
  inferred from the body); `$f()`, `call_user_func()`, invokable objects,
  `callable(…): R` / `Closure(…): R` doc types and `array_map()` results
  are typed from it; `array_filter()` without callback drops null/false.
- Out parameters: `@param-out` (phpstan-/psalm- variants) and builtin
  outputs (`preg_match()` `$matches`, `exec()` `$output`, `parse_str()`,
  `str_replace()` count, …) type the variable after the call.
- Conditional return types (`@return ($x is string ? int : float)`,
  class-constant and literal targets, template subjects) resolved from the
  call's arguments (also PDOStatement::fetch()/fetchAll() modes).
- `$this->prop` is non-null after a non-null assignment in the same method
  until a call may reset it; writing an element into a nullable array
  (`?array $n; $n['k'] = 1;`) drops null.
- LSP: "Suppress <Rule> for this statement" code action; saving a file
  re-checks the other open files.

### Changed
- `custos fix` fixes files in parallel (WordPress 7.1 s → 1.6 s);
  `--stats` labels a `--php` version as "flag".
- `make fixcheck` also applies all fixes of each file together.

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
- `custos fix`: fixes of different rules no longer combine into invalid
  PHP (an insertion such as UnqualifiedReference's `\` touching another
  fix's rewrite is deferred to the next pass); 569 of 9,138 files fixed on
  WordPress, Laravel, Drupal and Nextcloud were broken before.
- Parser: `$a = &f() && $b` binds as `($a = &f()) && $b`.
- Unsafe or invalid quick-fixes (WordPress / Laravel / Drupal / Nextcloud
  review): MagicMethodsValidity `_set` → `__set` rename,
  ReturnTypeCanBeDeclared on PHP 4 constructors, `__serialize()` and
  unknown return values, StaticInvocationViaThis `self::` (now `static::`),
  TypeUnsafeComparison `===` on possibly numeric/bool operands,
  ClassConstantCanBeUsed on `'\A\B'`, UsingInclusionOnceReturnValue
  (plain include re-ran the file), IsEmptyFunctionUsage on nullable
  scalars, SenselessProxyMethod next to PHP 4 constructors,
  ForeachInvariants on mutated arrays, MultiAssignmentUsage,
  ReferencingObjects on overriding methods, ComparisonOperandsOrder with
  assignments, OpAssignShortSyntax with side effects,
  NotOptimalRegularExpressions value rewrites, SubStrShortHandUsage,
  ArrayIsListCanBeUsed on `[]`.
- False positives in OnlyWritesOnParameter (`global`), ForgottenDebugOutput
  (`wp_die` no longer a default), SuspiciousAssignments,
  MissingIssetImplementation, UnsupportedStringOffsetOperations,
  AlterInForeach, IssetArgumentExistence, SlowArrayOperationsInLoop,
  PregQuoteUsage, MockingMethodsCorrectness, PhpUnitDeprecations,
  PassingByReferenceCorrectness, ClassConstantUsageCorrectness,
  ProperNullCoalescingOperatorUsage, CallableParameterUseCaseInTypeContext,
  UnserializeExploits, EncryptionInitializationVectorRandomness,
  CryptographicallySecureRandomness, SecurityAdvisories (metapackages),
  DisconnectedForeachInstruction, DisallowWritingIntoStaticProperties,
  UnqualifiedReference, EmptyClass, InvertedIfElseConstructs,
  NullPointerException.

### Notes
- custos intentionally diverges from upstream where upstream behaviour is
  wrong (fixes that change semantics or produce invalid PHP, false positives);
  see `docs/decisions.md`.
