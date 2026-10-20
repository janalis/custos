# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `custos mcp` exposes project analysis, rule knowledge, semantic navigation,
  resumable checkpoints and prepared fixes over stdio and authenticated
  Streamable HTTP. Workspace roots and identity permissions constrain access.
- Composer's launcher supplies the consuming project as the MCP default,
  including custom vendor/bin directories and symlinked package installations.
- MCP analysis handles large dependency trees: vendor files are indexed without
  being retained (up to 100,000 files), vendor symlinks are skipped and
  reported instead of failing, and unchanged dependencies are not parsed again.
- MCP findings filter by directory and case-insensitive rule ID and summarize
  counts by severity, rule and file; symbol search ranks exact names first and
  returns PHP-style signatures; types can be queried by line and column.
- The language server applies the project baseline, like `custos analyse`,
  and reloads it when the file changes.
- Native rule pages carry a *native* badge, and the largest rule group of the
  sidebar is split by initial letter.

### Changed

- Discovery skips hidden directories (`.phpunit.cache`, `.php-cs-fixer`, and
  any other directory whose name starts with a dot) below each scanned path,
  so generated code kept there is no longer analyzed. A scanned path may still
  name one.
- `--all` no longer enables rules that `custos.json` turns off with
  `"enabled": false`. `--rule` still runs exactly the rules it names.
- A `rules` entry in the language server's `initializationOptions` overrides
  the project's entry field by field instead of replacing it: setting only a
  severity keeps the project's `enabled` and options.
- Rule pages and `custos explain` read better: spec item labels (D1, E2…)
  only appear where the text refers to them, references to items a page does
  not show and boilerplate sentences are gone, option descriptions say what
  they do instead of naming spec items, and notes about upstream test
  configurations are dropped.
- The rule status table and the rule reference agree on which rules have a
  quick-fix.
- Symbol lookups no longer contend on a lock once the project index is built,
  which shortens whole-project runs.
- The language server no longer rebuilds the whole project's flow summaries
  on every edit: an edit that leaves a file's declarations unchanged
  recomputes only the summaries it can affect, and other open files are
  analyzed again only when a summary changed. An analysis that waited for a
  free slot is skipped if the file changed meanwhile.
- Type inference knows more types:
  - closures passed to `array_map()`, `array_filter()`, `array_walk()`,
    `array_reduce()`, `usort()`, `uasort()` and `uksort()` get their untyped
    parameters typed from the array's elements or keys, and parameters
    declared `array` or `iterable` keep the elements' array shape;
  - a method parameter without its own `@param` takes the documented type of
    the method it overrides or implements (a Symfony form type's
    `buildForm()` `$options`);
  - constants initialized with an array literal keep its keys and value types
    (except where `static::` may read an override); a computed-key read of a
    literal or constant array gives the union of its values; destructuring a
    typed array types each variable; a promoted parameter's `@var` documents
    the property;
  - more array builtins are typed from their arguments (`array_merge()`,
    `array_replace()`, `array_chunk()`, `array_combine()`, `array_column()`,
    `array_fill*()`, `array_pad()`, `array_keys()`, `array_flip()`,
    `array_count_values()`, `array_intersect*()`/`array_diff*()`,
    `iterator_to_array()`, integer `range()`, `array_values()` and its
    siblings), `pathinfo()` and `parse_url()` return shapes, `fgetcsv()`,
    `str_getcsv()`, `file()` and `glob()` return string lists, and
    `preg_split()` with `PREG_SPLIT_OFFSET_CAPTURE` returns `[string, int]`
    pairs;
  - an array filled by appending shaped rows keeps the rows' shape, and
    passing an array by reference (`sort()`, `usort()`…) no longer drops its
    shape for reads before the call;
  - conditional return types accept a union of literals or constants as the
    tested value, read a parameter default given as a class constant and
    bind the method's templates from the call: Symfony's
    `$container->get('id')` is a non-null object (no NullPointerException on
    `$container->get('id')->method()`), and Prophecy's
    `prophesize(Foo::class)->reveal()` is a `Foo`;
  - a function or method declared `array` without `@return` returns the array
    type its body returns, where no override can run instead;
  - arithmetic with one numeric operand is a number even when the other is
    unknown (`$totals['count'] += $row['n']`), and arithmetic on `int`,
    `bool` and `null` operands is an `int` (UnnecessaryCasting now reports
    `(int)` on `strtotime($d) + 60`);
  - a variable passed by reference without `@param-out` takes the
    parameter's `@param` type after the call (checks that need a proof still
    treat it as unknown);
  - `self` and `parent` in parameter types and inline `@var` comments name
    the enclosing class;
  - a string that passes `is_a($x, Foo::class, true)` or `is_subclass_of()`
    is a `class-string<Foo>`, not a `Foo` instance;
  - `new` binds a generic class's template arguments from its constructor's
    documented parameters (`new ArrayCollection($items)`);
  - inside a trait, `$this->property` and `$this->method()` take the types
    the trait declares (except in checks that need a proof);
  - iterators without generic annotations type their keys and values from
    `getIterator()`, `key()` and `current()`, so `DirectoryIterator` entries
    are no longer reported by TypeUnsafeComparison when compared to a string;
  - `class-string<Foo>` types `new $class`, `$class::method()` and
    `$value instanceof $class` as `Foo`;
  - a list built with `$list[] = …` types `$list[0]` as the appended values.
- Type narrowing understands more ways of writing a check:
  - a value of unknown type takes the concrete type its guards prove
    (`is_string($x)`, `$x instanceof Foo`, `Assert::string($x)`), so
    StrlenInEmptyStringCheckContext no longer suggests a redundant `(string)`
    after `is_string($x) &&`;
  - code after an `if` that leaves through every branch, a `never` call,
    `$x or exit;`, `$x || throw …;`, `$x ?: throw …;` or a ternary with an
    exit; code after a loop that only its condition ends; statements at file
    level;
  - loose comparisons (`$x == false`, `!= null`, `!= ''`, `!= 0`, `!= '0'`),
    silenced tests (`@$x`), truthiness of static properties, `===` against a
    value that is never null, `elseif` chains and the right side of `?:`;
  - substring searches (`str_contains()`, `strpos(...) !== false` and the
    like), `strlen()` comparisons, `array_key_exists()` with any key and
    `is_array($x) || isset($x['k'])`;
  - nested literal array keys, array elements named by an unchanged local key
    (`$_FILES[$field]`), booleans saved from a test of a property or an
    element, and copies of a tested value.
- Rules that check for a guard now read it from type narrowing instead of
  matching a few spellings, so saved test results, loose comparisons,
  `elseif`, ternary, `match (true)` and `&&`/`||` positions, `or exit;` and
  an exit followed by `else` count as guards, which removes false
  positives. This covers the `*FailureUnchecked` family and
  AeadDecryptionTagLengthUnchecked, AppendModeSeekUsedForOverwrite,
  ConnectTimeoutAssumedToBoundReads, DateLastErrorsFalseUnchecked,
  DecryptionFailureUsedAsPlaintext, DirectoryIteratorDotEntries,
  EmptySplCollectionExtraction, ExitStatusReadWithoutNormalExitCheck,
  FailedClosureBindingInvoked, FilePermissionFailurePassesPolicy,
  ForkFailureTreatedAsParent, HashFileFailureUsedAsDigest,
  HeadersAfterCommittedOutput, ImagickSingleBlobDropsRequiredFrames,
  LocalizedNumberTrailingInputAccepted,
  MysqliEscapedValueSurvivesCharsetChange, MysqliMultiQueryResultsNotDrained,
  NonblockingSocketConnectPendingRejected, NonExhaustiveEnumMatch,
  NullPointerException, OpenSslSigningFailureIgnored,
  OptionalRegexCaptureReadWithoutNullFlag, ParseUrlFailureDereferenced,
  PdoFetchIntoRetainsSharedRows, PgReturnedRowsUsedAsAffectedRows,
  PosixAccountLookupFailureDereferenced, ReadFromWriteOnlyStream,
  RealpathFailureUsedAsPath, RedirectContinuesProtectedExecution,
  ReflectionCompositeTypeAssumedNamed, SelectFailureTreatedAsReadiness,
  SessionCookieOptionsSetAfterStart, ShortStreamReadUnchecked,
  SignalWaitFailureAcceptedAsSignal, SqliteForeignKeySettingInsideTransaction,
  StreamUseAfterClose, SysvMessageSerializationMismatch,
  TempnamDirectoryFallbackUnchecked, TlsHandshakePendingAccepted,
  TruncationAssumedToRewind, UntrustedFilesystemPath, UntrustedHeaderValue,
  UploadClientMimeTrusted, WriteToReadOnlyStream and
  XmlWriterMemoryOutputInUriMode.
- In checks that need a proof, a guard stops counting where the value may
  change unseen: after a call of project code, a write through a reference,
  a call that may take the value by reference, `eval()` (now a dynamic write
  like `include`), or a jump label between the test and the use. For
  example, UploadClientMimeTrusted reports a move after a project call that
  follows the client type test.

### Fixed

- NullPointerException no longer reports a local assigned a call that an
  identical guarding call was tested for null, nor a local of undeclared type
  (`$row = $rows[$i] ?? null;`) that an enclosing condition proves non-null
  (`Kind::A === $row?->kind`, an earlier `$row->p` in the same `&&`).
- CallableParameterUseCaseInTypeContext no longer reports a local assigned to
  a parameter when it holds a call result that may be `false` or carries an
  inline `@var` tag.
- MagicMethodsValidity no longer reports a constructor that calls the parent
  constructor through a helper no subclass can replace.
- DateLastErrorsFalseUnchecked no longer reports `getLastErrors()` read right
  after a date parse failed.
- UnlinkFailureReportedAsSuccess no longer reports a function that deletes
  the scratch file it made with `tempnam()` before returning `true`.
- OffsetOperations no longer reports an array key checked with
  `is_a(..., true)` or `is_subclass_of()`.
- Type narrowing no longer keeps a fact that stopped being true: about a
  property reassigned inside the guarded branch, a `static` local after a
  call that may run the function again, a value carried to the next loop
  iteration, a value assigned right before `continue 2`, or a variable
  assigned in a later condition than the guard.
- A namespaced function named like a type check (`App\is_string()`) no longer
  narrows its argument as PHP's own check would, and `$x or throw …;` no
  longer narrows `$x` for PHP versions before 8.0.
- Checks that need a proof treat a strict comparison with a constant the
  declared type rules out (`$flag === 1` for a `bool`) as a branch that never
  runs, not as a proof, and no longer accept a test of a hooked or magic
  property, or of a property written through another variable.
- CsvEscapeDefaultDependency and HeapIterationConsumesCollection no longer
  report after `eval()`, and XslTransformationFailureConsumed no longer
  reports output guarded by a silenced test.

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

[Unreleased]: https://github.com/janalis/custos/compare/v0.1.6...HEAD
[0.1.6]: https://github.com/janalis/custos/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/janalis/custos/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/janalis/custos/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/janalis/custos/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/janalis/custos/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/janalis/custos/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/janalis/custos/releases/tag/v0.1.0
