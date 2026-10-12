# Changelog

All notable changes to custos. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `IsIterableCanBeUsed` offers a quick fix for an adjacent
  `is_array($v) || $v instanceof Traversable` pair on a plain variable.
- Security inspections cover more sinks: interpolated, `print` and SVG/XHTML
  output, the backtick operator and string-form `proc_open`, remote stream
  reads and `curl_setopt_array` destinations, LDAP filters escaped for the wrong
  context, bit-or permission modes and cURL basic authentication set through
  `CURLOPT_URL`.

### Fixed

- Quick fixes no longer produce invalid PHP around adjacent keywords
  (`return!is_null(...)`), closing `?>` tags, line comments, heredocs below
  PHP 7.3 and attribute closures, and they keep comments inside the rewritten
  range or are withheld.
- Quick fixes add the parentheses needed to preserve operator precedence, and
  are withheld (the finding is still reported) when the rewrite would duplicate,
  drop or reorder side effects.
- Code-style fixes no longer change behaviour: postfix increments where the
  value is used, compound assignments over `+`/`*` chains, `self` in `catch`
  clauses, `static::` inside string interpolation, modifier words read from
  comments and `parent::` calls overridden through traits.
- Control-flow and performance rewrites are withheld when the result could
  differ: multibyte case conversions, non-literal `strtr`/`stripos` operands,
  merged `isset`/`str_replace` calls, loop-length caching, reordered conditions
  and `foreach` conversions whose counter stays observable.
- Language-migration fixes handle namespace blocks and clashing imports,
  case-sensitive class constants, version-suffixed `PHP_VERSION` comparisons,
  return types that could throw at runtime and casts that need quoting or
  parentheses; `rand()`/`mt_rand()` are reported without an automatic rename.
- PHPUnit assertion rewrites no longer change what a test checks, and ignore
  same-named methods of non-PHPUnit classes.
- Calls using named arguments, argument unpacking or first-class callable
  syntax are no longer misread by argument-position checks.
- Variables that can change through references, by-reference arguments,
  `global`, `extract()`, variable variables, destructuring or `foreach` targets
  are no longer treated as stable by type, cast, unused-code and security
  checks.
- Architecture and unused-code inspections respect property hooks, abstract
  and promoted properties, trait methods, `__sleep()` lists and PHP-version
  differences of built-in classes.
- Probable-bug inspections that reason about a known value (sizes, counts,
  indexes, nonces, handles, statements) no longer trust it after `++`/`--`,
  a by-reference capture or argument, destructuring, a `foreach` target,
  `extract()`, a variable variable or a possible replacement in a closure.
- Failure-check inspections (files, streams, sockets, images, regexes, intl,
  XML, databases, processes) accept more guard forms: early exits, `elseif`
  chains, `&&`/`||` operands, ternaries, truthiness, `instanceof`, type
  predicates and constant-first comparisons; nullsafe chains and
  first-class callable syntax are no longer read as checked calls.
- HTTP and cookie inspections compare option and header names
  case-insensitively and ignore text inside quoted parameter values; PDO,
  mysqli and SQLite inspections tolerate SQL comments, restored connection
  state and statements bound elsewhere.
- Several probable-bug inspections now report equivalent spellings they
  missed: constant-first comparisons, brace-less and parenthesised forms,
  `new self()` in abstract classes and guards that only `continue` or `break`.
- Probable-bug quick fixes are withheld when the rewritten value may have
  changed since it was assigned or when the matched text sits inside another
  quoted value.
- Security inspections recognise more validation idioms (early exits,
  combined conditions, allow-lists, server-side MIME inspection, `realpath`
  containment) and no longer report numeric, boolean or `match`-selected
  constant values as tainted.
- Value-based inspections no longer trust a variable's assigned value after
  a `foreach`, `list()` or `catch` target, `extract()`, a write
  through a reference alias, or a call that can run code holding a reference
  to it (by-reference closure capture, `global`, by-reference parameter);
  copies taken afterwards are unknown too. Request data the variable held, or
  that the reference holder writes to it, is still reported. A known integer
  stepped with `++`, `--`, `+=` or `-=` on the only path to its use has the
  stepped value; an unrelated property, element or compound assignment no
  longer discards what is known about other variables. A write through
  `$GLOBALS['name']`, or through a property, static property or array slot
  holding a reference to a local, is seen as a write to that variable, and
  `$a = &$b` makes both names read the same value, until `unset`, `static` or
  `global` gives one of them other storage. Long functions (hundreds of
  locals and branches) no longer exhaust the analysis budget, and deeply
  nested loops or a `switch` in such a function no longer make its analysis
  take seconds.
- Project analysis no longer rebuilds the whole cross-file summary table for
  every analysed file, which took about a third of the time on large projects,
  and no longer analyses a file again when nothing it calls changed.
- Malformed input (`curl_exec(` at end of file, `new` without an argument
  list inside an array access) no longer makes four inspections skip the file.
- Request-data inspections no longer treat a `match` subject, arithmetic
  results or number- and boolean-returning builtins as carrying request text;
  the short ternary `$a ?: $b` now keeps the request data of `$a`. Request
  data is now reported when it is combined with a value of unknown origin
  (`$this->bin . $_GET['a']`), appended with `.=`, stored in an array element,
  iterated with `foreach`, imported by `extract($_GET)`, or passed to a
  function that cannot be resolved before it reaches the sink. Array elements
  with a constant key are tracked individually and follow the array into
  another variable, so a clean element next to a request-derived one is not
  reported; keys are compared as PHP folds them (`0x1`, `01`, `1.7`, `true`
  and `'1'` are the key `1`). `sort()`, `preg_match()` groups and
  `parse_str()` results keep the request data they work on, while counts,
  exit statuses and other numeric results a function writes to a reference
  argument, the output of `exec()` and an array sorted with a request-supplied
  callback are clean. `extract()` honours `EXTR_SKIP`, also combined with
  `EXTR_REFS` or passed as a named argument, and cannot create superglobals;
  `settype($v, 'int')` and `$n += ...` on a number yield clean values.
  `array_map('intval', ...)` yields clean values but keeps the request's keys
  (`array_keys()` of the result is still reported). The server-chosen fields
  of an uploaded file (`size`, `tmp_name`, `error`) are trusted only on the
  upload entry itself, not on an array a client field was copied into. A value
  passed to a variadic PHP function (`sprintf`, `max`) keeps what was known
  about it.
- `UnescapedHtmlOutput` no longer stops at the first dynamic part of an
  output statement: request text emitted after an escaped value, a number or
  a boolean (`echo '<i>' . htmlspecialchars($a) . '</i><b>' . $b . '</b>'`,
  `"<i>$page</i><b>$name</b>"`) is reported.
- Request-data inspections keep the request data of a variable bound to a
  statement parameter (`mysqli_stmt::bind_param`, `PDOStatement::bindParam`),
  report client-chosen upload fields at any depth
  (`$_FILES['f']['name'][0]`), and see a write through a reference to an
  array element in copies of that array.
- Failure-value inspections that follow a local back to the call that
  produced it (`parse_url`, `scandir`, `posix_getpwnam`, `realpath` and the
  like) no longer let a guard vouch for a later assignment: a result
  reassigned in a loop after an early-exit check, or on one branch only, is
  reported when that assignment is unchecked. `isset($v['key'])`,
  `!empty($v['key'])` and `array_key_exists('key', $v)` count as checks of
  `$v`, as do the condition of a `for` loop and a fallback chain tested in
  place (`($u = a()) === false && ($u = b()) === false`).
- Failure-value and request-data inspections understand more ways of
  ruling a failure out: a statement that leaves on failure
  (`$h !== false or exit;`, `$h = fopen($f, 'r') or die();`,
  `$v ?: throw …;`), a failing branch that ends in a call that never returns
  (a callee declared `never` such as PHPUnit's `fail()`, or a project function
  or final/private method that always throws or exits), a test stored in a
  boolean (`$ok = $v !== false; if (!$ok) { return; }`) while neither
  variable is written, and, for PHP 8.0 and later, a builtin that only accepts
  arrays or objects having returned (`count($v)` throws on `false`).
  Positions PHP may skip (the right side of `??=`, arguments of a null-safe
  call) prove nothing.
- Functions of about two thousand lines and very long `elseif` chains no
  longer exhaust the flow analysis budget.
- `FileReadFailureUnchecked` accepts the same failure checks as the other
  failure-value inspections (`$s !== false or exit;`, a test stored in a
  boolean, a branch ending in a never-returning call, an early exit that also
  has an `else` branch, loop conditions).
- Passing an element of an array to a function or method (`$o->m($parts['host'])`)
  no longer discards what is known about the array itself: an earlier failure
  check still holds, and request data in its other elements is still reported.
  Elements beside the one passed keep what was known of them.
- Functions containing a `switch` are now followed by the flow analysis
  (subject once, cases in order, fall-through, `default` anywhere, `break`
  and `continue` with a level): value, failure-check and request-data
  inspections report there as everywhere else. Only `goto` still leaves a
  function unfollowed.
- Request data a project function writes to a by-reference parameter
  (`function fill(&$x) { $x = $_GET['a']; }`) reaches the caller's variable,
  across files; a function that overwrites the parameter with clean data on
  every path clears it, and one that may leave it alone keeps what it held.
- A silenced call (`$s = @file_get_contents($f)`) is followed like the call
  itself by the inspections that track a result back to its producer.
- `ForkChildFallsIntoParentWork` counts the levels of `break 2`/`continue 2`
  instead of assuming any level leaves the child block.
- Statements written before a `do … while` loop are no longer analysed as
  part of its body: a read before the loop does not see values assigned in
  it, and a variable stepped in the loop is no longer taken for its value
  after one iteration.
- Failure-value inspections accept more guards: truthiness, `!empty()`, loose
  comparison with `false`/`null`, type predicates, constant-first comparisons,
  an assignment tested in place, `elseif`/`else` arms, short-circuit operands,
  ternary arms and an early exit that also has an `else` branch.
- Handle and object inspections notice when a tracked local is rebound by
  destructuring, a loop target, a by-reference argument or `unset`, no longer
  count a call in another `match` arm or behind a null-safe access as always
  executed, and resolve `new self`.
- Operand comparisons keep keyword-spelled constant names and named-argument
  labels case-sensitive (`A::Array` is not `A::ARRAY`); `\X41` hex escapes are
  decoded; `exit()` with empty parentheses ends a response with a known body.

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
