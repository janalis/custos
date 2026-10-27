# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Without a `comparisonStyle` in `custos.json`, the comparison style comes
  from the project's PHP-CS-Fixer configuration (`.php-cs-fixer.php`, then
  `.php-cs-fixer.dist.php`): its `yoda_style` rule, or the `@Symfony` and
  `@PhpCsFixer` rule sets, which enable it; else from Mago's
  `yoda-conditions` rule in `mago.toml` or `mago.dist.toml`. Projects formatted to
  `null === $x` no longer get every comparison reported, and quick-fixes
  write the order their formatter keeps.
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
- MissingOrEmptyGroupStatement indents the braced body it writes with the
  file's own indentation step instead of always four spaces.
- MissingOrEmptyGroupStatement no longer reports a commented empty `if` or
  `elseif` branch followed by another branch: it deliberately stops the
  rest of the chain.
- SuspiciousLoop's dead-parameter exemption also covers promoted
  constructor parameters, `foreach (func_get_args() as $x)`, and an inner
  loop reusing a name an outer loop already took from the parameter.
- An `include` or `eval` written after a check no longer throws away what
  the flow analysis proved before it: StreamOpenFailureUnchecked and the
  other rules asking whether a value can be false or null keep their facts
  for code evaluated before it, outside any loop around it (php-src build
  scripts).
- ClassReImplementsParentInterface no longer leaves a double space when it
  removes the first interface of a list (`implements A, B` → `implements B`).
- SuspiciousLoop no longer reports a loop variable reusing a parameter
  whose value was already used and is never read again: nothing after the
  loop, up to a `return` or `throw`, uses it.
- MagicMethodsValidity no longer reports a constructor that skips a parent
  constructor which only stores its parameters when the child sets those
  properties through a method of its own class (`$this->replace($files)`,
  Symfony's `FileBag`).
- ForeachInvariants writes `foreach (array_keys($c) as $i)` when no element
  access of the loop body can use a value variable, instead of declaring a
  `$iValue` that stays unused.
- MissingOrEmptyGroupStatement no longer reports an empty `foreach` kept
  for the last entry it leaves: its target is a property or array element,
  or its variable is used after the loop.
- PropertyInitializationFlaws no longer reports a constructor assignment
  of the default value after earlier code may have changed the property
  (`$this->init();` then `$this->sent = false;`): it resets the property.
- AlterInForeach no longer reports an empty by-reference loop
  (`foreach ($keys as &$key) {}`) whose variable is never used afterwards:
  it only turns the elements into references.
- AutoloadingIssues no longer reports a file that also declares functions:
  no autoloader loads it by class name (php-src's extension stubs).
- PhpUnitTests puts a literal or constant first when it turns
  `assertTrue(count($rows) == 2)` into `assertEquals(2, count($rows))`:
  PHPUnit takes the expected value first, so its failure message read
  backwards.
- A file passed on the command line both on its own and inside a passed
  directory is analysed and fixed once: `fix -dry-run -diff` no longer shows
  its diff twice and the file counts are right.
- UnsetConstructsCanBeMerged indents a merged `unset(` written over several
  lines with the file's own indentation step (two spaces in Drupal code)
  instead of always four spaces.
- SuspiciousLoop no longer reports an inner loop reusing the variable of an
  outer `foreach` when the outer value is never read again, also when the
  inner loop sits in an `if` or an intermediate loop; plain writes such as
  a later loop reusing the name do not count as reads.
- UnsetConstructsCanBeMerged no longer offers to merge `unset()` statements
  separated by a blank line and a comment: the comment introduces the
  statements below it, which the merge would move away.
- TernaryOperatorSimplify no longer doubles the parentheses when the
  ternary it simplifies is already wrapped in some
  (`((A && B) ? TRUE : FALSE)` becomes `(A && B)`).
- PhpUnitTests no longer suggests an assertion that the test's own base
  class declares outside PHPUnit (the redis extension's strict
  `assertEquals()`): it may not mean what PHPUnit's does.
- PregQuoteUsage no longer reports a `preg_quote()` result compared with
  `===`, `!==`, `==` or `!=`: it is matched against another value, not used
  as a pattern.
- IssetConstructsCanBeMerged keeps the line breaks of a condition written
  over several lines instead of joining it onto one line.
- NotOptimalIfConditions no longer moves `class_exists()`,
  `interface_exists()`, `trait_exists()` or `enum_exists()` calls that may
  run the autoloader: loading the class is a side effect that later checks
  can rely on (`!class_exists($c) && !interface_exists($c, false)`).
- MissingOrEmptyGroupStatement no longer reports an empty `foreach` over a
  value known to be an iterator (`iterable`, `\Traversable` and its
  subtypes, generators): the loop drains it for its effects.
- UnsetConstructsCanBeMerged no longer merges long `unset()` statements
  written on separate lines into one line of hundreds of characters: past
  120 characters the merged call keeps one line per original statement.
- PregQuoteUsage no longer reports `$chars = preg_quote($chars);` when the
  variable is then only read inside patterns whose delimiter preg_quote()
  escapes anyway, including interpolated ones (`"{^[$chars]++}"`).
- MagicMethodsValidity no longer asks for a `__set` companion in a class
  that unsets one of its own properties (`unset($this->service)`) so that
  `__get` initializes it lazily.
- ForgottenDebugOutput no longer reports `error_log()` inside the methods of
  a PSR-3 logger (a class implementing `Psr\Log\LoggerInterface`), whose
  output it is.
- StaticLambdaBinding no longer offers to remove `static` from a closure
  defined in a static method: the closure would still have no `$this`. The
  report stays, without a fix.
- TypeUnsafeComparison no longer warns about `'true' != $content['success']`
  when the other operand may be a boolean (unknown, `mixed` or `bool`): the
  loose comparison accepts both `true` and `'true'` on purpose.
- DeprecatedIniOptions no longer reports reading, restoring or setting to
  their default value the `assert.*` and `session.sid_*` directives, which
  PHP only deprecates when given another value (`ini_set('assert.active', 1)`).
- SuspiciousLoop no longer reports a `for` condition list whose expressions
  before the last are assignments, increments or calls
  (`for ($x = 0; $c = $x * 2, $x++ < 75;)`): they run on every iteration
  and joining them with `&&` would change the loop.
- MissingOrEmptyGroupStatement keeps the braces of its quick-fix on the same line in one-line template islands (`<?php if ($a) print "x" ?>`), instead of spreading them over three lines inside the HTML.
- CryptographicallySecureRandomness no longer suggests `random_bytes()` in the fallback branch that runs once `function_exists('random_bytes')` failed.
- TraitsPropertiesConflicts no longer reports a promoted constructor property that declares a trait property the same way: promotion has to declare it.
- MagicMethodsValidity no longer asks a decorator constructor that takes an instance of its parent class (`__construct(private parent $inner)`) to call the parent constructor.
- LoopWhichDoesNotLoop no longer reports an empty `foreach` with a by-reference value (`foreach ($a as &$v) {}`), which turns every element into a reference.
- StreamOpenFailureUnchecked and FileReadFailureUnchecked treat any `php://temp…` name, such as `php://temp,`, as a stream that always opens: PHP only reads the `temp` prefix.
- MkdirRaceCondition no longer reports `mkdir()` of a randomly named directory (built with `random_bytes()`, `uniqid()` and the like). Its quick-fix accepted a directory someone else had created in advance.
- PhpUnitTests no longer suggests an assertion method that the test base class does not have, such as a custom `TestSuite` that declares only `assertTrue()`; the quick-fix called a missing method.
- DateTimeConstantsUsage no longer reports, or rewrites to ATOM, the initializer of a constant that declares an ISO8601 alias (`const ISO8601 = DATE_ISO8601;`).
- ForeachInvariants no longer suggests `foreach` when the counter is passed by reference to a `self::`, `static::` or `parent::` method; the rewrite broke loops that skip ahead.
- TypeUnsafeArraySearch no longer reports `in_array()` over class constants declared in another file whose values are plain non-numeric strings.
- Type inference keeps the array an `iterable` or `callable` parameter may
  hold when `is_array()` narrows it, instead of keeping only an array
  literal assigned on another path; UnnecessaryCasting no longer reports
  `(string) $cb[1]` on such a value.
- ProperNullCoalescingOperatorUsage no longer reports a lookup of one key
  in two maps of different objects (`$nodes[$id] ?? $groups[$id]`).
- IsIterableCanBeUsed no longer reports an `is_array()`/`instanceof
  Traversable` pair whose line explains in a comment why it avoids
  `is_iterable()`.
- ForgottenDebugOutput no longer reports files under a `Fixtures`
  directory, whose dumped output their tests read.
- TypeUnsafeComparison no longer asks for `===` between arrays: `==`
  compares their key/value pairs in any order, which `===` would change.
- LogarithmInvalidRealDomain no longer reports `log(0)` or `log1p(-1)`
  written out as literals, a deliberate way to get an infinity.
- UnnecessaryCasting fixes no longer leave a self-assignment: for
  `$key = (string) $key;` the fix deletes the statement.
- FileTruncatedBeforeLock no longer reports a lock file: a handle that is
  only locked and closed, never read or written, holds no content that
  opening it with `w` could destroy.
- AlterInForeach no longer asks to unset a by-reference loop variable that
  the next statement binds by reference again (`$list[] = &$v;`), using the
  lingering reference on purpose.
- TraitsPropertiesConflicts names the parent class, which declares the
  property, when a trait re-declares a parent property, instead of the
  child class that only uses the trait.
- UndeclaredDynamicProperty no longer reports writes through `$this` in a
  closure handed to `Closure::bind()`, `bindTo()` or `call()`, where
  `$this` is the object the closure is bound to.
- ForgottenDebugOutput no longer reports a debug function called only after
  `function_exists()` confirmed that very function, such as
  `xdebug_get_function_stack()` used to build a stack trace.
- ZipOpenTruthyResult fixes write `$zip->open($f) !== true` (or
  `true !== $zip->open($f)` with the yoda style) instead of
  `($zip->open($f)!==true)`.
- ClassReImplementsParentInterface fixes on a multi-line `implements` list
  remove the line of the dropped entry instead of leaving a line of spaces.
- ReferencingObjects no longer reports test code, whose fixtures declare
  by-reference objects on purpose for tests that compare the signature.
- UnusedConstructorDependencies no longer reports test classes and
  fixtures, whose private properties tests read by reflection.
- BcryptPasswordTruncation no longer reports test code, which hashes
  over-long passwords on purpose to check the 72-byte limit.
- Failure-check rules (ImageDecodeFailureUnchecked, FileReadFailureUnchecked
  and the others sharing their flow proofs) now count a PHP function with a
  class-typed parameter, such as `imagesx(GdImage $image)`, as proof that
  the value it received was not `false`: on PHP 8 it would have thrown a
  `TypeError`. A second use after `imagesx($image)` is no longer reported.
- UntrustedHeaderValue no longer reports a header built from `json_encode()`
  output without `JSON_PRETTY_PRINT`: JSON escapes line breaks.
- SuspiciousAssignments no longer reports a write after an `if` that only
  binds a reference (`if ($key) { $out = &$byKey[$key]; } $out = $data;`):
  the write goes through that reference.
- CurlSslServerSpoofing no longer reports an array that maps cURL options to
  names (`[CURLOPT_SSL_VERIFYPEER => 'verify_peer', …]`) as if it disabled
  verification.
- PotentialMalware no longer reports `get_defined_functions()` in code that
  never calls a function by a computed name, such as an error handler
  suggesting close names for an undefined function.
- MagicMethodsValidity no longer reports a constructor that redoes all of
  its parent constructor's initialization itself, calling the same methods
  on `$this` and setting the same properties, instead of calling it.
- PhpUnitDeprecations no longer reports old assertion APIs in compatibility
  code: a polyfill method named after the replacement (whose rename fix
  made it call itself forever) or a `method_exists()` fallback for older
  PHPUnit versions.
- DisconnectedForeachInstruction no longer reports a call made after the
  loop body wrote a file or process state, such as a command run after
  `file_put_contents()` rewrote its configuration: the call may read what
  was written.
- EmptyClass no longer reports empty classes in test files and test
  namespaces, where dummy classes are empty on purpose.
- NotOptimalIfConditions no longer asks to move a check before a strict
  `in_array($v, ..., true)` that narrows the `$v` it uses, such as
  `in_array($k, KEYS, true) && array_key_exists($k, $map)`: reordered, a
  non-key value would throw.
- RecursiveTraversalOmitsDirectories no longer reports a traversal that
  keeps only files: one filtered by a `RegexIterator`, or read by loops
  that test each entry with `isFile()` or `getExtension()`.
- NullPointerException reports a nullable value once when its first
  dereference throws on null and always runs (`$box->open(); $box->close();`
  reports only the first call), instead of repeating the report for every
  later use.
- When neither `custos.json` nor `composer.json` sets a PHP version, the
  version pinned by a `.php-version` file at the project root is used
  instead of the PHP 8.4 default, so a PHP 5.6 project is no longer offered
  `??`, short list syntax or return types.
- LongInheritanceChain stops counting at the first builtin or `vendor/`
  ancestor: extending a deep framework class such as Doctrine's
  `ServiceEntityRepository` counts as one level, since the project cannot
  shorten the library's own hierarchy.
- UsingInclusionReturnValue no longer reports a file-level
  `return require 'x.php';`, which forwards to another file and passes its
  value on, as routers and configuration files do.
- ByteLengthUsedAsCharacterCount no longer reports a `strlen()` compared
  with another byte length or offset, such as `strlen($a) === strlen($b)`
  or `$this->assertSame(strlen($a), strlen($b))`.
- OffsetOperations no longer reports an element read with a dynamic key,
  such as `$days[$key]['total']`, as a scalar because another key of the
  same array holds one.
- NullPointerException no longer reports a call chained on a method
  documented with `mixed` beside `null` (`@return Box|mixed|null`).
- NullPointerException no longer reports dereferences in unreachable code,
  such as statements left after a `break` in a `case`.
- ParameterDefaultValueIsNotNull no longer asks for `null` instead of a
  `new` default value (`$clock = new SystemClock()`), which is what replaces
  a `null` default since PHP 8.1.
- ClassReImplementsParentInterface fixes also delete the `use` import of the
  removed interface when nothing else in the namespace mentions it, instead
  of leaving it unused.
- AutoloadingIssues no longer reports a script that declares a class and
  instantiates it or calls it statically in its own top-level code: no
  autoloader looks for that class.
- SubStrUsedAsArrayAccess fixes no longer wrap the replacement in extra
  parentheses when the call is a function argument:
  `strtolower(substr($s, 0, 1))` becomes `strtolower($s[0] ?? '')`.
- NestedPositiveIfStatements fixes no longer parenthesize a parent
  condition that already binds tighter than `&&` (`$a == $b && $c`), and
  move the merged code one indentation level out instead of keeping its
  old nesting depth.
- ThrowRawException no longer asks for a message when the exception's own
  constructor may supply one: a constructor that substitutes a default for
  an empty message, rewrites it or sets `$this->message` is no longer
  treated as a plain pass-through.
- UsingInclusionReturnValue no longer reports an inclusion guarded by a
  logical operator in a statement of its own, such as
  `class_exists($class, false) || require $file;`: its value is discarded.
- An `exclude` entry of several directory names, such as the default
  `var/cache`, now matches at any depth below a scanned path, like a single
  name: the generated cache of a nested Symfony application is no longer
  analyzed.
- InvertedIfElseConstructs (opt-in) no longer offers to swap the branches
  when a comment sits between them: it would end up above the wrong branch.
- ParameterDefaultValueIsNotNull (opt-in) no longer reports public and
  protected methods of a class whose parent hierarchy cannot be resolved:
  the unknown parent likely dictates the signature.
- NullPointerException (opt-in) no longer reports a read on the right side
  of the assignment that makes the variable nullable, as in
  `$node = $this->find($node->children)`: it sees the earlier value.
- EmptyClass (opt-in) no longer reports a class that holds `@phpstan-type`
  or `@psalm-type` aliases for other files to import.
- JsonEncodingApiUsage also accepts a fallback on an element of the
  decoded value, as in `$documents[0] ?? []` or `isset($doc->title)`, as a
  failure test.
- RedundantElseClause quick-fixes now move the `else` body to the `if`
  statement's indentation instead of leaving it one level too deep.
- ThrowRawException no longer offers its `\RuntimeException` quick-fix
  inside a `try` that catches `RuntimeException`: the exception would reach
  another handler.
- JsonEncodingApiUsage no longer asks for `JSON_THROW_ON_ERROR` when the
  result is tested with `empty()` (or, after `json_decode()`, `isset()`):
  the flag would turn that fallback into an exception.
- ForgottenDebugOutput no longer reports output behind an explicit debug
  switch, such as `if ($this->debugMode) { dump($row); }` or a `switch` on
  the configured debug output.
- Several checks that search an expression or a block for a match kept
  looking after finding one, and a later node could undo it: a per-iteration
  `echo` followed by another call in DisconnectedForeachInstruction, a
  `parent::__construct` callable followed by another array entry in
  MagicMethodsValidity, a varying call followed by a plain one in
  SuspiciousBinaryOperation, and the same pattern in
  LibxmlErrorBufferNeverCleared, ReferencingObjects, IncorrectRandomRange and
  the flow analysis.
- SlowArrayOperationsInLoop sees every other mention of the accumulator in
  the loop, not only the last one it visits: a merge whose target is also
  written by element in a nested loop is no longer reported.
- RandomApiMigration no longer suggests `random_int()` in a file that seeds
  the generator with `mt_srand($seed)` or `srand($seed)`: `random_int()`
  ignores the seed.
- The taint rules no longer carry request data through
  `mysqli_num_rows()`, `mysqli_affected_rows()`, `mysqli_insert_id()` and
  their `mysqli_stmt_` forms, which return a number.
- JsonEncodingApiUsage sees a failure test that follows an
  `if ($data instanceof \stdClass) { … }` on the decoded value, and no
  longer offers to make such a decode throw.
- `$this`, `self` and `static` in the constructor arguments of an anonymous
  class (`new class($this->name) { … }`) now refer to the class that creates
  it, not to the anonymous class: UnusedConstructorDependencies no longer
  reports a property passed on that way.
- LoopWhichDoesNotLoop no longer reports returning the first element of a
  property whose type is unknown, such as an untyped Doctrine relation.
- DuplicateMatchCondition no longer treats two properties or array elements
  with the same default value as one condition; its quick-fix deleted the
  second, still reachable, arm.
- FileReadFailureUnchecked no longer reports a read decoded by
  `json_decode()` with `JSON_THROW_ON_ERROR`, where a failed read throws.
- UntrustedHeaderValue no longer reports request data passed through
  `urlencode()`, `rawurlencode()`, `base64_encode()`, `bin2hex()` or
  `http_build_query()`, whose output holds no line break.
- UnqualifiedReference no longer says a backslash in a callback string such
  as `'\strval'` allows compile-time binding: a callback string always names
  a global function, and the message now says so.
- RecursiveTraversalOmitsDirectories no longer reports a traversal that
  passes `RecursiveIteratorIterator::LEAVES_ONLY` explicitly.
- PregMatchFailureAsNonMatch no longer reports the UTF-8 validity check
  `!preg_match('//u', $s)`: a pattern that matches everything can only fail.
- DateOnlyParseRetainsCurrentTime no longer reports a `createFromFormat()`
  call whose value is only tested for success.
- TempnamDirectoryFallbackUnchecked no longer reports test code, whose
  temporary files are throwaway fixtures.
- DisconnectedForeachInstruction no longer asks to move out of a loop a
  call that reads, moves, truncates or locks a stream, or that pushes or
  pops an error handler or an output buffer.
- PropertyCanBeStatic no longer reports a property that a subclass
  re-declares: making it static would be a fatal error.
- BadExceptionsProcessing no longer reports the size of a `try` block that
  has no `catch`, only a `finally`.
- EmptyClass no longer reports a class or trait declared inside a block,
  such as the empty fallback of a version check.
- EmptyClass no longer reports a class registered by a docblock annotation
  such as Drupal's `@FieldFormatter(...)` or Doctrine's `@ORM\Entity`.
- HardcodedCredentialAtKnownSink matches its `excludedPaths` globs without
  regard to letter case, so the default `**/tests/**` also skips `Tests/`
  directories.
- TypesCastingCanBeUsed no longer suggests `(string)$dir` for `"$dir"` when
  `$dir` is already a string: UnnecessaryCasting would report that cast.
- SessionMutationAfterClose no longer reports `$_SESSION = [];` after the
  session is closed: it clears the in-memory copy.
- ClassMethodNameMatchesFieldName no longer reports a method named like a
  private property of an ancestor class, and advises renaming the property
  when an interface or the parent class imposes the method name.
- SuspiciousAssignments no longer reports destructuring an `iterable`
  value, which may be an array.
- The rules that look for unchecked failures take `$this->fail()`,
  `self::markTestSkipped()` and `markTestIncomplete()` as leaving the test
  in a class that extends PHPUnit's `TestCase` when PHPUnit itself is not
  indexed, as when symfony/phpunit-bridge installs it.
- PhpUnitTests and PhpUnitDeprecations take the PHPUnit version from the
  schema that `phpunit.xml` or `phpunit.xml.dist` names when PHPUnit is not
  in `vendor/` and composer.json does not constrain it, as in projects that
  install it through symfony/phpunit-bridge, instead of suggesting
  `assertRegExp()` to PHPUnit 10+ projects.
- ForgottenDebugOutput no longer reports dumps whose output is captured
  between `ob_start()` and a later `ob_get_clean()` (or another call that
  reads or ends the buffer), nor a first-class callable such as
  `var_dump(...)`.
- UnsetConstructsCanBeMerged no longer offers to merge `unset()` calls when
  one of them has a comment after it on its line: the comment would end up
  on the merged line.
- ReturnTypeCanBeDeclared no longer offers a quick-fix on a method whose
  body can end without a `return`, such as an `if` with no `else`: PHP
  throws when the end is reached, even under a nullable type.
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
