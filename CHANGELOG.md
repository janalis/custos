# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- PossiblyNullAccess and PossiblyNullArgument (warnings, off by default)
  report a value the code alone (not PHPDoc) shows may be null, accessed
  with `->`, `::` or `[]`, or passed to a parameter whose declared type
  rejects null. Null checks, early exits, `isset()`, `?->`, `??`,
  `assert()`, PHPDoc assertions, a check of the same call made earlier
  and a check of a property before a call keep them silent, and so does
  null passed to a scalar parameter of PHP's own functions outside strict
  mode. NullPointerException leaves to them what they report when they
  run in the same analysis.
- UndefinedMethod and UndefinedProperty (errors, off by default) report a
  method call, a property read or a static property access that the
  receiver's class does not declare, when the code alone (not PHPDoc)
  shows the class and its whole hierarchy is known. Magic methods,
  `@method`, `@property` and `@mixin` tags, `stdClass` and
  `#[AllowDynamicProperties]`, members a subclass or an implementation
  declares, traits, closures that may be rebound, classes declared twice
  or conditionally, members PHP declares in another version, and
  `method_exists()`, `is_callable()` (also of a `'Class::method'` string)
  or `property_exists()` checks keep them silent.
- UndefinedFunction and UndefinedClass (off by default) report calls and
  class names nothing declares. They report only when the project's world
  is complete: Composer's packages all installed and indexed, every
  required `ext-*` described by the stubs, and nothing the analysis skips
  (excluded, hidden, symlinked or vendor test directories, files too large
  to analyze) able to declare the name. Existence, extension and version
  checks about the name, early exits included, keep them silent. The
  language server indexes the project again when `installed.json` changes
  or when `composer.json` changes its autoload configuration,
  requirements, suggestions or vendor directory, keeping unsaved edits.
- InvalidArgumentType (error) and DocumentedArgumentType (warning), both
  off by default, report an argument the parameter type can never accept.
  The first judges the argument by the code alone, where PHP would throw a
  `TypeError`; the second takes PHPDoc into account: a `@param` type, or an
  argument typed only by PHPDoc. Both follow PHP's weak-mode conversions
  outside `declare(strict_types=1)` and check PHP's own functions from PHP
  8.0, except those whose stubs declare other parameter types than PHP.
  A value read from `$this` in a closure or an arrow function, which may
  be bound to another object, is reported only when its type does not
  depend on that object (by InvalidArgumentType and InvalidReturnType).
- InvalidPropertyAssignment (error) and DocumentedPropertyAssignment
  (warning), both off by default, report a value stored in a property
  whose type can never accept it, through plain and compound assignments,
  `??=`, by-reference assignments of a local, list destructuring
  (`[$this->a, $this->b] = $pair`), foreach keys and values, and `++`
  (`null` incremented to an integer). The first judges the value by the
  code alone, where PHP would throw a `TypeError`; the second also checks
  `@var` types, made nullable by a `null` default or, on an untyped
  property, by no default, and values typed only by PHPDoc. Both stay
  silent on hooked properties, classes with `__set`, writes the scope
  cannot perform and readonly writes that may be re-initializations.
- InvalidReturnType (error) and DocumentedReturnType (warning), both off
  by default, report a returned value the function's own return type can
  never accept, closures and arrow functions included. The first judges
  the value by the code alone; the second takes PHPDoc into account: a
  `@return` type, or a value typed only by PHPDoc. Generators, `void` and
  `never`, `self` and `static` in traits, templates and inherited
  declarations are left alone.
- `ArgumentCountMismatch` (off by default) reports a call that leaves a
  required parameter without an argument, and a call passing more arguments
  than a PHP function or method declares. Overrides in subtypes that accept
  the call, unpacked or unknown named arguments, magic methods and, before
  PHP 8.0, PHP's own functions are left alone.
- `--parse-once` (or `"parseOnce": true` in `custos.json`) parses each
  analysed file once and keeps its syntax tree from indexing until its
  analysis, instead of parsing it again in each phase. On laravel's `src`
  a run takes about 5–15% less time for about 40% more memory at peak.
  Findings are the same.
- `analyse` and `fix` keep the cross-file summaries of the checked files in
  the cache, next to the declarations of dependencies. A run after an edit
  summarizes again only what the edit reaches, as the language server does;
  the summaries are set aside when declarations, dependency files, the set
  of checked files or the working directory change. Findings are the same
  with or without them, and `--no-cache` bypasses them too. A re-run of
  composer's `src` after editing one file took about 0.55 s instead of
  1.1 s, at no more memory; the cache keeps about 5 MB for laravel's `src`.

### Changed

- DocumentedReturnType, DocumentedArgumentType and
  DocumentedPropertyAssignment compare integer and string literals with
  integer ranges and non-empty strings: `return -1;` under
  `@return positive-int`, `f(0)` against `@param positive-int` and
  `$this->label = ''` against `@var non-empty-string` are reported, as is
  `f('up')` against `@param 'asc'|'desc'` (literal string values), and the
  message names the value (`not 0.`). Inference gives an integer literal its
  value and a string literal its byte length and value, which a variable
  keeps only when no write its assignments do not show may have replaced
  it: through a reference, a `global` or `static` local, or a call that may
  take it by reference.
- The NestedPositiveIfStatements quick-fix writes `elseif` instead of
  `else if` in a file that uses `elseif` more often than `else if`.
- `analyse` and `fix` are faster on projects with large classes in
  dependency cycles: the cross-file summaries no longer rebuild their
  collection each round, and a file summarized again keeps the summaries
  of the methods that no changed contract reaches. Runs on the `src` of
  laravel, composer and guzzle take about 15%, 11% and 7% less time.
- The language server updates cross-file summaries much faster after an
  edit within a dependency cycle: files whose source did not change keep
  the summaries their contracts still support. In laravel, the update after
  an edit to `Str.php` takes about 0.3 s instead of 2.2 s, and to the
  Eloquent `Model.php` about 0.6 s instead of 2.4 s. The server keeps up
  to about 30 MB more on a project of that size.
- The language server reports findings faster in a class with many
  descendants that inherits many trait methods: ReturnTypeCanBeDeclared no
  longer looks up, for each descendant, the trait methods it only inherits.
  Each edit to laravel's Eloquent `Model.php` takes about 60% less time
  to analyze.

### Fixed

- Type inference keeps more null checks: a local passed to a by-value
  parameter of a method inside a branch, or reassigned on one branch
  after an early exit, a variable assigned in the left operand of `&&`
  (read in its right operand or the body it guards), an `if`/`else`
  whose other branches `continue` or `break`, a `return` inside `try`,
  a value stored in a `do`/`while` body and tested by its condition, a
  test of `$a['k'] ?? null`, and a property set to `x ?? []`.
- The language server no longer replaces the declarations of an open
  document's unsaved edits with the file on disk: file changes replayed
  after a re-index skip open documents, and their buffers are indexed
  again once the rebuild ends.
- A condition on a property (`if ($this->buffer === null)`) no longer holds
  past a call that may run project code, which may write the property: rules
  using PHPDoc types no longer read a property refilled by `$this->send()`
  as still null (a DocumentedReturnType false positive).
- UntrustedForwardedClientAddress takes a proxy check that leaves on
  failure (`knownProxy($ip) or exit;`) or whose stored result is tested
  (`$trusted = knownProxy($ip); if (!$trusted) { return; }`) as a guard.
- ForgottenDebugOutput no longer reports a debug function called after
  `function_exists()` confirmed it through an early exit
  (`if (!function_exists('f')) { return; }`), an `else` arm, a
  short-circuit operand or a stored result.
- DateLastErrorsFalseUnchecked no longer reports `getLastErrors()` read
  after a failed parse when the failure is known from an `else` or `elseif`
  arm, a short-circuit operand (`DateTime::createFromFormat(…) ?: throw …`),
  a stored result, or statements between them that parse nothing.
- JsonEncodingApiUsage takes a result tested as an operand of `&&`, `||`,
  `and` or `or` (`$data || throw …;`, `if ($ready && $data)`) as a
  handled failure.
- TransactionEarlyReturn accepts a `finally` cleanup written as
  `$pdo->inTransaction() && $pdo->rollBack();` (or `!… || …`) or as the
  `else` arm of `if (!$pdo->inTransaction())`.
- UntrustedForwardedClientAddress reports a forwarded-address comparison
  written as the condition of an `elseif` arm; it only looked at `if`
  conditions. The earlier conditions of the chain count as proxy guards.
- MkdirRaceCondition no longer reports a silenced `@mkdir()` that only runs
  once `is_dir()` returned false when the guard is an early exit
  (`if (is_dir($d)) { return; }`), an `else` or `elseif` arm, or a
  short-circuit operand (`is_dir($d) || @mkdir($d);`).
- VariableFunctionsUsage no longer rewrites `call_user_func([$cb, 'm'])`
  into `$cb->m()` when `$cb` is typed `callable`: a callable may be a class
  name, which the direct call cannot take.
- UnlinkFailureReportedAsSuccess no longer reports the deletion of a
  `tempnam()` file when the namespace declares its own `tempnam()` (a test
  double).
- FileReadFailureUnchecked no longer reports `unserialize()` or
  `json_decode()` of a read compared by `assertSame` with an expected value
  typed as neither false nor null (`array $expected`, an object, `[]`), or by
  `assertEquals` with a non-empty array or an object: the failure value
  fails the assertion.
- Arithmetic operators on a `\GMP` or `\BcMath\Number` operand, and shift
  and bitwise operators on a `\GMP` operand, are typed as that object
  instead of an int or a number, so type-based rules no longer treat the
  result as a scalar.
- Generic arguments given to a trait by an `@use` tag on the `use`
  statement inside the class (`/** @use EnumeratesValues<TKey, TValue> */`),
  as Laravel writes them, are now read; only the class's own doc was. The
  trait's methods then return, and pass to callbacks, the class's element
  type, which can add ReturnTypeCanBeDeclared suggestions.
- PhpUnitTests no longer rewrites to `assertRegExp()`, `assertNotRegExp()`,
  `assertFileNotExists()` or `assertDirectoryNotExists()` when the PHPUnit
  version is only assumed, not configured or detected: those methods no
  longer exist from PHPUnit 10. The call is still reported, and the message
  names both the PHPUnit 9.1 method and the one it replaced.
- PhpUnitTests treats a `PHP_UNIT_VERSION` that names no version as unset:
  the version is taken from the installed PHPUnit when it is indexed, and
  is otherwise unknown, not PHPUnit 8.

## [0.1.11] - 2026-11-09

### Added

- `custos self-update` updates custos the way it was installed: it runs
  `brew upgrade` for Homebrew, `composer update janalis/custos` for the
  Composer package, and replaces a downloaded release binary after checking
  it against `SHA256SUMS`. `--check` only reports whether a release is out.
- `analyse` and `fix` print a one-line notice when a newer release exists.
  They check at most once a day, while the command runs (waiting at most a
  second for it), and only when stderr is a terminal. The check is skipped
  in CI and when `CUSTOS_NO_UPDATE_CHECK` is set.

### Changed

- The README shows the time and peak memory of custos and the other tools
  on the largest benchmarked project, and a compact feature table; notes and
  sources stay on the benchmark page.

## [0.1.10] - 2026-11-08

### Added

- A [benchmark page](https://janalis.github.io/custos/guide/benchmark)
  compares the time, peak memory and findings of custos with PHPStan, Psalm,
  Phan, Mago, PHP_CodeSniffer, PHP-CS-Fixer and Rector on real projects. A
  weekly workflow moves every tool and project to its latest release and
  refreshes the numbers.
- The benchmark page and the README compare what custos and those tools
  offer: type checking, taint analysis, automatic fixes, formatting, language
  server, MCP server, baseline, custom rules and SARIF output, each cell
  linked to the tool's documentation.

### Changed

- Faster analysis and fixes, with less memory and the same findings and
  edits. On laravel/framework (Linux, 4 cores), `custos analyse` runs about
  2.1 times as fast as in 0.1.9 and `custos fix` about 2.4 times as fast;
  both need about 20% less memory. Lookups through large class hierarchies
  are cached once per run, files no longer wait on each other for method
  lookups, flow analysis does less serial work and no longer keeps every
  syntax tree, machines with up to 8 cores use all of them for flow
  analysis, and `custos fix` runs only the rules that offer a quick-fix.

## [0.1.9] - 2026-11-07

### Added

- When your code calls a function or method of a dependency, custos
  summarizes the one file under `vendor/` that declares it, on demand and
  once per run or editor session, so request data passed through library
  wrappers is followed to their sinks. Files over 512 KB and callables
  declared more than once are left out. Far fewer calls are reported as
  `unresolved-dispatch` (from 2,074 to 331 scopes on one Symfony
  application) for about 5% more time on a cold run.

### Changed

- Flow analysis evaluates the small pure helpers that reshape data on its
  way to a sink, keeping track of which input each result came from:
  string transforms (case changes, trims, padding, splitting, joining,
  replacement, reversal), array builders and queries (`array_map` and
  `array_filter` with closures or named callbacks, `array_column`,
  `array_combine`, `array_merge`, slices, `array_pad`, `range`,
  `array_unique`, strict `in_array` and `array_search`, `min`, `max`, sums
  and products), `match` and `switch` returns, and `foreach` loops that
  build, filter or reduce arrays. Integer overflow, version-dependent errors
  and uncertain coercions stop the evaluation instead of guessing.
- Callables are followed further: local and returned closures, arrow
  functions, first-class callables, invokable objects, callable names and
  callbacks given to project functions carry their return values and sinks
  back to the caller.
- Streams and other resources are tracked through wrapper functions,
  closures and returned arrays, so StreamUseAfterClose and the failure-check
  rules see the opens and closes made in helpers.
- Cross-file flow follows dependencies on a first run too: long call chains
  no longer take one round per file. `--flow-order` is now only a scheduling
  hint and no longer changes results.
- Type inference keeps element types through array unions, spreads,
  computed keys and documented generics, and no longer drops known types
  when an impossible narrowing meets a loop or branch join. Loop variables
  narrowed back to the same type on each iteration are inferred, so fewer
  files report `inference-recursion`.
- `$this->method()` in a trait resolves to the trait's own or used method.
  Long fluent call chains no longer exhaust the analysis depth limit and
  take linear time.
- SuspiciousAssignments checks parameter overwrites against flow: a read
  through `goto`, an arrow function or closure created before the
  overwrite, `compact()`, `get_defined_vars()` or `func_get_arg()` prevents
  the report, while introspection after the overwrite does not.
- Builtin calls bind positional, named and unpacked arguments alike;
  CurlSslServerSpoofing reads the resolved option and value.
- NotOptimalRegularExpressions checks constant heredoc and nowdoc patterns.
  The PHPUnit rules read metadata attributes and `@depends clone` or
  `shallowClone`. PdoPlaceholderBindingMismatch keeps a known parameter set
  through conditional rebinding of the same parameters.
- Flow analysis allocates much less per edit; the background flow update
  after an edit in the language server takes about 20% longer because of
  the dependency summaries, while diagnostics are published as quickly.

### Fixed

- Calls through an interface, an abstract method or an overridable method
  no longer yield a complete value from the declared body: the result keeps
  what that body may return but no longer proves facts an override could
  contradict.
- DisconnectedForeachInstruction treats a nullsafe call (`?->`) like `->`
  when deciding whether its receiver depends on the iteration.
- DynamicCallsToScopeIntrospection and the other value checks consider both
  results of a short ternary (`$a ?: $b`). UnqualifiedReference skips the
  global namespace block of a file that mixes namespaces.
- CascadeStringReplacement's quick-fix no longer drops the merged elements
  when the target list is an empty array literal, and calls with an empty
  search or replacement list are merged when the fix provably keeps the
  result and the order of conversions.

## [0.1.8] - 2026-11-04

### Added

- Without a `comparisonStyle` in `custos.json`, the comparison style comes
  from the project's PHP-CS-Fixer configuration (its `yoda_style` rule, or
  the `@Symfony` and `@PhpCsFixer` rule sets) or from Mago's
  `yoda-conditions` rule. Projects formatted to `null === $x` no longer get
  every comparison reported, and quick-fixes write the order they keep.
- `custos analyse` and `custos fix` cache the declarations of the files they
  only index (`vendor/` and the rest of `paths`) between runs, in the user
  cache directory or `$CUSTOS_CACHE_DIR`; `--no-cache` turns it off. Checking
  one file of a project with 20,000 vendor files went from about 6 s to
  0.4 s of CPU time. The language server builds its startup index from the
  same cache.
- `--flow-order` (or `"flow": {"order": true}` in `custos.json`) keeps the
  order in which cross-file flow was summarized, so the next run summarizes
  each file after the ones it calls: about 12% faster on symfony/src, and
  long call chains are followed to the end instead of being reported as
  unresolved.

### Changed

- Type inference knows more values: subclasses documented on declared
  types, elements created by nested writes, the matches of
  `preg_match_all()` and `preg_replace_callback()`, `static` locals, closure
  parameters from documented callback signatures, returns inherited through
  `{@inheritDoc}`, array casts, `$this` in traits declaring
  `@phpstan-require-extends`, and types imported with `@phpstan-import-type`.
- The command line, the language server and the MCP server discover files
  alike. `custos analyse` and `custos fix` index every file under `paths`
  even when the command line names only some, so cross-file results no
  longer depend on the files a run checks. `exclude` applies under `vendor/`
  too, and entries such as `var/cache` match at any depth. php-src's
  `*.stub.php` and Drupal's `*.api.php` files are indexed but not analyzed,
  and a file named twice is analyzed and fixed once.
- The language server publishes an edited file's diagnostics before it
  updates the cross-file flow summaries: under 100 ms instead of 15 s or
  more on symfony/console's Application.php. It summarizes again only the
  files an edit can change when that provably gives the same result (40%
  less CPU time there), and no longer rebuilds every summary when a file is
  opened.
- A crash on one file no longer ends a run or a server: the file gets an
  `internal` finding and the `internal-failure` status. Quick-fixes apply
  whole or not at all, and a fixing pass that would add syntax errors is
  undone. `custos fix` reads each file once instead of three times.
- A `-rule` run loads the project index like a full run, so it reports and
  fixes what a full run does.
- Taint and value tracking reach further: chains of more than eight wrapper
  functions across files (past 32 files, calls are reported as
  `unresolved-dispatch`), recursive functions passing their argument to a
  sink, and methods called on a nullable receiver or reaching `__call()`.
- A `.php-version` file at the project root sets the PHP version when
  neither `custos.json` nor `composer.json` does.

### Fixed

- Fewer false positives and safer quick-fixes across 89 rules. The
  failure-check rules (FileReadFailureUnchecked, JsonEncodingApiUsage,
  StreamOpenFailureUnchecked, PregReplaceFailureUnchecked and others) take
  PHPUnit and `webmozart/assert` assertions, typed PHP parameters and test
  exits such as `$this->fail()` as checks. UnnecessaryAssertion,
  OnlyWritesOnParameter, ReturnTypeCanBeDeclared, TypeUnsafeArraySearch,
  TypeUnsafeComparison, NullPointerException and the PHPUnit rules report
  far less on real projects.
- Quick-fixes of NestedPositiveIfStatements, MissingOrEmptyGroupStatement,
  the `str_contains()` family and others keep comments, indentation,
  needed parentheses and the code's `elseif` spelling.
- Flow analysis keeps what it proved before a later `include` or `eval`,
  narrows through a test stored in a condition
  (`if ($negated = str_starts_with($line, '!'))`), and no longer lets a value
  flow back through a loop from a block that returns.
- `$this`, `self` and `static` in the constructor arguments of an anonymous
  class refer to the class that creates it.
- The MCP server applies `flow.assertions` and reads `exclude` relative to
  each scanned path. The `excludedPaths` of HardcodedCredentialAtKnownSink
  fall back to the project or workspace root instead of the working
  directory.
- The taint rules no longer carry request data through `mysqli_num_rows()`
  and the other functions returning a number.

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

[Unreleased]: https://github.com/janalis/custos/compare/v0.1.11...HEAD
[0.1.11]: https://github.com/janalis/custos/compare/v0.1.10...v0.1.11
[0.1.10]: https://github.com/janalis/custos/compare/v0.1.9...v0.1.10
[0.1.9]: https://github.com/janalis/custos/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/janalis/custos/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/janalis/custos/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/janalis/custos/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/janalis/custos/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/janalis/custos/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/janalis/custos/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/janalis/custos/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/janalis/custos/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/janalis/custos/releases/tag/v0.1.0
