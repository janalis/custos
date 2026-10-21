# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- Type inference knows more values: a property or return type documented as
  a subclass of its declared type (`@var Text` on a `Message` property), the
  elements nested writes create (`$groups[$key][] = $row`), and the matches
  of `preg_match_all()` with `PREG_SET_ORDER` or `PREG_OFFSET_CAPTURE`.
- HardcodedCredentialAtKnownSink matches `excludedPaths` from the workspace
  root in the language server and the MCP server when a file has no
  `custos.json` or `composer.json` above it, instead of the server's working
  directory; over MCP it no longer looks outside the workspace for them.

### Fixed

- A crash on one file no longer ends `custos analyse`, `custos fix`, the
  language server or the MCP server: the file gets an `internal` finding and
  the `internal-failure` status, and the other files are analyzed as usual.
  A quick-fix that crashes is left out, and a failure building shared type or
  flow data skips only the checks that need it instead of being reported
  once per rule.
- Quick-fixes apply whole or not at all: a fix whose edits overlap each other
  or leave the file is skipped, and a fixing pass that would add syntax errors
  is undone.
- Taint is followed through a chain of more than eight wrapper functions
  across files. Past 32 files, calls into the chain are reported as
  `unresolved-dispatch` instead of passing as clean.
- The MCP server applies `flow.assertions` to cross-file flow summaries, like
  the CLI and the language server.

## [0.1.7] - 2026-10-21

### Added

- `custos mcp` serves project analysis, rule knowledge, semantic navigation,
  resumable checkpoints and prepared fixes to AI agents over stdio or
  authenticated Streamable HTTP, confined to workspace roots and identity
  permissions. It handles large vendor trees, filters and summarizes
  findings, and Composer's launcher points it at the consuming project.
- The language server applies the project baseline, like `custos analyse`,
  and reloads it when the file changes.
- Rule pages show a *native* badge, and large sidebar groups are split by
  initial letter.

### Changed

- Discovery skips hidden directories (`.phpunit.cache`, `.php-cs-fixer`…)
  below each scanned path.
- `--all` no longer enables rules that `custos.json` turns off; `--rule`
  still runs exactly the rules it names.
- A rule entry in the language server's `initializationOptions` overrides the
  project's entry field by field instead of replacing it.
- The language server recomputes only the flow summaries an edit can affect
  and skips analyses superseded while waiting; whole-project runs no longer
  contend on symbol lookups.
- Type inference knows many more types: callback parameters of array
  functions, inherited `@param` types, constant and literal array shapes,
  more array builtins, appended rows, conditional return types (Symfony's
  `$container->get()`, Prophecy's `reveal()`), generic constructors, traits,
  iterators, `class-string<T>`, `self`/`parent` and by-reference outputs.
- Type narrowing understands more ways of writing a check (exits, loose
  comparisons, substring and length tests, saved results, guards on values
  of unknown type), and rules that look for a guard now read it from type
  narrowing, which removes false positives across the failure-check rules,
  NullPointerException and the stream, socket and process checks.
- In checks that need a proof, a guard stops counting where the value may
  change unseen (a project call, a reference write, `eval()`, a jump label).
- Rule pages and `custos explain` read better: no unreferenced spec labels or
  boilerplate, and option descriptions say what they do.

### Fixed

- Fewer false positives in NullPointerException,
  CallableParameterUseCaseInTypeContext, MagicMethodsValidity,
  DateLastErrorsFalseUnchecked, UnlinkFailureReportedAsSuccess,
  OffsetOperations, CsvEscapeDefaultDependency,
  HeapIterationConsumesCollection and XslTransformationFailureConsumed.
- Type narrowing no longer keeps facts that stopped being true (reassigned
  properties, `static` locals, loop-carried values), no longer treats a
  namespaced `is_*()` function as PHP's, and respects the PHP version for
  `or throw`.
- Checks that need a proof reject tests that prove nothing: a comparison the
  declared type rules out, a hooked or magic property, or a property written
  through another variable.

### Security

- Require Go 1.27.2 to include current HTTP, TLS and root-confinement fixes.

## [0.1.6] - 2026-10-19

### Added

- Analysis gaps, places where a rule stopped at a work bound instead of
  answering, are reported apart from findings in JSON, SARIF, the CLI summary
  and editor logs.
- The `flow.assertions` setting (off by default) lets `assert()` calls count
  as guards, assuming assertions run and a failed one stops execution.

### Changed

- Re-analysis after an edit is faster: a file's functions are analyzed once
  for summaries and findings, and results that read no changed summary are
  kept. Repeated builtin and method lookups are cached.
- Documentation matches the binary more closely, and the rule index shows
  each rule's origin (native or EA-modeled).

### Fixed

- UnescapedHtmlOutput understands far more HTML contexts (script and style
  text, URL, event-handler, `style` and `srcdoc` attributes, unquoted and data
  attributes) and more escaping functions, follows request superglobals
  rewritten or refilled before output, reports every unsafe part of an output
  statement and highlights the exact value.
- Failure-check rules accept webmozart/assert and PHPUnit assertions and
  guards on properties, report assertions that would pass with the failure
  value, and count PHP's own methods and duplicated library functions as
  string consumers.
- The flow analysis follows `goto` through `finally`, generators across
  `yield`, constant conditions, integer ranges, loops, references, globals
  written in other functions and try blocks that cannot throw, so findings are
  both more complete and less noisy.
- Unpacked and named arguments are no longer read as omitted or misplaced by
  argument checks.
- Type inference handles more PHPDoc forms, enum cases, typed constants,
  by-reference outputs of PHP functions, property hooks and version-specific
  builtin signatures; documentation availability tags no longer restrict PHP
  versions.
- Fewer false positives in ExitStatusReadWithoutNormalExitCheck,
  NonVoidFunctionFallsThrough, ReadLoopProcessesEofFailure and the cURL and
  socket zero-result checks.
- An unterminated backtick no longer crashes fix preparation, and signature
  fixes are withheld when a class hierarchy is only partly known.

## [0.1.5] - 2026-10-14

### Added

- `IsIterableCanBeUsed` offers a quick-fix for an adjacent
  `is_array($v) || $v instanceof Traversable` pair on a plain variable.
- Security rules cover more sinks: interpolated, `print` and SVG/XHTML
  output, the backtick operator and string-form `proc_open`, remote stream
  reads, `curl_setopt_array` destinations, LDAP filters escaped for the wrong
  context, bit-or permission modes and cURL basic authentication set through
  `CURLOPT_URL`.
- The flow analysis follows functions containing `switch`, `match (true)` and
  `goto`, by-reference outputs, silenced calls, `global` variables written by
  other functions and `static` locals, so value, failure-check and
  request-data rules report in more places.

### Changed

- Project analysis is faster on large projects: summaries are computed in
  parallel, built once instead of per file, and a file is not analyzed again
  when nothing it calls changed. Very long functions and files with many
  `goto` labels no longer exhaust the analysis budget or take seconds.

### Fixed

- Quick-fixes no longer produce invalid PHP or change behavior: they add
  needed parentheses, keep comments, respect keywords, closing tags, heredocs
  and PHP-version differences, and are withheld (the finding is still
  reported) when the rewrite could duplicate, drop or reorder side effects.
- Code-style, control-flow, performance, language-migration and PHPUnit
  rewrites are withheld where the result could differ; PHPUnit rewrites
  ignore same-named methods of other classes.
- Calls using named arguments, argument unpacking or first-class callable
  syntax are no longer misread by argument-position checks.
- Variables that can change through references, `global`, `extract()`,
  variable variables, destructuring, `foreach` targets or code PHP runs
  implicitly (conversions to text, magic methods, destructors, autoloading,
  callbacks) are no longer treated as stable.
- Failure-check rules accept many more guard forms (early exits,
  `or exit`, stored tests, never-returning calls, `switch` and `match`
  allow-lists, type predicates, `isset`/`empty` on an element), and no longer
  let a check vouch for a later, unchecked assignment or for a value that a
  `goto` carries past it.
- Request-data rules track array elements, references, by-reference
  parameters, statement bindings, `extract()` and `parse_str()` more
  precisely, report request data combined with unknown values, and no longer
  report numbers, booleans or `switch`/`match`-selected values.
  `UnescapedHtmlOutput` examines every part of an output statement.
- Architecture and unused-code rules respect property hooks, abstract
  and promoted properties, trait methods, `__sleep()` lists and PHP-version
  differences of built-in classes.
- Inspections that skip unreachable code no longer skip code a `goto` label
  makes reachable, nor property hooks after a top-level `return`.
- HTTP, cookie, PDO, mysqli and SQLite rules compare names
  case-insensitively and tolerate SQL comments and state restored elsewhere.
- Malformed input no longer makes some rules skip the file.

## [0.1.4] - 2026-10-10

### Fixed

- Boolean ternary simplification now requires opposite arms, and array membership
  rewrites preserve strictness and key evaluation.
- Array deduplication and list checks require compatible values and stable
  evaluation; indexed append checks avoid sparse-array false positives.
- Substring-to-array-access fixes require safe offsets, preserve source
  evaluation, and respect PHP-version differences.
- String checks and PHP 8 string-function rewrites preserve byte semantics by
  excluding multibyte operations and mixed string units.
- Ignored immutable date results now offer an assignment fix when receiver
  ownership and the modification can be proven safe.
- Replacement backreference fixes handle all supported references in one pass
  while preserving string quoting and escaping.

## [0.1.3] - 2026-10-10

### Added

- Native PHP rules covering references, generators and fibers, arrays,
  JSON, cryptography, HTTP, cURL, databases, processes, streams, text, images
  and extension contracts. Each has an independent specification and own
  fixtures; intent-sensitive policies remain opt-in.
- Safe fixes for NaN comparisons, ZIP opening, OpenSSL verification, literal
  regex replacements, normalized query keys and sodium nonce lengths.
- Separate native rule metadata that survives upstream fact extraction; native
  rules use custos IDs and have no upstream conformance requirement.
- Bounded flow analysis with project function and method summaries, resource
  state tracking and context-specific validation; editor summaries include
  unsaved document changes.

### Changed

- Documentation diagrams now render with Mermaid.

### Fixed

- Native rules handle branch reachability, mutation, callbacks,
  namespace resolution and PHP-version constraints more accurately, reducing
  false positives and withholding fixes when their safety cannot be proven.

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

- 178 EA-modeled PHP rules in 12 groups (probable bugs, performance, security,
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
- custos is a clean-room implementation and intentionally differs from
  upstream where upstream behavior is wrong (false positives, fixes that
  change semantics or produce invalid PHP).

### Security

- Hostile input cannot crash, hang or exhaust the analyzer: nesting depth,
  file size (10 MB), syntax errors and findings per file are capped, PHPDoc
  types are parsed in linear time, and known quadratic paths were removed.
- Only regular files are read, with size caps; `paths` and `baseline` must
  stay inside the project; `custos fix` never writes through symlinks; the
  language server rejects malformed or oversized messages.

[Unreleased]: https://github.com/janalis/custos/compare/v0.1.7...HEAD
[0.1.7]: https://github.com/janalis/custos/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/janalis/custos/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/janalis/custos/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/janalis/custos/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/janalis/custos/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/janalis/custos/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/janalis/custos/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/janalis/custos/releases/tag/v0.1.0
