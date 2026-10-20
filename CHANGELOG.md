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

### Changed

- Symbol lookups no longer contend on a lock once the project index is built,
  which shortens whole-project runs.
- Rule option descriptions in `custos explain` and the rule reference no longer
  include notes about upstream test configurations.
- Type inference narrows past `if` blocks that leave through every branch of a
  nested `if`/`else`, `try` or `switch`, or through a call to a `never`
  function; past loops that only their condition ends; and on nested literal
  array keys (`$a['x']['y']`).
- Closures passed to `array_map()`, `array_filter()`, `array_walk()`,
  `array_reduce()`, `usort()`, `uasort()` and `uksort()` get their untyped
  parameters typed from the array's elements or keys; parameters declared
  `array` or `iterable` keep the passed elements' array shape.
- A method parameter without its own `@param` takes the documented type of
  the same parameter of the method it overrides or implements (for example
  the `$options` array of a Symfony form type's `buildForm()`).
- `new` gives a generic class the template arguments its constructor's
  documented parameters bind (`new ArrayCollection($items)`), so iterating or
  reading the new object keeps the element type.
- Inside a trait, `$this->property` takes the type the trait declares for it,
  and `$this->method()` the return type the trait declares (except in the
  checks that need a proof, since a using class may replace the method).
- `foreach` over an iterator class without generic annotations types its keys
  and values from `getIterator()`, or from `key()` and `current()`; for
  example, `DirectoryIterator` entries are no longer reported by
  TypeUnsafeComparison when compared to a string.
- A `class-string<Foo>` value types `new $class`, `$class::method()` and a
  successful `$value instanceof $class` as `Foo`.
- A list built with `$list[] = …` from an empty array types its integer-key
  reads (`$list[0]`) as the appended values.
- Loose comparisons narrow types: `$x == false` and `$x != true` test
  truthiness, and `$x != null`, `!= ''`, `!= 0` exclude both null and false
  (`!= '0'` excludes false).
- Failure checks of the `*FailureUnchecked` family read a loose comparison of
  the result with `''`, `0`, `0.0` or `'0'` (and `== true`) as a test of the
  failure value, also for `static` locals.
- Truthiness tests narrow static properties (`if (self::$cache) { … }`).
- Comparing a value with `===` to an expression whose type is known and never
  null, or with `==` to an object, rules null out (`if ($node === $root)`).
- PregReplaceFailureUnchecked accepts a null test from any earlier condition
  of an `if`/`elseif` chain for a result kept in a property, not only the
  first one.
- DateLastErrorsFalseUnchecked no longer reports a diagnostics read on a
  property inside a branch, ternary arm or `else` that a `false` test guards.
- NullPointerException reads its guards from type narrowing: a dereference
  is no longer reported after a test with the assignment written inside it
  (`if (null === $x = f()) { return; }`), after `continue 2` back to the
  guarding loop, or in a `match` arm after one that lists `null`.
- FailedClosureBindingInvoked reads guards of a binding kept in a property
  through type narrowing, which also accepts an exit with an `else` branch
  and a predicate compared with `true`.
- PosixAccountLookupFailureDereferenced reads its guards from type
  narrowing, and trusts a property test on a class that is not final.
- ImageDecodeFailureUnchecked reads its guards from type narrowing.
- ParseUrlFailureDereferenced reads its guards from type narrowing, and
  trusts a property test on a class that is not final.
- IntlFormattingFailureUnchecked reads its guards from type narrowing.
- RealpathFailureUsedAsPath reads its guards from type narrowing, including
  a test between the path join and the call that consumes it.
- FilePermissionFailurePassesPolicy reads its guards from type narrowing.
- SelectFailureTreatedAsReadiness, DecryptionFailureUsedAsPlaintext,
  PgReturnedRowsUsedAsAffectedRows, the comparison check of
  HashFileFailureUsedAsDigest and the stream-mode rules
  (WriteToReadOnlyStream, ReadFromWriteOnlyStream, TruncationAssumedToRewind,
  AppendModeSeekUsedForOverwrite) read failure tests from type narrowing:
  loose comparisons, truthiness and type checks count, not only `=== false`.
- The checks that need a stream or socket call to have succeeded
  (AppendModeSeekUsedForOverwrite, TruncationAssumedToRewind,
  NonblockingSocketConnectPendingRejected, SysvMessageSerializationMismatch
  and the nonblocking-stream checks) read its success from type narrowing:
  `call() or exit;`, a result kept in a boolean and tested later, `elseif`
  and ternary arms now count. A seek counts as successful wherever its
  result is known to be 0: `if (!fseek($h, 0))`, or after leaving on
  `fseek(...) === -1` or `fseek(...) < 0`.
- FileLockFailureUnchecked and StreamUseAfterClose read the result of
  `flock()` and `fclose()` through type narrowing, which also follows
  `elseif` chains, `match` and `switch` arms.
- TlsHandshakePendingAccepted accepts a completion test on a stored
  handshake result, `… === true or exit;` and an `if (… === true)` body.
- ForkFailureTreatedAsParent reads the fork result through integer
  narrowing: a test inside `if ($pid !== -1)` or `if ($pid >= 0)` no
  longer reports.
- SignalWaitFailureAcceptedAsSignal reads the wait result through integer
  narrowing, which also follows `while` conditions.
- ExitStatusReadWithoutNormalExitCheck reads the wait result through integer
  narrowing: after a blocking wait, which never returns 0, `$r >= 0` and
  `$r > -1` prove the wait succeeded.
- AeadDecryptionTagLengthUnchecked reads the tag length from type
  narrowing: a loose `strlen()` comparison, a test inside a loop that never
  writes the tag and any guarding position now count; a call of the project
  between the test and a request tag ends the proof.
- ShortStreamReadUnchecked reads the record length from type narrowing:
  `isset($d[3], $other)`, an exit followed by `else`/`elseif`, an `elseif`
  condition after a failed length test and `while` conditions now count,
  and a decode in a branch the length tests contradict is not reported.
- UntrustedFilesystemPath reads directory prefix tests from type narrowing:
  a saved test, ternary positions and an exit followed by `else` now count.
- UntrustedHeaderValue reads line-break rejections from type narrowing: a
  saved test result, ternary, `match (true)` and `&&`/`||` positions, an
  `elseif` after a failed test and an exit followed by `else` now count.
- TempnamDirectoryFallbackUnchecked reads the canonical-directory check from
  type narrowing: a comparison kept in a local, compared with a boolean or
  used as an `&&` operand, and an exit followed by `else` now count; a
  rebinding of the path or a jump label between the check and the write
  voids it.
- OptionalRegexCaptureReadWithoutNullFlag reads capture existence tests
  from type narrowing: `match (true)` arms and an enclosing `if` now count.
- Type narrowing follows a boolean saved from a test of a property or an
  array element (`$ok = $_FILES['f']['type'] === 'image/png'; if ($ok)`)
  while the member stays unchanged, and the tests of a copied value along a
  chain of plain copies. In proof-grade checks, a call of the project
  between the copy of a property and its test now ends what the test proves
  of the copy. UploadClientMimeTrusted reports a move guarded by a saved
  client type test.
- Type narrowing follows array elements named by a local key
  (`$_FILES[$field]`) while neither the local nor any element of the array
  changes.
- DirectoryIteratorDotEntries reads entry tests from type narrowing: a
  saved test, `match (true)` arms and an exit followed by `else` now count,
  and only tests on the path to the mutation (not any earlier exit in the
  loop) are used.
- UploadClientMimeTrusted reads client type tests from type narrowing:
  ternary and `||` positions, an exit followed by `else`, disjunctions of
  MIME tests and `in_array()` with named arguments now count. A call of the
  project between the test and the move no longer counts as proof.
- OpenSslSigningFailureIgnored reads signature tests from type narrowing:
  comparisons with `''` or `null` (also as tests themselves), an
  `openssl_verify()` result kept in a local or compared loosely,
  `match (true)` arms and an exit followed by `else` now count.
- The rules that need a method call to have succeeded
  (ImagickSingleBlobDropsRequiredFrames, MysqliEscapedValueSurvivesCharsetChange,
  MysqliMultiQueryResultsNotDrained, PdoFetchIntoRetainsSharedRows,
  SqliteForeignKeySettingInsideTransaction, XmlWriterMemoryOutputInUriMode)
  read that success from type narrowing: a saved result, `elseif`, ternary
  and `&&`/`||` positions, comparisons with booleans and an exit followed by
  `else` now count; a jump label between the test and the use voids it.
- RedirectContinuesProtectedExecution reads the authorization guard from
  type narrowing: comparisons with `true`/`false`, `match (true)` arms and an
  exit followed by `else` now count.
- HeadersAfterCommittedOutput reads `headers_sent()` tests from type
  narrowing: a saved test, `while` and `match (true)` guards and an exit
  followed by `else` now count.
- SessionCookieOptionsSetAfterStart reads failed starts and
  `session_status()` tests from type narrowing: saved results, ordered
  comparisons (`< PHP_SESSION_ACTIVE`) and an exit followed by `else` now
  count.
- LocalizedNumberTrailingInputAccepted reads the offset guard from type
  narrowing: a comparison with the literal length, a `<` exit, a saved
  comparison and an exit followed by `else`/`elseif` now count.
- EmptySplCollectionExtraction reads `isEmpty()` and `count()` tests from
  type narrowing: a saved test, comparisons with `true`/`false`, mirrored
  and other integer comparisons (`0 < $q->count()`, `=== 1`, `< 1` exits),
  a truthy `count()`, `match (true)` arms, `for` steps and an exit followed
  by `else` now count.
- NonExhaustiveEnumMatch reads the reachable cases from type narrowing.
  Comparisons now narrow inside loop conditions, enclosing `match` arms
  (including `default`), negated comparisons with `false` and booleans
  holding a comparison, and inside enum methods on `$this`. A subject that
  changes after the match, or a by-reference parameter that nothing between
  the comparison and the match can reach, no longer cancels the narrowing.
  A call to an unresolved function or method that receives the subject now
  cancels it, as does a by-reference parameter or `global` local followed by
  any call.
- Proof-grade type facts about a local end where it may change unseen: a
  write through a reference bound to it, a call that may take it by
  reference (including unresolved callees), any call for a by-reference
  parameter or `global` local, and `eval()`, which now counts as a dynamic
  write like `include`. CsvEscapeDefaultDependency and
  HeapIterationConsumesCollection no longer report after `eval()`.
- ReflectionCompositeTypeAssumedNamed reads `instanceof` guards from type
  narrowing: `while` conditions, `is_a()`, comparisons with `true`/`false`
  and a test result kept in a local now count; a jump label between the
  test and the call ends the proof.
- PregReplaceFailureUnchecked reads string tests (`str_ends_with`,
  `strpos(...) !== false`, offsets) from type narrowing, which also covers
  `stripos`, `strstr` and the other substring searches.
- ConnectTimeoutAssumedToBoundReads reads a checked `stream_set_timeout()`
  through type narrowing: `stream_set_timeout($s, 2) or exit;` and a result
  kept in a boolean now count.

### Fixed

- A condition on a property no longer narrows the property after it is
  reassigned inside the guarded branch.
- A condition on a `static` local no longer narrows it past a call that may
  run the function again and assign it.
- A namespaced function named like a type check (`App\is_string()`) no longer
  narrows its argument as PHP's own check would.
- A value assigned inside a loop and carried to the next iteration is no
  longer narrowed by a condition outside the loop, which it never passed.
- Statements at file level (outside any function) narrow what follows them:
  `if (!$path) { exit; }` now guards the rest of a script.
- `$x or throw …;` no longer narrows `$x` when analysing for a PHP version
  before 8.0, where `throw` is not an expression.
- A silenced test (`if (@$x)`) narrows `$x` like the plain test;
  XslTransformationFailureConsumed no longer reports output it guards.
- The checks that need a proof treat a strict comparison with a constant of
  a type the declaration rules out (`$flag === 1` for a `bool`) as a branch
  that never runs, not as a proof.
- A guard written before a variable is assigned in a later condition
  (`if ($x = f())`) no longer narrows the new value.
- A needle found in a haystack (`str_contains($x, 'a')`, `strpos($x, 'a')
  !== false`, also with named arguments) rules out null and false for the
  haystack.
- A comparison of `strlen($x)` with an integer that rules out length 0
  rules out null and false for `$x` (`strlen($p) > 0` after `parse_url()`).
- A true `array_key_exists($key, $a)` narrows `$a` to an array whatever the
  key, not only for a literal one.
- `is_array($x) || isset($x['k'])` narrows `$x` to an array: a key probe on
  a value that can only be `false` no longer keeps `false` in the union.
- The failure checks that need a proof no longer accept a property test past
  a call that may run project code, a write of a property of the same name
  through another variable, or on a hooked or magic property.
- The right side of `$a ?: $b` is narrowed by `$a` being falsy.
- An `elseif` condition is narrowed by the earlier conditions of its chain
  having failed.
- Type inference narrows past a statement that leaves on one outcome of a
  test: `$x or exit;`, `$x || throw …;`, `$x ?: throw …;`,
  `$x ? null : exit(1);`.
- A value assigned right before `continue 2` (or a deeper level) no longer
  reaches the reads of the loops it leaves, and a `for` loop's step
  expressions are narrowed by its condition when the body does not assign
  the variable.

### Security

- Require Go 1.27.2 to include current HTTP, TLS and root-confinement fixes.

## [0.1.6] - 2026-10-19

### Added

- Analysis gaps, places where an inspection stopped at a work bound instead of
  answering, are reported apart from findings in JSON, SARIF, the CLI summary
  and editor logs.
- The `flow.assertions` setting (off by default) lets `assert()` calls count
  as guards, assuming assertions run and a failed one stops execution.

### Changed

- Re-analysis after an edit is faster: a file's functions are analysed once
  for summaries and findings, and results that read no changed summary are
  kept. Repeated builtin and method lookups are cached.
- Documentation matches the binary more closely, and the rule index shows
  each rule's origin (native or EA-modelled).

### Fixed

- UnescapedHtmlOutput understands far more HTML contexts (script and style
  text, URL, event-handler, `style` and `srcdoc` attributes, unquoted and data
  attributes) and more escaping functions, follows request superglobals
  rewritten or refilled before output, reports every unsafe part of an output
  statement and highlights the exact value.
- Failure-check inspections accept webmozart/assert and PHPUnit assertions and
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

- `IsIterableCanBeUsed` offers a quick fix for an adjacent
  `is_array($v) || $v instanceof Traversable` pair on a plain variable.
- Security inspections cover more sinks: interpolated, `print` and SVG/XHTML
  output, the backtick operator and string-form `proc_open`, remote stream
  reads, `curl_setopt_array` destinations, LDAP filters escaped for the wrong
  context, bit-or permission modes and cURL basic authentication set through
  `CURLOPT_URL`.
- The flow analysis follows functions containing `switch`, `match (true)` and
  `goto`, by-reference outputs, silenced calls, `global` variables written by
  other functions and `static` locals, so value, failure-check and
  request-data inspections report in more places.

### Changed

- Project analysis is faster on large projects: summaries are computed in
  parallel, built once instead of per file, and a file is not analysed again
  when nothing it calls changed. Very long functions and files with many
  `goto` labels no longer exhaust the analysis budget or take seconds.

### Fixed

- Quick fixes no longer produce invalid PHP or change behaviour: they add
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
- Failure-check inspections accept many more guard forms (early exits,
  `or exit`, stored tests, never-returning calls, `switch` and `match`
  allow-lists, type predicates, `isset`/`empty` on an element), and no longer
  let a check vouch for a later, unchecked assignment or for a value that a
  `goto` carries past it.
- Request-data inspections track array elements, references, by-reference
  parameters, statement bindings, `extract()` and `parse_str()` more
  precisely, report request data combined with unknown values, and no longer
  report numbers, booleans or `switch`/`match`-selected values.
  `UnescapedHtmlOutput` examines every part of an output statement.
- Architecture and unused-code inspections respect property hooks, abstract
  and promoted properties, trait methods, `__sleep()` lists and PHP-version
  differences of built-in classes.
- Inspections that skip unreachable code no longer skip code a `goto` label
  makes reachable, nor property hooks after a top-level `return`.
- HTTP, cookie, PDO, mysqli and SQLite inspections compare names
  case-insensitively and tolerate SQL comments and state restored elsewhere.
- Malformed input no longer makes some inspections skip the file.

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

- Native PHP inspections covering references, generators and fibers, arrays,
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

- Native inspections handle branch reachability, mutation, callbacks,
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
