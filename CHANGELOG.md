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
- Narrowing: negated compound conditions (`if (!is_scalar($k) && !$k
  instanceof \Stringable) throw …;` leaves `scalar|\Stringable`; the true
  branch of `A || B`); every elseif and else of a chain sees the earlier
  conditions as false; `for` conditions; `match (true)` / `match ($x)` arms
  and `switch (true)` / `switch ($x)` cases (earlier arms and cases as
  failed, `default` included); `$x?->m()`, `isset($x->p)`, `$x->p !==
  null` make `$x` non-null; `is_a()`, `is_subclass_of()`, `gettype()`,
  `get_debug_type()`, `get_class()` / `$x::class` comparisons, strict
  `in_array()`, `$x === 'lit'` / `=== Enum::Case`, `true === is_*()`;
  `count($x) > 0` also drops null; truthiness turns `bool` into `true`.

### Changed
- `custos fix` fixes files in parallel (WordPress 7.1 s → 1.6 s);
  `--stats` labels a `--php` version as "flag".
- `make fixcheck` also applies all fixes of each file together.

### Fixed
- Type engine (review round 5): `extract()`, `$$name =` and one-argument
  `parse_str()` make earlier locals unknown; variables possibly read before
  any assignment include null; the type of a variable past an
  if/elseif/else chain joins what each path leaves (and `instanceof` keeps
  only compatible members); `class_alias()` names resolve; imported classes
  named like pseudo-types (`Number`) win; `var_export()`/`print_r()` with
  `true` return strings; `assert()` narrows; the SpecOnly casting typer
  respects early exits; the ancestor cap is exposed
  (`Index.AncestorsComplete`).
- Unsafe quick-fixes (TYPO3 / MediaWiki / Moodle / phpBB / Flarum /
  Pimcore review): IsEmptyFunctionUsage rewrote `empty()` of a value typed
  only by PHPDoc (a failed lookup passed an access check);
  RealpathInStreamContext replaced `realpath($base . '/..')` by `dirname()`
  for bases that may be symlinks; StringCaseManipulation rewrote one-sided
  case conversions to `stripos()` (matches more); CascadeStringReplacement
  merged calls whose arguments read the intermediate result;
  NestedAssignmentsUsage split a chained assignment in a brace-less `if`
  body (detaching the `else`); MissingOrEmptyGroupStatement put the brace
  after a statement-ending `?>`.
- False positives (same review): GetTypeMissUse and PrintfScanfArguments
  (properties a subclass may redeclare, parameter defaults, `sscanf()`
  results used as values), MissingIssetImplementation (properties declared
  by subclasses; 16 → 6), UnusedConstructorDependencies (dynamic `$this`
  reads), OnlyWritesOnParameter (`compact($vars)`, variable variables),
  UselessUnset (unsets observed later), DisconnectedForeachInstruction
  (output, by-reference method arguments), CallableParameterUseCaseInTypeContext
  (`@` calls, false markers, invokable objects, int for float),
  StaticInvocationViaThis (`XMLReader::open()`), SuspiciousBinaryOperation
  (identical calls that may differ), PhpUnitTests (`@covers` of a
  function), IssetArgumentExistence (outer loops), ClassMockingCorrectness
  (stored builders), NotOptimalRegularExpressions (preg_quote() text,
  mandatory inner groups), OffsetOperations (`SimpleXMLElement|false`,
  loosely documented indexes), MultiAssignmentUsage (by-reference pairs),
  ReferencingObjects (hierarchies beyond the ancestor cap), PregQuoteUsage
  (delimiter-free constants); ReturnTypeCanBeDeclared adds `null` for a
  possibly unassigned returned variable and lets the `@return` tag decide
  over `mixed` results.
- OffsetOperations was quadratic on huge classes (Moodle's tcpdf.php
  4.8 s → 0.4 s).
- Unsafe quick-fixes (Firefly III / Monica / Bagisto / Dolibarr /
  Roundcube / FreshRSS review): DynamicInvocationViaScopeResolution
  rewrote `self::m()` to `$this->m()` when a subclass overrides `m()`
  (Dolibarr's printipp ran the subclass hook from the base constructor);
  NotOptimalRegularExpressions rewrote `preg_replace()` to `str_replace()`
  with replacements holding back-references or backslashes (escaped SQL
  values lost their unescaping); MkdirRaceCondition turned a `mkdir()`
  lock (`if (!@mkdir($lock)) return; … rmdir($lock);`) into a lock every
  process takes; OneTimeUseVariables dropped a nameless `/** @var T */`
  and left an indentation-only line.
- ReturnTypeCanBeDeclared and DeprecatedConstructorStyle combined into
  `__construct(): void` (fatal) on PHP 4-style constructors at PHP 8.x;
  such methods get no return type suggestion at any version.
- False positives (same review): MagicMethodsValidity (`__` names imposed
  by an interface or parent), UsingInclusionOnceReturnValue (results kept
  in success flags; 36 → 18), PrintfScanfArguments (`%ld`, `%.0lf`),
  NotOptimalRegularExpressions (`\P` searched in the decoded pattern),
  OffsetOperations (classes with unresolvable ancestors),
  IsEmptyFunctionUsage (`=== null` for file-scope variables not certainly
  assigned: Dolibarr templates' `@var` guards; 216 → 67),
  UnusedConstructorDependencies (re-entered constructors),
  PropertyCanBeStatic (properties written per instance),
  DisconnectedForeachInstruction (`var_dump()`/`print_r()` output, repeat
  loops that never read their variable).
- `custos analyse` no longer keeps the quick-fix closures of every finding
  (and with them every file's syntax tree) until the report: peak memory
  on Dolibarr (4,350 files, 273,000 findings with `--all`) 3.0 GB → 1.4 GB.
- Type engine (review round 4): `iterable` is refined by its doc type;
  generic classes given fewer arguments bind them to the templates without
  bound or default (Shopware collections) and use template defaults;
  boolean aliases (`$ok = is_object($r); if ($ok)`) and properties of
  variables (`$ref->value`, chains) narrow; overriding methods without a
  return type keep the overridden one; keys absent from an array literal
  no longer take the other elements' type; `str_replace()` on an unknown
  subject is unknown; anonymous classes are typed by their parent and
  interfaces; promoted properties read the constructor's `@param`.
- Unsafe quick-fixes (Sylius / Shopware / API Platform / Mautic / Kimai /
  Akeneo review): ReturnTypeCanBeDeclared declared `: ArrayCollection`
  for Doctrine collections (TypeError on hydrated `PersistentCollection`s
  and on repositories returning arrays); RealpathInStreamContext replaced
  a `realpath()` whose `false` result was tested (the existence check went
  dead); MkdirRaceCondition threw where a later `realpath()` /
  `file_exists()` already handled the failure; UnnecessaryCasting removed
  a cast from a loop variable of unknown type reassigned in a branch.
- False positives (same review): OffsetOperations (`iterable`,
  interfaces such as Laravel's `Application`, reads under `isset()`/`??`,
  `mixed`/`void` indexes; 227 → 106), ClassMockingCorrectness (PhpSpec
  helper methods), SuspiciousAssignments (fall-through reading the
  previous value, `func_get_arg()`), DisconnectedForeachInstruction
  (fluent chains, statements after a conditional `continue`, shared
  resets; 20 → 6), ClassOverridesFieldOfSuperClass (re-declarations
  narrowing the `@var` type), EmptyClass (attribute-configured classes),
  CallableParameterUseCaseInTypeContext (inline `@var`, `object` values,
  `match` of calls), DateIntervalSpecification (guarded fallbacks),
  ThrowRawException (constructor default messages),
  ClassMethodNameMatchesFieldName (promoted properties documented by the
  constructor), OnlyWritesOnParameter (`mixed` locals);
  SimpleXmlLoadFileUsage only reports below PHP 8.0.
- `custos fix --diff` printed `a//abs/path` headers for absolute paths;
  NotOptimalRegularExpressions said `[^\s]` → `\S` "matches more under
  /u" (same set in every mode).
- Type engine (review round 3): definitions on paths that always leave
  (return, throw, exit, break, continue) and in mutually exclusive
  branches no longer reach later reads; do-while and for back edges are
  narrowed by the loop condition; static properties (`self::$p`,
  `static::$p`) narrow like `$this->p`; a `/** @var Class $name */` hint
  over a class-name string no longer retypes it; stub PHPDoc no longer
  widens builtin return types (`substr()` is `string` on PHP 8); `@method`
  tags no longer hide real methods.
- Unsafe or invalid quick-fixes (Magento / Joomla / CakePHP / Yii2 /
  Laminas / Craft review): UnnecessarySemicolon removed the required `;`
  of a short echo tag at the end of a file; SelfClassReferencing put
  `self` into intersection types (compile error); MissingOrEmptyGroupStatement
  placed the opening brace after a line comment; MkdirRaceCondition
  inverted `A || !mkdir($d)` and threw where a later `is_dir()` already
  handled the failure; ForeachInvariants deleted a property write made
  after the loop, dropped extra header expressions and read stale values
  after element writes; RealpathInStreamContext kept a trailing `/`;
  SlowArrayOperationsInLoop hoisted the length of an array the loop
  modifies; SubStrUsedAsArrayAccess rewrote `string|int` sources.
- False positives (same review): OffsetOperations (loose
  `array|int|string|float|bool` docs, `DOMNodeList`/`ResourceBundle`;
  3,973 → 455), MagicMethodsValidity (`_construct()` hooks),
  ClassConstantUsageCorrectness (imported names written in another case),
  MissingIssetImplementation (`stdClass`/`SimpleXMLElement` subclasses,
  unresolved parents), SuspiciousAssignments (`array|bool`, `sscanf()`),
  PassingByReferenceCorrectness (`extract()`), PregQuoteUsage
  (delimiter-free literals), CallableParameterUseCaseInTypeContext
  (failure markers of method calls), GetClassUsage (guards and
  reassignments), UselessUnset (before `include`/`get_defined_vars()`),
  OnlyWritesOnParameter (element writes on possibly ArrayAccess values),
  IssetArgumentExistence (destructuring), UsingInclusionOnceReturnValue
  (`(bool) include_once`), LoopWhichDoesNotLoop (`foreach ($this)` in a
  trait), UnknownInspection (more PhpStorm names), UnnecessaryCasting
  (variables with a definition of unknown type).
- The project index honours composer.json `config.vendor-dir` (Joomla's
  `libraries/vendor` was not indexed).
- Type engine (review round 2): `str_replace()`/`preg_replace()` & co.
  follow the subject's members (a string subject is `string|null` for
  `preg_*`, never `array`); `max()`/`min()` return their arguments' type;
  PHPDoc and native intersections (`A&B`, `(A&B)|null`) stay intersections
  and find members on either side (Doctrine `matching()`); writing
  `$a[1] = …` no longer changes the type of `$a[0]`; `hash()`,
  `hash_hmac()`, `array_chunk()` and other builtins that failed with
  false/null before PHP 8.0 are typed so on 7.x.
- MagicMethodsValidity: no "does not call parent::__construct()" when the
  parent method's body is empty.
- UnnecessaryCasting: a cast is not removed when the operand's type comes
  only from PHPDoc (`@param array<int, int>`, `@var` properties, inline
  `@var`).
- Unsafe quick-fixes (phpMyAdmin / Matomo / PrestaShop / Composer /
  PHPUnit / Doctrine ORM review): NestedAssignmentsUsage split
  `$t = $list[] = x` into a read of `$list[]` (fatal);
  NullCoalescingOperatorCanBeUsed dropped `= &` references;
  TypeUnsafeArraySearch's `, true` changed results (now reported without a
  fix); InArrayMissUse `array_key_exists()` for numeric-string needles;
  ForeachInvariants on `$i <= count($a)`; StrEndsWithCanBeUsed with
  possibly empty needles; IsEmptyFunctionUsage `=== null` for
  `SimpleXMLElement`/`GMP`; RealpathInStreamContext on `/..name`
  segments; MissingOrEmptyGroupStatement and ObGetCleanCanBeUsed output
  formatting.
- False positives in OffsetOperations (736 → 322), MissingIssetImplementation
  (dynamic properties, 65 → 2), StaticInvocationViaThis (union receivers),
  CallableParameterUseCaseInTypeContext (untyped parameters),
  UnusedConstructorDependencies (attributes), LoopWhichDoesNotLoop (empty
  loops over iterators), MagicMethodsValidity (inherited `_get`),
  PhpUnitTests (fully qualified `@covers`), PregQuoteUsage, MkdirRaceCondition,
  SuspiciousAssignments, PropertyInitializationFlaws,
  DisconnectedForeachInstruction, CryptographicallySecureRandomness (PHP
  7.4+), UsingInclusionOnceReturnValue, OnlyWritesOnParameter
  (anonymous-class arguments).
- IsEmptyFunctionUsage crashed on a recovery tree (fuzz).
- A variable assigned in a `switch` case ending with `break` no longer
  reaches the later cases (only through an enclosing loop).
- Type inference (from the WordPress/Drupal review): builtin return types
  follow the target PHP version (`substr()` is `string|false` before 8.0);
  `while ($x = f())` / `if ($x = f())` narrow `$x`; an unconditional
  reassignment hides earlier values in the casting typer; an
  `-assert-if-true` / `-if-false` assertion is negated in the other branch
  (`is_wp_error()` guards); `$_SERVER['SERVER_PORT']`/`['REMOTE_PORT']` are
  strings; an absent key of a literal array is unknown; `array_reduce()`
  is the initial value's type with the callback's; a function declared
  twice (`apply_filters()` and its no-op) returns either declaration's
  type; `#[\AllowDynamicProperties]` declared in another file is seen
  (MissingIssetImplementation).
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
