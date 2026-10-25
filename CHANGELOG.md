# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `custos analyse` and `custos fix` keep the declarations of the files they
  index without checking them (`vendor/` and the rest of `paths`) between
  runs, in the user cache directory or `$CUSTOS_CACHE_DIR`. A run over a few
  files no longer parses every dependency again: on a project with 20,000
  vendor files, checking one file went from about 6 s to 0.4 s of CPU time.
  `--no-cache` turns it off. The language server builds its startup index
  from the same cache.
- `--flow-order` keeps, in the same cache, the order in which `custos
  analyse` and `custos fix` summarized the cross-file flow of the checked
  files, so the next run with it summarizes each file after the ones it
  calls: on symfony/src, a repeated run takes about 12% less time. A call
  chain across files too long for a run without it may then be followed to
  the end instead of being reported as unresolved. `"flow": {"order": true}`
  in `custos.json` turns it on for every run and for the language server,
  which reads and saves the same order and follows it when an edit makes it
  rebuild every summary.

### Changed

- Type inference knows more values: a property, parameter or return type
  documented as a subclass of its declared type (`@var Text` on a `Message`
  property), the elements nested writes create (`$groups[$key][] = $row`),
  the matches of `preg_match_all()` with `PREG_SET_ORDER` or
  `PREG_OFFSET_CAPTURE` and those `preg_replace_callback()` passes its
  callback, `static` locals, and the parameters of a closure
  passed where a function or method documents the callback's signature
  (`@param Closure(T, TKey): bool $p` on a generic collection). An array or
  an empty one (`$list ?? []`, `$found ? $list : []`) keeps the elements of
  the array, and a method declaring only `: array` or `: iterable` keeps the
  documented return of the method it overrides (`{@inheritDoc}` on
  Doctrine's `getAllMetadata()` reads `list<ClassMetadata>`). The array a
  `foreach` iterates no longer takes the types its body assigns to it later.
  A type check narrows an element written `$a['k'] ?? null` too, and an
  element of an array nothing is known of. A write below an element
  (`$groups[$id]['files'][] = $file`) updates it along its keys, so the
  element keeps its shape and `files` holds the files written, where the
  element became a plain array or kept the empty `files` it started with.
  A callback parameter declared `?array` takes the elements passed as one
  declared `array` does, a closure declaring only `: array` returns the
  array its body builds, `(array) $x` keeps the elements of $x (a scalar
  becomes a list of itself), and `array_replace_recursive()` keeps the
  elements of its arrays. Inside a trait that requires its users to extend
  a class or implement an interface (`@phpstan-require-extends`,
  `@require-implements`, …), `$this` is of that type, after the trait's
  own members. A type alias imported with `@phpstan-import-type` (or
  `@psalm-import-type`) reads as the class it comes from declares it, where
  it read as `mixed`, following that class's own imports, also in the
  members other files read.
- HardcodedCredentialAtKnownSink matches `excludedPaths` from the workspace
  root in the language server and the MCP server when a file has no
  `custos.json` or `composer.json` above it, instead of the server's working
  directory; over MCP it no longer looks outside the workspace for them.
- The language server builds cross-file flow summaries once when it indexes
  the workspace instead of twice, and `custos fix` reads each file once
  instead of three times, fixing the same bytes it indexed.
- The language server publishes an edited file's diagnostics before it
  updates the file's cross-file flow summaries, then checks open files again
  if they changed. Editing a file on a cycle of calls across files, such as
  symfony/console's Application.php in the Symfony repository, no longer
  holds its diagnostics for 15 seconds or more while every summary is
  rebuilt: they arrive in under 100 ms. The update waits until editing
  pauses for a second; updates run one at a time, and edits made during one
  are summarized together afterwards.
- After an edit within a cycle of calls, the language server summarizes
  again only the files the edit can change and those whose contracts they
  consult, as long as that provably gives what summarizing every file does;
  otherwise it still summarizes every file. On the Symfony repository, the
  update after an edit to symfony/console's Application.php takes 40% less
  CPU time.
- Opening a file in the language server no longer rebuilds the cross-file
  flow summaries of the whole workspace: a buffer holding the saved file
  changes nothing, and one with other text but the same declarations
  updates the summaries as an edit does.
- The command line, the language server and the MCP server discover files
  alike. `custos analyse` and `custos fix` index every file under `paths`
  even when the command line names only some, so cross-file results no
  longer depend on which files a run checks. `exclude` applies under
  `vendor/` too, the `janalis/custos` package is not indexed, and a project
  path that contains `vendor/` no longer indexes it twice. A symlink found
  while scanning a directory is analyzed only when it leads to a regular file
  inside that directory. Files that rules read by name, such as
  `composer.json`, are no longer parsed as PHP for the index.

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
- The MCP server applies `flow.assertions` to cross-file flow summaries and
  to the types its `type` tool reports, like the CLI and the language server.
- The MCP server reads an `exclude` path relative to each scanned path, as
  documented and as the command line does, instead of the project root.
- On the command line, `excludedPaths` of HardcodedCredentialAtKnownSink fall
  back to the project root, not the working directory, for a file with no
  `custos.json` or `composer.json` above it.
- A recursive function or method that passes its argument on to a sink no
  longer keeps growing its flow summary until it is dropped: calls into it
  are followed instead of being reported as `unresolved-dispatch`.
- TypeUnsafeComparison no longer prefers `===` for a truthiness test
  (`$flag == true`) or an emptiness test on a `mixed` value.
- SelfClassReferencing no longer offers to rewrite `Name::m()` as
  `self::m()` in a class whose hierarchy is not fully indexed.
- NullCoalescingOperatorCanBeUsed no longer offers a quick-fix that would
  drop a comment.
- ReturnTypeCanBeDeclared no longer offers a quick-fix on abstract and
  interface methods, and keeps `static` through `parent::create()` calls.
- IssetConstructsCanBeMerged, CascadeStringReplacement, OpAssignShortSyntax,
  IfReturnReturnSimplification and StrTrUsageAsStrReplace no longer offer a
  quick-fix that would drop a comment.
- OneTimeUseVariables no longer deletes a doc comment typing another
  variable when it inlines one.
- NestedPositiveIfStatements no longer offers a quick-fix that moves a
  directive comment (`@phpstan-ignore-next-line`) onto another statement.
- PhpUnitTests no longer reports a `@covers` class target as unresolved
  when the test's own parent is not indexed (the framework is missing).
- TypeUnsafeArraySearch takes class constants holding plain strings
  (`[self::USER, Webform::PAGE]`) like the string literals they hold.
- PackedHashtableOptimization no longer asks to write numeric-string keys
  as integers: PHP already stores them as integers.
- MagicMethodsValidity no longer asks for a companion magic method in a
  trait or when an ancestor is not indexed.
- The failure rules no longer take an array for a call's result once an
  element is written with it (`$files[] = $file` from `glob()`), and the
  untrusted-path and destination rules no longer follow request data
  through the text `str_replace()` or `preg_replace()` searches for.
- JsonEncodingApiUsage also accepts a ternary on `json_last_error()` in
  the next statement and a property tested at its first read.
- CallableParameterUseCaseInTypeContext no longer reports a method given
  the parameter whose union return type fits it in part
  (`$form = $tokens->replace($form)`, `string|array`).
- NotOptimalRegularExpressions no longer asks for `/s` on a `/m` pattern.
- ReturnTypeCanBeDeclared no longer suggests a `@template` name as a return
  type, nor `void` in a hierarchy where another declaration of the method
  returns a value.
- TypeUnsafeComparison no longer reports an object union with a
  string-convertible member as uncomparable, nor prefers `===` for an
  emptiness test on a value that may be false or null (`getDocComment() !=
  ''`) or a comparison with its own cast (`(int) $v == $v`).
- NotOptimalIfConditions no longer reports `$r instanceof Status && $r ==
  'QUEUED'` when the class converts to a string.
- NestedPositiveIfStatements no longer offers a quick-fix that would drop
  a comment.
- CallableParameterUseCaseInTypeContext no longer reports a ternary that
  gives the parameter back in one branch for that branch's type, nor takes
  arithmetic on a `mixed` value (`max($values) + 1`) for a float.
- ArrayRandInvalidCount no longer takes a property's default as its value:
  other methods may fill it.
- UnusedConstructorDependencies no longer reports a property holding an
  object with a destructor, which it keeps alive.
- OnlyWritesOnParameter no longer reports an array that only keeps FFI
  buffers or objects with destructors alive.
- MagicMethodsValidity no longer asks a constructor to call a parent
  constructor that does nothing without arguments (each statement only
  runs for an optional parameter given a value).
- FopenBinaryUnsafeUsage no longer asks for the `b` flag on streams other
  than plain files (`php://stdout`, `php://fd/3`, registered wrappers) or on
  a mode that comes from a parameter's default.
- JsonEncodingApiUsage looks through `@` for the failure test, and accepts
  an `if` checking `json_last_error()` right after the call.
- PropertyInitializationFlaws no longer reports a re-declaration that
  gives the property another type, nor takes `''` for an inherited `'    '`
  declared in another file.
- TypeUnsafeArraySearch no longer reports a search of a string among the
  values of a class constant listing only strings, nor in a trait's own
  typed static property read as `self::$values`.
- SuspiciousLoop no longer reports a `for` loop that starts from the
  parameter's own value (`for ($index += $step; …)`).
- UnusedConstructorDependencies no longer reports a property assigned an
  array of references (`[&$this->data]`), which makes clones share them.
- AlterInForeach no longer asks to unset the value of an inner
  by-reference loop when the unset follows an outer loop that leaves the
  variable alone.
- ParseUrlFailureDereferenced no longer reports a `parse_url()` result
  read after `$parts ?? null` (or another falsy fallback) was checked.
- TypeUnsafeArraySearch no longer asks for strict mode when the haystack
  lists `Name::class` constants.
- LoopWhichDoesNotLoop no longer reports an empty `foreach` body whose key
  or value goes to a property, an element, a `list()` or a variable read
  after the loop.
- CallableParameterUseCaseInTypeContext no longer calls an `is_*()` test
  "always false" where the parameter holds a value assigned since (or
  bound by `foreach`) rather than the one it was given.
- MagicMethodsValidity no longer asks a constructor to call
  `\Exception::__construct()` or `\Error::__construct()`, which only store
  arguments that default to what the properties hold.
- NotOptimalRegularExpressions no longer suggests `\d` for `[0-9]` (nor
  `\w`, `\D`, `\W` for their classes) in a pattern with the `u` modifier,
  where the shorthand matches the digits and letters of every script.
- ProperNullCoalescingOperatorUsage no longer reports a `true` or `false`
  fallback (`$seen[$id] ?? false`), a deliberate "absent" marker, nor a
  `void` call on the left, which evaluates to null.
- StreamOpenFailureUnchecked and FileReadFailureUnchecked no longer report
  PHP's own streams (`php://memory`, `php://temp`, `php://output`, the
  standard streams), which always open. On PHP 8.0 and later, a stream
  call that returned (`fread($h, 4)`) proves the handle open for the rules
  that look for unchecked failures, since PHP throws for `false` there:
  only the first unguarded use of a handle is reported.
- UnknownNamedArgument no longer reports a named argument to a method of a
  class the project declares more than once (an older copy of the code
  left in the project): which declaration runs depends on the autoloader.
- JsonEncodingApiUsage no longer asks for `JSON_THROW_ON_ERROR` when the
  code tests the result for failure where it is produced (`?: []`,
  `is_array()`, `false === $json`): the flag would turn that fallback into
  an uncaught exception.
- PotentialMalware no longer reports `touch()` with three arguments or
  `get_defined_functions()` in test code, where they set up fixtures.
- The assertion methods of PHPUnit and `webmozart/assert`
  (`$this->assertNotFalse($handle)`, `self::assertIsResource($stream)`)
  now count as guards for the rules that look for unchecked failures
  (StreamOpenFailureUnchecked, GlobFailureUnchecked and the like): they
  throw when they fail.
- SenselessProxyMethod no longer reports an override whose parent method
  carries attributes other than `#[\Override]`: the override hides them
  from reflection on the child (a `#[Route]` on a base controller action),
  so removing it is not neutral.

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
