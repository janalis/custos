# Decisions log

Decisions taken during the port. "user decision" marks choices made by the
project owner; everything else is an engineering default that can be revisited.
Last updated: 2026-10-07.

## Process

- **Clean-room, MIT** (user decision). EA (LGPL-2.1) is never copied:
  specs are written in our own words by a separate step (`spec-rule` skill,
  the only role allowed to read EA); implementers work from specs only
  (`implement-rule`); EA fixtures are read only at test time from a local
  checkout. `make cleanroom` scans the repo for verbatim EA text.
- **Go, own parser** (user decision).
- **Editor integration via LSP** (user decision): `custos lsp`, run by the
  editor as a supplementary PHP server next to its main one.
- **Correctness over upstream fidelity** (user decision): custos diverges
  from EA when upstream behaviour is wrong (false positives, fixes that change
  behaviour or produce invalid PHP, non-existent APIs). Each divergence is
  documented in the rule spec's *Divergences* section; affected EA
  conformance cases are listed in `testdata/ea-divergences.json` with the
  reason. `docs/divergence-candidates.md` inventories the "kept upstream
  bug" items found in the specs (174: 46 class A — fix changes behaviour or
  produces invalid PHP, 59 class B — false positives, 69 class C — false
  negatives/cosmetic). **Classes A and B are done** (2026-10-07): all 46 A items
  and 58 of 59 B items fixed (one B item proved correct upstream behaviour:
  ClassConstantUsageCorrectness keeps reporting wrong-case imports because
  `::class` uses the import as written). **Class C is done too:** 61 of 69
  fixed, 8 declined with reasons recorded in each spec's Divergences (e.g.
  needs flow analysis, would add false positives, or has no visible effect).
  Follow-ups found while fixing (StrlenInEmptyStringCheckContext precedence,
  PrintfScanfArguments mixed formats, case-sensitive function names in three
  more rules) are fixed as well.
- **Name matching audit** (2026-10-07): every rule now matches function,
  method and class names, keywords (self/static/parent, `::class`), magic
  constants and true/false/null case-insensitively; variables, properties and
  other constants stay case-sensitive. Builtin functions are matched only when
  the call resolves to the global function (`ctx.GlobalFunctionName`,
  `ctx.IsGlobalFunctionCall`, which now also honours same-named functions
  declared in the current namespace). 88 rules fixed, each with a
  "custos diverges" spec entry; no EA conformance case changed outcome.
- **Commits** (owner-approved, 2026-10-07): custos initial import `ebd152f`.

## Conformance harness

- EA test cases without an explicit language level run at **PHP 5.6**
  (pinned by fixture evidence: DeprecatedIniOptions, ArgumentUnpacking,
  AccessModifierPresented).
- Parsing uses PHP 8.5 in *permissive* mode (PhpStorm parses every construct
  at any level, incl. legacy `$s{0}` and `<?` short tags); rules are gated at
  the case level.
- Fixes: all quick-fixes of the first analysis are applied once (IDE "apply
  all" semantics, overlapping edits dropped); output compared
  whitespace-collapsed.
- Files configured earlier in the same EA test method are indexed as
  companions (they stay in the IDE project).
- Only expectations carrying upstream's message prefix are compared (IDE
  parse-error markers in fixtures are ignored); message text itself is never
  compared.
- Clean room: for EA cases, fix mismatches report only the divergence offset
  and custos' own output (`HideExpected`), never upstream's expected text.
- Own fixtures (`testdata/rules/<ID>/`) are the CI gate and default to PHP 8.4;
  option sidecars accept JSON bools/numbers/strings.

### Listed EA divergences (`testdata/ea-divergences.json`)

Each entry is an EA case where custos intentionally differs; the reason and
the correct behaviour are in the rule spec's *Divergences* section.

| Rule | Fixture | Reason |
|---|---|---|
| ArgumentUnpackingCanBeUsed | `argument-unpacking.php` | below PHP 8.0 no fix when the array may have string keys (unpacking would throw) |
| ArrayIsListCanBeUsed | `array_is_list.php` | loose ==/!= forms are not reported: they ignore key order and are not equivalent to array_is_list() |
| CascadeStringReplacement | `cascade-str-replace.74.php` | array-typed arguments are spread via ...array_values() so string keys neither throw (<8.1) nor collide (>=8.1) |
| ClassConstantCanBeUsed | `class-name-constant-ns.php` | get_parent_class() in a class without extends is not rewritten to parent::class (fatal error) |
| ConstantCanBeUsed | `constants-usage.php` | get_class() outside a class is not rewritten to __CLASS__ (empty there); version_compare(..., '7.1', '==') not rewritten (patch level makes it non-equivalent) |
| DuplicateArrayKeys | `duplicate-array-keys.php` | keys are compared as PHP stores them (escapes decoded, '7' == 7, any integer base); duplicate integer keys are reported too |
| ForeachInvariants | `foreach-invariants.php` | upstream's IDE formatter re-spaces an untouched inner for header (custos copies the body verbatim); additionally a limit variable reused by several loops is resolved from the assignment reaching each loop (one extra report); loops whose counter or header-assigned limit is mentioned after the loop are not reported (two fewer reports) |
| GetDebugTypeCanBeUsed | `get_debug_type.php` | reported without a fix: get_debug_type() names differ from gettype() for scalars (int/integer, float/double, null/NULL) |
| InstanceofCanBeUsed | `instanceof-can-be-used.php` | only exact equivalents get a fix (is_a, get_class on a final class, class_implements with an interface literal on an object); get_parent_class/is_subclass_of/class_parents and non-final get_class are reported without a fix |
| InvertedIfElseConstructs | `if-inverted-condition-else-normalization.php` | for a bool operand the fix emits x() instead of false !== x() |
| MkdirRaceCondition | `mkdir-race-conditions.php` | the or-form re-check emits || is_dir($concurrentDirectory) (upstream negates it, inverting the logic) |
| NotOptimalIfConditions | `if-instanceof-flaws-false-positives.php` | under && the broader (redundant) instanceof is reported, not the more specific one |
| NotOptimalIfConditions | `if-optimal-conditions.php` | method/static calls are not reordered (they may have side effects); isset($x[...]) && $x is not reported (isset also guards $x itself) |
| PhpUnitDeprecations | `deprecations.phpunit91.php` | fix emits PHPUnit's real method names (assertFileDoesNotExist / assertDirectoryDoesNotExist) instead of upstream's misspelled ones |
| PhpUnitTests | `assert-resource-exists.php` | below PHPUnit 9.1 the fix suggests assertFileNotExists/assertDirectoryNotExists (the DoesNotExist names do not exist yet) |
| ReferencingObjects | `referencing-objects.php` | upstream's IDE formatter re-spaces an unreported `array& $arr` parameter; custos edits only the reported ranges |
| ReturnTypeCanBeDeclared | `return-type-hints.php` | no suggestion when the declaration would be a compile error (": void" with return $x where $x is null/void-documented, "?T" with a bare return) |
| SenselessProxyMethod | `senseless-proxy-signature.php` | an override that calls the parent without return, where the parent returns a value, is not reported (removing it would change return values) |
| SlowArrayOperationsInLoop | `slow-array-operations.for-termination.php` | the generated limit variable avoids names already used in the scope ($iMax1 instead of overwriting $iMax) |
| SubStrUsedAsArrayAccess | `substr-used-as-index-access.php` | no fix below PHP 7.0 (?? does not parse); ?? '' guard from 7.0; negative offsets other than -1 skipped (strlen($s) - n can go negative) |
| SubStrUsedAsStrPos | `substr-used-as-strpos.php` | 4-argument mb_substr fix emits mb_strpos($h, $n, 0, $enc); case-folded comparisons only against literals already in folded case |
| TraitsPropertiesConflicts | `traits-properties-conflicts.php` | an own property incompatible with the trait's (different default, visibility, static, readonly or type) is reported as an error: PHP refuses to compose such a class |
| UnnecessaryCasting | `unnecessary-casting.php`, `unnecessary-casting.php8.php` | an untyped private property without default (not set by the constructor) holds null: casting it is not redundant |
| UnnecessaryAssertion | `unnecessary-assertion.php` | assertInternalType is reported only when the declared return type always satisfies the named type; unknown/contradicting type names are not reported (that call fails, it is not redundant) |
| VariableFunctionsUsage | `variable-functions-php54.php` | calls with call-time & arguments are not reported/rewritten ($fn($a, &$b) is a fatal error since PHP 5.4) |

## Rule-level decisions

| Rule | Decision |
|---|---|
| PhpUnitDeprecations | Real PHPUnit method names in the fix (listed divergence). |
| SubStrUsedAsStrPos | 4-argument `mb_substr`: emit `mb_strpos($h, $n, 0, $enc)`; upstream puts the encoding in the offset slot (listed divergence). |
| ReferencingObjects, ForeachInvariants | No IDE-style reformatting of untouched code (listed divergences). |
| UnsetConstructsCanBeMerged | The second statement's fix merges the whole run (single-pass fixing cannot apply several insertions at one point). |
| UnnecessaryIssetArguments | Edit choice adjusted to avoid overlapping comma edits. |
| ClassReImplementsParentInterface, CascadeStringReplacement | Fixes compute the combined result so single-pass "apply all" gives the same output as repeated fixing. |
| NullPointerException | Commented out upstream; ported as experimental, disabled by default. |
| PhpUnitTests | Regex assertion names follow the configured PHPUnit version (`assertMatchesRegularExpression` from 9.1). A stricter "only inside TestCase classes" check was tried and reverted: it would have skipped every upstream case, losing all regression coverage for a rare false positive. |
| Test-file detection | Spec definition everywhere: path ends with `Test.php`, `Spec.php`, `.phpt` or contains `/Fixtures/` (case-sensitive) — `ctx.IsTestFile()` (`analysis.IsTestPath`); "test file or test class" is `util.InTestContext`, other test-class checks stay per spec. |
| SecurityAdvisories | Runs on `composer.json`; the CLI discovers `composer.json` when the rule is enabled (`FilePatterns`). |

## Spec-level false positives (applied as divergences)

Found on real code (corpus A, Symfony) and fixed through spec → implement.

| Rule | Was | Now |
|---|---|---|
| ForeachInvariants | Reported loops whose counter/limit is written elsewhere; suggested `foreach` changed behaviour. | Skipped when counter or limit is written outside the init/step. |
| NotOptimalIfConditions | Reordering advice ignored side effects. | No reorder when an operand calls a non-pure function (purity list), any method/static/nullsafe call or `new`, passes by reference, or uses include/eval/exit/print (listed divergence). |
| BadExceptionsProcessing | Caught variable used after the try reported as discarded. | Uses later in the scope count. |
| MultiAssignmentUsage | Nested destructuring (under a type guard) reported. | Only destructuring directly in the loop body. |
| OnlyWritesOnParameter | Object captured by `use` reported as lost writes; unused imports reported when the closure `include`s a file. | Object-typed imports skipped for write-only findings; no unused-import finding when the closure includes/requires. Also fixed an implementation bug (array-literal values counted as writes). |
| Value discovery (shared, ~20 rules) | `++`/`--`/compound assignments ignored (`$n = 0; ++$n; mt_rand(1, $n)` → min > max). | Such variables yield an unknown value set; consumers stay silent. |
| SuspiciousAssignments | `false` in a type with arrays reported as non-destructurable. | `false` treated like `null`; engine types `hrtime()` by argument. |
| MkdirRaceCondition | Random temp-dir creation still asked to re-check `is_dir`. | Kept (low impact). |
| ClassMockingCorrectness | Final classes reported in projects using dg/bypass-finals (528 on corpus B). | FINAL skipped when `\DG\BypassFinals` is indexed. |
| PhpUnitTests | `@covers ::m` under `@coversDefaultClass C` reported; regex advice named `assertRegExp()` on PHPUnit 10+. | `::m` resolved as `C::m` first; unset `PHP_UNIT_VERSION` inferred from the indexed PHPUnit (also PhpUnitDeprecations). |
| UnnecessaryCasting / CallableParameterUseCaseInTypeContext (T-rules) | `(string) preg_replace(…)` reported as redundant; guards ignored; `x ?: y` kept `false`. | `preg_*` results include null; variable types narrowed by guards (removal only, sets are partial); `?:` drops false. |
| ReturnTypeCanBeDeclared | `: void` for `return null;`, `?T` with bare `return;` (fix = compile error). | Skipped (D14); EA case `return-type-hints.php` now differs (not yet listed in `ea-divergences.json`). |
| ProperNullCoalescingOperatorUsage | `iterable ?? []` reported. | `iterable` expands to array + `\Traversable`. |
| TraitsPropertiesConflicts | Trait property re-declared with `#[ORM\…]` attributes reported. | Compatible attributed re-declarations skipped (errors kept). |
| NotOptimalIfConditions | `$o instanceof X && $o !== $other` reported. | D3 only for comparisons with literals. |
| StaticInvocationViaThis | Symfony `$this->assertResponseIsSuccessful()` etc. reported. | `EXCEPT_PHPUNIT_ASSERTIONS` also covers `Symfony\Bundle\FrameworkBundle\Test\`. |
| OnlyWritesOnParameter | Variables read via `compact('v')` / `get_defined_vars()` reported. | Counted as reads. |
| PropertyInitializationFlaws | Default reported as always replaced after an early `return`. | Pattern O skipped after a `return`. |

**corpus C / corpus D review (2026-10-07).** First run on corpus C (506 files, PHP
8.0, default + `--all`), corpus C/vendor and corpus D/vendor (sampled); every fix
also applied with `custos fix --all` on a copy (`php -l` clean).

| Rule | Was | Now |
|---|---|---|
| ReturnTypeCanBeDeclared | `: int` for `return $this->id;` over an untyped `@var int` property without default (TypeError on a null/nullable column); `: \never` for `return exit();`; `: {array}` from `@return {array}`. | Untyped properties without non-null default and not set by the constructor add `null` (`?int`, D5b); `never`/`parent` are non-suggestible built-ins; malformed doc types are ignored. 1915 → 1916 findings, ~360 messages now nullable. |
| UnnecessaryCasting | `(float) ($cell * 100)` with a `string|null` cell reported (string × int is an int); `(float) $this->cout` over an untyped private property without default reported (null → 0.0). | Sound arithmetic typing (unknown operand → unknown, numeric strings → int\|float); implicit `null` for untyped private properties without default (listed divergences `unnecessary-casting.php`, `unnecessary-casting.php8.php`). corpus C 16 → 12. |
| CallableParameterUseCaseInTypeContext | `$ttl = time() + $untyped` reported as float; `$text = mb_convert_encoding($text, …)` reported as array. | Sound arithmetic here too; `mb_convert_encoding()` typed by its input. |
| OffsetOperations | Offset access on an unresolvable class (`new Highchart()` in the wrong namespace) reported; `{array}`/`Foo::*` doc types became classes. | Unresolvable class empties S (D1). corpus C `--all` 59 → 1. |
| NullPointerException | `if ('v' === $node->name) { $node->x }` still reported. | `X === E` / `X == E` with `X` a chain rooted at the variable proves non-null (U12). |
| PreloadingUsageCorrectness | Symfony `config/preload.php` (`require …/App_KernelProdContainer.preload.php`) and `vendor/autoload.php` rewritten to `opcache_compile_file()` (disables preloading). | Inclusions whose path mentions `autoload`/`preload` skipped (E3). |
| MkdirRaceCondition | Fix dropped the `@` of `@mkdir(...)` (warning in the very race it handles). | `@` kept in every fix form. |
| NotOptimalRegularExpressions | `'/\[entité\]/'` without `/u` reported as an error. | Only non-ASCII in a class, before a quantifier, or a letter under `/i`. |
| MagicMethodsValidity | Always-throwing `__toString()` reported; `return call_user_func(…)` reported as `got 'mixed'`. | Bodies that cannot complete are skipped; `mixed` treated as unknown. |
| MockingMethodsCorrectness | Methods added with `getMockBuilder(…)->addMethods([...])` reported as missing (5 on corpus C). | `addMethods` honoured like `setMethods`. |
| LoopWhichDoesNotLoop | `while (@ob_end_flush()) {}` reported. | Empty bodies reported for `foreach` only. |
| ClassOverridesFieldOfSuperClass | `protected $table = 'invoices';` / `protected bool $skipScalars = true;` (default overrides) told to drop the re-declaration. | Only re-declarations with the same default. corpus C/vendor 371 → 58, corpus D/vendor 141 → 29, corpus B 36 → 30. |
| PropertyInitializationFlaws | Typed `private array $items = [];` default removed (uninitialised for objects built without the constructor); static re-declarations (own storage) reported; S told to remove the default (which yields null, not the inherited value). | Pattern O skipped for typed properties; S skipped for static properties and its message now says to drop the re-declaration. corpus D/vendor 62 → 23. |
| RedundantElseClause | Moved code started at column 0. | Indented like the `if` (cosmetic). |

**Coverage audit (2026-10-07).** Own fixtures were extended until every
statement of `internal/rules` is executed by `TestOwnFixtures` (90.3% →
99.99%; the single remaining statement, SecurityAdvisories'
`FilePatterns()`, is called only by the CLI and is covered by a unit test).
Unreachable branches were removed (nil/zero-span guards on nodes the parser
always fills, re-checks already implied by `GlobalFunctionName`, fallbacks
after lookups that cannot fail); guards against error-recovery trees stay and
are each exercised by a broken-PHP fixture (`broken*.php`, `recovery.php`,
`bad-*/composer.json`). Bugs found in previously untested branches (each with
a spec Divergences entry and fixtures; no EA conformance outcome changed;
findings on corpus A `src/` and corpus B unchanged):

| Rule | Was | Now |
|---|---|---|
| ArgumentUnpackingCanBeUsed | `call_user_func_array('', $a)`, `'my fmt'`, `'1fmt'`, `'Cls::m'` rewritten to invalid calls. | Only identifier paths (optionally `\`-qualified) are reported. |
| NullCoalescingOperatorCanBeUsed | `!empty($c) ? $c::$p : null` → `$c::$p ?? null` (throws for `''`/null: `??` does not guard the class lookup). | Static form only when the probe is known to hold objects. |
| ForeachInvariants | Fix dropped the braces of interpolated `{$c[$i]}abc` / `"$c[$i][0]"` (reads another variable / an offset); counter or header-assigned limit read after the loop (search loop + `if ($i == count($a))`, `echo $n`) still rewritten; fix panicked on an unterminated string at end of file (fuzz). | `{$iValue}` when the string continues the name; such loops not reported (D8c/E9, listed divergence updated: two fewer EA reports); EOF guarded. |
| StrTrUsageAsStrReplace | Heredoc/nowdoc arguments decoded with quoted-string escapes (`\'` in a nowdoc is two characters). | Heredoc: double-quoted rules except `\"`; nowdoc: no escapes. |
| StringNormalization | D4 trim lists: heredoc/nowdoc always accepted, `b'…'` taken for a heredoc, interpolated strings judged by source text, ranges (`'@..Z'`) covering letters ignored. | Bodies decoded, interpolation skipped, letter ranges count as letters. |
| TypeUnsafeComparison | Class with an unresolvable parent/interface/trait reported as lacking `__toString()`. | Skipped (the missing ancestor may declare it). |
| PropertyInitializationFlaws | `protected $kind = self::KIND;` in a child overriding `KIND` reported as repeating the parent default. | `self`/`parent` resolved against the declaring class. |
| CurlSslServerSpoofing | Destructuring keys (`[CURLOPT_SSL_VERIFYPEER => $peer] = $opts`, `list()`, foreach targets) reported as option writes. | Destructuring patterns ignored (E2). |
| SecurityAdvisories | Escaped JSON keys (`"roave\/security-advisories"`) not recognised (`strconv.Unquote`). | Decoded with `encoding/json`. |
| MockingMethodsCorrectness | Heredoc/nowdoc method names with an indented closing marker kept their indentation (reported as missing). | Indentation stripped (D9). |
| UnnecessaryAssertion | `assertInternalType('resource', f())` with `f(): resource` (a class named resource) reported. | `resource` never reported. |
| StaticLambdaBinding | `#[Pure] static fn () => $this` missed (looked for `static` as the first token). | `static` searched after attributes. |
| MkdirRaceCondition | `mkdir(...);` (first-class callable) reported; fix produced `mkdir($concurrentDirectory = ...)` (parse error). | Skipped. |
| SuspiciousBinaryOperation | D6 never fired for a declared return type without `@return` (`types.Union` with an unknown side is unknown). | Unknown parts dropped, as the spec says. |
| UnnecessaryCasting | Objects of user classes named `Integer`/`Boolean` counted as int/bool (cast removed). | Only the scalar types. |
| CascadeStringReplacement | "All searched items identical" compared keys, not values (`['x'=>'a','y'=>'a']` missed, `['-1'=>'a']` fixed to search `'-1'`). | Values compared. |
| DisallowWritingIntoStaticProperties | Methods of anonymous classes never checked. | Checked (incl. used traits). |
| IsEmptyFunctionUsage | `empty($a ?? $b)` → `$a ?? $b === null`. | Low-precedence subjects parenthesised. |
| IsNullFunctionUsage | No outer parentheses: `'v=' . is_null($x)` changed meaning, `is_null($x) == $y` became a parse error. | Wrapped in tighter contexts (and include/print/yield operands). |
| StaticClosureCanBeUsed | Closures assigned to a property, array element, static property or via `??=` reported (may be bound later). | Treated as escaping. |
| NotOptimalRegularExpressions | D22 recommendations unimplemented: numeric comparisons misread (`1 > C`, `C <= 0`, `== 2`), calls inside arithmetic/casts/`@` rewritten, missing parentheses, D22d replaced the enclosing comparison, `A`/anchored-`m` patterns, unsafe explode/trim characters. | Implemented as specified. |

Coverage is kept at 100%: `make coverage` (part of `make verify`) fails on any
statement of `internal/rules` not executed by own fixtures or the rule
packages' tests, and on any statement of `cmd/`, `internal/` and `tools/` not
executed by the whole test suite (the EA run and local corpora do not count;
the only exemptions are each command's one-line
`func main() { os.Exit(run(...)) }` wrapper, listed by file:line in the
Makefile's `COVER_MAINS`; an exemption that matches no uncovered block fails
the gate). The generators under `tools/` keep their logic in `run` functions
taking the repository root or input/output paths, tested on temporary trees
with invented inputs (fake plugin layout, stub PHP, specs, "upstream" text),
so the gate needs neither the EA checkout nor phpstorm-stubs. Refactoring them
(2026-10-07) changed no generated artifact (`make extract`, `go generate`,
`make stubs` decoded and compared, `make cleanroom`); it fixed `genstubs`
ignoring the output file's `Close` error and `cleanroom` truncating reported
text inside a multi-byte character. Complexity-regression tests use
`testbudget.Of(d)`, which scales the limit by a CPU calibration measured at
test time, so a loaded machine does not fail them while a quadratic or
exponential regression (seconds to minutes) still does; the LSP edit-latency
budget (p95 < 30 ms) is enforced only by `make bench` (`CUSTOS_PERF=1`).
`runner.BuildIndex` recovers a crash while indexing one file (that file's
symbols are dropped) instead of taking down the CLI or the language server. A test that
guards a complexity bound compares timings at n and 4n (linear ≈ 4×,
quadratic ≈ 16×) where a fixed limit proved flaky (`TestAncestorsBounded`
under coverage instrumentation); `make coverage` prints the failing tests
instead of hiding them.

NotOptimalRegularExpressions string-API rewrites made exact (2026-10-07):
`^T$` without the `D` modifier also matches `"T\n"`, so the identity
comparison (D22a) and `$`-anchored trims of a character other than `\s`
need `D`; `\s` trims spell PCRE's set (`" \t\n\r\v\f"`, which differs from
trim()'s default `\0`); trims accept only the neutral modifiers `D`/`S`
(and `i` for caseless characters) — `U` makes `+` lazy, `x` ignores
whitespace. Two upstream fixtures expect the old rewrites (listed
divergences); no finding changed on corpus A, corpus B, corpus C or corpus D.

Parser bug reported by the audit: a close tag right after a control header
(`if ($a) ?>html`, also `while`/`for`/`foreach`) was a syntax error; PHP
treats it as `;` (empty body, the HTML is the next statement). Parsed as an
empty `Nop` body now; MissingOrEmptyGroupStatement's fix inserts `{} ` before
the tag instead of wrapping it (which produced unbalanced braces).

Harness: own-fixture `.json` sidecars gained `companions` (extra `.inc`
files indexed with the fixture), `calls` (EA list-option calls such as
`registerCustomDebugMethod`) and JSON-array list options.
Engine/infer: `Analyze`'s retry loop has no unreachable exit; option
accessors, severity overrides, rule selection, panic propagation and nested
suppressions have unit tests; `resolveAsserts`, `assertIs` (mixed is never
asserted) and `computeProp` (initial value always present) lost unreachable
branches; method templates, assertions and untyped-property writes gained
edge-case tests. The parser gap the audit reported (`if ($a) ?>html<?php`
rejected) is fixed; see the close-tag note above.

## Engine

- **Parsing:** version-aware lexer/parser (5.3–8.5); `syntax.ParseBest`
  parses at the target version and falls back to the newest grammar
  (permissive) when that yields errors (e.g. 8.4 fixtures in an 8.1 project);
  old code using later keywords as identifiers keeps working.
- **Types:** PhpStorm-like atom sets (`int`, `\Foo`, `\Foo[]`, `true`/`false`
  literals typed precisely); unknown is distinct from `mixed`; rules stay
  silent on unknown. `never` disappears from unions.
- **Variables:** union of reaching definitions — an unconditional `$x = v;`
  in a block, or an if/elseif/else chain assigning `$x` in every branch (or
  leaving), hides earlier definitions; loops add back-edge definitions; a use
  inside its own assignment sees only earlier definitions; inline
  `/** @var T $x */` overrides the next assignment; a foreach binding hides
  earlier definitions inside the loop body; element writes (`$a[k] = v`)
  widen the element type of the reads they reach (below).
- **corpus C review (2026-10-07):** builtins win over polyfills — a source
  declaration of a function the stubs know at the target version (e.g.
  symfony/polyfill-mbstring's untyped `mb_strtolower`) is ignored, as PHP
  cannot redeclare a builtin (it was hiding `string` returns); an
  unconditional assignment hides earlier definitions made in the same block
  even for uses after the block; enclosing conditions that precede the last
  reaching definition no longer narrow it (`if (null === $x) { $x = f(); }`);
  `!($x instanceof I)` also removes the subtypes of `I`; a type check on a
  set containing `mixed` keeps every checked type (`string|mixed` passing
  `is_scalar()` may be an int); doc types that are not type or class names
  (`{array}`, `self::KIND_*`) are unknown instead of classes; `array_rand()`
  with one key and `mb_convert_encoding()` are typed by their arguments.
- **Narrowing:** `is_*()`, `instanceof`, null/true/false comparisons, isset,
  truthiness — in ternary branches, if/elseif/else bodies, `&&`/`||` operands,
  `while` bodies and after early-exit guards.
- **Array shapes and emptiness (2026-10-07):** a type may carry array
  facts beside its atoms (`types/shape.go`): per-key types, a sealed flag,
  non-emptiness, and the same facts for the elements of `T[]` members. They
  never change the atom set (`array{a: int}` is still the atom `array`, a
  uniform literal still `int[]`), so `Has("array")`, `IsArrayLike`, `Atoms`,
  `Equal` and `String` behave as before; `ShapeString` shows them (tests),
  `DocString` round-trips them through the index's doc-type strings. Sources:
  literal arrays with literal/implicit keys (`[]` is the sealed empty shape),
  doc shapes `array{k: T, k2?: U, ...}`, `list{T, U}` / `array{T, U}`,
  `non-empty-array/list`, element shapes (`list<array{…}>`, `array{…}[]`),
  and destructuring (`['k' => $v] = $shape`, `[$a, $b] = …`). Unions merge
  shapes key by key (a key missing from a sealed side becomes optional); a
  member without facts (plain `array`) drops them. Capped at
  `MaxShapeKeys` = 32 keys. `$a['k']` with a literal key listed by the shape
  gives that key's type, widened by the writes to the same key or to a
  computed key that reach the read (a nested `$a['k'][…] = v` or a
  destructuring write makes it unknown); a key added by `$a['k'] = v` to a
  sealed literal is typed from its reaching writes; anything else falls back
  to the element type / unknown as before.
- **Order-aware element writes (2026-10-07):** element writes are
  definitions in the variable's reaching-definition list
  (`infer/reaching.go`, the same kill/barrier/back-edge walk as variables)
  that never kill: a write reaches the reads after it in its block, after
  branches that may have run it, and the reads earlier in an enclosing loop
  (back edge); a whole-variable assignment that kills earlier definitions
  also hides earlier element writes. As for variables, a back-edge write of
  unknown type (usually a cycle through the read) adds nothing.
- **foreach over shapes:** iterating a sealed shape (`['a' => 1, 'b' =>
  'x']`, `array{int, string}`, `[]` filled by `$a[] = v` writes) gives the
  union of its values and, for the key, `int` or `string` when every listed
  key and reaching write key is of that kind (`int|string` otherwise);
  iterating a `T[]` variable also unions the reaching element writes.
  Unsealed shapes (`array{a: int, ...}`) stay unknown. Passing the variable to a by-reference parameter
  (`sort($a)`), binding it by reference or iterating it by reference drops
  its shape. `array_values/reverse/unique` drop the shape, `array_slice/
  filter` also the non-empty fact.
- **Non-empty narrowing:** `[] !== $x`, `$x != []`, `!empty($x)`, `$x`
  truthiness, `count($x) > 0` / `>= 1` / `!== 0` / `=== n≥1` (either operand
  order, negations, `sizeof`) mark the array members non-empty, as do
  non-empty literals and doc types. The fact is dropped when a mutation of the
  variable (assignment, by-reference argument, unset, reference) lies between
  the condition/definition and the use, outside branches exclusive with the
  use, or anywhere in a loop entered after it; for `$this->prop`, any
  non-builtin call also counts. On a non-empty array `reset/end/array_pop/
  array_shift` lose their `false`/`null` member, `current` too unless the
  pointer may have moved (next/prev/end/reset/each); these functions also use
  the union of a sealed shape's values and the variable's element writes.
- **PHPDoc:** parsing is single-pass (each sub-expression once), capped at
  `types.MaxDocTypeLen` = 4096 bytes (longer: unknown), 32 nesting levels
  and 4096 parts per type including alias expansions (beyond: mixed);
  `FuzzFromDoc` checks speed and `DocString` round-trips.
  `@template` names map to `mixed` (except class templates bound by
  generic arguments, see Generics, and method templates bound by a call's
  arguments, see Method-level templates); `@phpstan-type`/`@psalm-type`
  aliases expand (including aliases used inside an alias; a self-reference
  reads as `mixed`); `@phpstan-import-type` → `mixed`; nested generics parse. Conditional
  types `(T is X ? A : B)` give `A|B`. A doc intersection refining an object
  declaration (`@return Mock&T` on `: Mock`) is kept. Native declarations
  know no scalar aliases (`Double`, `integer` are class names).
- **Flow (2026-10-07 review):** `$this->prop` is narrowed like a variable;
  a closure importing `&$v` makes `$v` unknown after it; an inline
  `/** @var T $x */` only overrides the definition in the statement it
  annotates.
- **Names (all rules):** function, method and class names, keywords and
  magic constants match case-insensitively (as PHP does); variables,
  properties and constants stay case-sensitive. A builtin function or
  constant only counts when the call/reference resolves to the global one
  (`ctx.GlobalFunctionName`, `util.GlobalConstName`) — a same-named
  namespaced or imported user symbol does not. Fixes that insert builtin
  calls/constants write `\name` when a bare name would be captured
  (`util.QualifiedBuiltin`, `util.QualifiedGlobalConst`).
- **Calls:** first-class callables (`f(...)`) are `\Closure`; named arguments
  bound in builtin return-type overrides.
- **Body return types (2026-10-07):** a call to a function or method
  declared in the current file without declared or `@return` type is typed
  from its body (`infer/returns.go`): the union of its `return` values, plus
  `null` for `return;` or a reachable end of body (`syntax.Terminates`).
  Unknown when any returned value is unknown, for generators, bodies that
  never return, on recursion (a body already being inferred) and beyond 4
  nested inferences. Methods only when the call cannot reach an override:
  private or final method, final class or enum, or a non-virtual static call
  (`self::`, `parent::`, `Name::`); trait and interface methods are skipped.
  `@phpstan-return`/`@psalm-return` alone do not count as documented.
- **Cross-file inferred returns (2026-10-07):** while building the project
  index (`runner.ExtractSymbols`, which parses every file anyway),
  `infer.AnnotateReturns` infers the body return type of each untyped
  function and non-private, non-abstract class/enum method and stores it as
  `Inferred` (separate from `Return`/`DocReturn`). The inference is
  self-contained: its index holds only that file over the stubs, so
  literals, `new`, casts, declared params/properties, `$this`, builtins and
  typed same-file calls work, anything from another file is unknown.
  Namespaced function names resolved through the global fallback are
  recorded (`ReturnDeps`); when the project declares one, the file's
  inferred returns are dropped (`DropStaleInferred`). Calls from other files
  use it under the same override rules as same-file inference (functions;
  private/final methods, final classes/enums, `self::`/`parent::`/`Name::`).
  A project function shadowing a builtin (a polyfill) is not typed from its
  body. Cost on corpus A (17.5k indexed files): 4,655 inferred returns, index
  build time within noise, +1.4 % allocations; no finding changed on the
  three reference corpora (their cross-file calls are typed).
- **Generics (2026-10-07):** a class atom may carry generic arguments
  (`types.TypeArgs`; `Collection<int, Foo>` is still the atom
  `\Collection`, so `Has`, `Classes`, `Equal`, `String` are unchanged;
  `DocString` round-trips them; unions keep them only when every member
  agrees). The index stores class templates with bounds
  (`@template[-covariant] T of X`, psalm-/phpstan- variants), the arguments
  given by `@extends`/`@implements`/`@use` (and `template-`, `phpstan-`,
  `psalm-` variants) with templates as `\~T` atoms, and per method the
  documented return mentioning class templates (`GenReturn`, preferring
  `@phpstan-return`/`@psalm-return`). A method call on `Collection<int,
  Foo>` (or a class whose `@extends` binds them, or `parent::` from such a
  class) substitutes the bindings (`T`, `T[]`, `?T`, `list<T>`,
  `Coll<TKey, U>`); method templates and unbound class templates still
  read as mixed (bounds are not used for returns). One argument given to a
  Traversable class with several templates binds the value template
  (`Collection<Foo>`, `Generator<Foo>`). foreach over `iterable<K, V>` or a
  Traversable class yields the bindings of `Traversable`'s TKey/TValue
  (through `@implements IteratorAggregate<int, Foo>`, Doctrine
  collections, `ArrayIterator<K, V>`, `Generator<K, V>`, stub classes such
  as `DOMNodeList<DOMElement>`; unbound templates fall back to their
  bound), mixed with `T[]` members (`Collection<Foo>|Foo[]`); a mixed or
  unknown value stays unknown. `instanceof` and declared types refined by a
  doc type (`@var Collection<int, Foo>` on `: Collection`) keep the
  arguments. The stubs were regenerated to carry the templates; their
  array shapes are stripped at generation (`pathinfo()` documents optional
  keys as required). Delta: +9 UnnecessaryCasting on corpus B, all true
  positives (`(int) $entity->id` on `?int` ids narrowed by a null check,
  reached through `Collection<int, Entity>` / `iterable<Entity>` foreach).
- **Per-key narrowing (2026-10-07):** `$a['k']` (literal key on a variable
  or `$this->prop`) is narrowed like a variable: `isset`, `!== null` /
  `!= null`, `!empty`, truthiness, `instanceof`, `is_*()` and early-exit
  guards (`if (!isset($a['k'])) return;`, `if (!isset($a['k'])) { $a['k']
  = default; }`). The array itself gains the key: `isset($a['k'])`,
  `array_key_exists('k', $a)`, `$a['k'] !== null`, `!empty($a['k'])` make
  a shape key required (and non-null where the condition says so), the
  array non-empty and not null/false. Both are dropped when the array or
  the element may have changed between the condition and the use (whole
  assignment, reference, unset, by-reference argument, a write to the same
  literal key or to a computed key; for `$this->prop` any call). Fixed on
  the way: `is_numeric()`/`is_scalar()`/… on a value with no matching atom
  (mixed) narrowed to the first checked atom only (`int`), now to all of
  them. No finding changed on the reference corpora.
- **Method-level templates (2026-10-07):** a function or method declaring
  `@template T` (psalm-/phpstan- variants) whose documented return type
  (preferring `@phpstan-return`/`@psalm-return`) uses T is indexed with
  `FuncTemplates`: the template names/bounds, the return type and the
  parameter types mentioning them (preferring `@phpstan-param`/
  `@psalm-param`), with method templates as `\~~T` atoms (class templates
  stay `\~T`). A call binds T from its arguments (positional or named):
  `class-string<T>` from `Foo::class`/`self::class`/`parent::class` (now
  typed `class-string<\Foo>`, a `string` atom carrying the class as
  generic argument; `static::class` stays `string`) or a `class-string<Foo>`
  value, `T`/`?T` from the argument's type (minus the other members),
  `array<T>`/`T[]`/`list<T>` from the element type (or a sealed shape's
  values), `iterable<K, V>` from the iteration types, `Box<T>` from the
  argument's own `Box<…>` arguments; several parameters binding T give the
  union. Class templates in the same return type bind from the receiver as
  for Generics. The bound type replaces the declared one when it agrees
  with it (declared unknown/mixed/`object`, every member a subtype of a
  declared member, or an intersection refining it: `@return T&Stub` on
  `: Stub`); otherwise, or when a template stays unbound or binds to
  mixed, the call is typed as before (templates read as mixed). So
  `$em->getRepository(Foo::class)->find(1)` is `?Foo`,
  `$container->get(Foo::class)` is `Foo`, `createStub(Foo::class)` is
  `Foo&Stub`. The embedded stubs were not regenerated (phpstorm-stubs
  hardly use method templates). Delta: +4 UnnecessaryCasting on corpus B,
  all true positives (`(int) $id` where `$id` is a `?int` entity id narrowed
  by a null check, the entity coming from `narrowList($rows, Foo::class)`
  documented `@param class-string<T>` / `@return list<T>`).
- **Assertion annotations (2026-10-07):** `@phpstan-assert`,
  `@psalm-assert` and their `-if-true`/`-if-false` variants are indexed per
  function/method (`index.Assertion`: kind, target = parameter, `$this` (the
  receiver) or `$this->prop`, negation `!T`, type; `=T` reads as T;
  method-call targets are skipped). A call statement narrows the argument
  bound to an asserted parameter for the following statements of the same
  block (until it is reassigned, like early-exit guards); an
  `-if-true`/`-if-false` call used as a condition narrows in the matching
  branch only (if/elseif/while bodies, ternary branches, `&&`/`||`
  operands, `!call()` guards that leave). `$this` targets narrow the
  receiver (`if ($e->isTestMethod()) { $e->… }`), `$this->prop` targets
  only calls made on `$this` (or `self::`/`static::`/`parent::` from an
  instance method). Positive assertions keep the members of the current
  type that belong to the asserted type (subclasses, `true`/`false` for
  `bool`, arrays for `iterable`), else take the asserted type; negative
  ones remove it. Templates bind as for method templates, so PHPUnit's
  `assertInstanceOf(Foo::class, $x)` (`@phpstan-assert =ExpectedType`)
  and webmozart's `Assert::isInstanceOf($x, Foo::class)` narrow to `Foo`.
  Like other guards, an assertion never makes an unknown type known.
  A method call on a union of classes uses the assertions only when every
  class resolves to the same method. Delta: −3 OffsetOperations on
  corpus B `--all` (false positives after `static::assertIsInt($pos)` /
  `assertIsString($label)`), and −1 on corpus A vendor (`Assert::notFalse()`).
- **Untyped property inference (2026-10-07):** a property without declared
  or doc type that only its own class can write — private, or protected in
  a final class; not static, no hooks, not declared in a trait nor in a
  class using traits — is typed as the union of its initial value (its
  default; else null; for a promoted constructor parameter, the
  parameter's declared/`@param` type) and every value the class's methods
  assign to it (`$this->p = v`, but also `$copy->p = v` on another
  instance, e.g. withers on a clone). The implicit null is dropped when the
  constructor assigns the property in a top-level statement. Unknown when
  an assigned value is unknown or mixed, when the property is written in an
  untracked way (by reference, destructuring, foreach target, unset,
  by-reference argument), on a dynamic property write anywhere in the class
  (`$this->$k = $v`), and for a property without default that the class
  never writes (ORMs and serializers hydrate those through reflection).
  `$this->p[] = v` / `$this->p['k'] = v` widen its arrays to plain `array`
  (keeping null; other members: unknown); `$this->p++` keeps an
  `int|float|null` type and adds `int`. Public properties are not inferred
  (any code may write them). Same-file reads compute it from the class body
  (cached per class: one walk collects every property's writes);
  `AnnotateReturns` stores it in the index as `Property.Inferred` for other
  files (dropped with the inferred returns by `DropStaleInferred`). On corpus A
  vendor + src and corpus B src 157 of 248 untyped instance properties are
  inferred (modern code types its properties); no finding changed on the
  three reference corpora.
- **Doc comments in infer** are parsed once per declaration and Env
  (`Env.DocOf`; the templates/aliases in scope at a declaration are cached
  too): `resolverFor` used to re-parse every enclosing doc comment for each
  `@param` and inline `@var` it resolved (quadratic: an 800 KB class doc
  with 5,000 inline annotations took 12 s, now 1 s). Inline `@var` scanning
  starts at the scope by binary search instead of from the file start.
- **Fixed on the way (2026-10-07):** an annotated assignment
  (`/** @var array{k: int} $v */ $v = f();`) counted as a mutation after
  the annotation, so the non-empty fact of the shape was dropped and its
  keys became optional; the assignment is now the definition point.
  `ArrAY <…>` (a builtin name followed by a space and `<`) attached generic
  arguments to the builtin atom, which did not round-trip (found by
  `FuzzFromDoc`).
- **T-rules typer** (`infer/trules.go`): shared by UnnecessaryCasting and
  CallableParameterUseCaseInTypeContext; `SpecOnly` mode follows the spec
  text literally.
- **Index:** project index built in parallel but added in path order, so
  duplicate declarations (polyfills) resolve the same way every run.
- **Stubs:** JetBrains phpstorm-stubs (Apache-2.0) compiled into a ~690 KB
  embedded index (`make stubs`).
- **Fix safety:** `make fixcheck` applies every quick-fix offered on the
  real-world corpus (corpus A vendor + src: 30,757 fixes over 8,597 files) one
  finding at a time and requires the result to parse; `FIXCHECK=php` also runs
  `php -l` on samples. Currently 0 broken fixes.
- **Robustness:** a rule that panics yields an `internal` error finding for
  that file and is skipped; other rules still run. `TestNoCrashOnCorpus`
  runs every rule over corpus A vendor + Symfony src (17,841 files), and
  `FuzzRules` (in `make fuzz`) runs every rule and fix on fuzzed sources
  (6.6M executions clean after fixing one crash; the crashing input is kept
  in `internal/rules/testdata/fuzz/` as a regression case).
  `ctx.Text`/`ctx.SpanText` return "" for empty, inverted or out-of-range
  spans.

### Coverage pass on infer, types, index, phpdoc, names, stubs (2026-10-07)

All six packages are at 100% statement coverage from their own tests and own
fixtures (EA fixtures and local corpora excluded). Most new tests sit in the
package itself (`*_cov_test.go`, `edges_test.go`). No exemptions; the
embedded stubs data is a gob file, not Go code. Bugs fixed on the way (no
finding delta on corpus A `src` or corpus B, vendor excluded):

- **Element writes reach whole-variable reads.** The writes `$a['k'] = v` /
  `$a[] = v` were applied only by the per-key readers (`$a['k']`, foreach,
  `current()`…), so the variable itself kept its pre-write type. After
  `$a = ['x' => 1]; $a['y'] = 'a'; $b = $a;`, `$b` was the sealed shape
  `{x: int}`, `$b['y']` was `int` and iterating `$b` gave `int`; returned or
  passed arrays had the same stale type. A variable read now applies the
  writes reaching it to its array members (`withElemWrites`): the element
  type gains the written values (plain `array` after a nested or unknown
  write) and the shape is dropped (`$b` is `(int|string)[]`; `$l = [];
  $l[] = new Foo;` returns `Foo[]`). Readers that apply the writes key by
  key read the base type (`baseType`), so `$a['x']` keeps its per-key
  precision. Destructuring a variable (`[$p, $q] = $a`) reads keys the same
  way. Strings and ArrayAccess objects are not affected. Only reads of
  array variables with reaching writes pay for it: `BenchmarkTypeOfElementWrites`
  (write-heavy) allocates about 30% more, `BenchmarkTypeOfVariables` is
  unchanged, and `analyse --all` on corpus A vendor takes the same time.
- **Bitwise operators on strings.** `&`, `|`, `^` (and `&=`, `|=`, `^=`)
  were always `int`, and so was `~`. Two string operands give a string in
  PHP, so these are now `string` for two strings, `int` when one operand
  cannot be a string, and unknown when both may be (unknown or mixed
  operands).
- `/=` was unknown. It is now `int|float` like `/` for numeric operands.
- A type emptied by `Without` (for example `null` minus `null` in
  `$x !== null && …` narrowing) printed as `""` instead of `?unknown`.
- With `SoundArithmetic` (T-rules typer), `$a + $b` on `T[]` arrays was
  `int|float`, and so were `$a - 1` and `$a + 1`. Now `$a + $b` is `array`
  and the other two are unknown (TypeError).

Removed as unreachable, each with its proof in the code or the tests:

- `names.Resolver.Scopes` and `types.HasTypeArgs` (unused).
- The `$this` case in doc type parsing (`scalarAliases` maps it first).
- The zero-atom fast path in `types.Of` (it gave the same unknown type).
- In index: the `x.templates` check in `genResolver`, since there the names
  in scope are exactly the class and member templates. Also the nil-doc
  guard of `assertions`, since callers pass a doc.
- In infer:
  - the unreachable tail of `unaryType` (the parser builds only the listed
    operators);
  - the nil-body checks in `scopeVars` and `collectMutations` (parser bodies
    are never nil, and abstract methods are skipped);
  - the class nil check in `genMethodReturn` (`genBindings` lists only
    indexed classes);
  - the no-argument `iterable` pattern (patterns mention a template, so they
    carry arguments);
  - the function-boundary case of `narrowExprAfter` (callers pass the
    enclosing function-like);
  - the "later write invalidates guards" branch of `guards` (`lastReset`
    already starts after the last assignment);
  - `FuncCall.Args` nil checks in the T-rules typer (the parser always
    builds the list).
- `index.directSubclasses` now reads the children's classes under one lock
  instead of re-locking per child (a removal in between gave a nil class).

Guards kept and triggered by tests: shape key caps (literal arrays, doc
shapes, merged unions), `maxVarDefs` for element writes, `maxBodyDepth` and
body recursion, a stale index whose spans no longer match the buffer (LSP
edits: the property and body inference give unknown), `TypeOf(nil)`,
malformed `define()` calls, constants with an empty initialiser, no
`Traversable` without stubs. Also the T-rules recursion guard and stub
functions lacking a parameter the overrides read. `stubs.Index` decoding is
split into `decode`/`must` so a corrupt embed panics through tested code.

## CLI and LSP

- `docs/rules-reference.md` (generated by `make rules-doc`) documents every
  rule from the specs' own summaries. The "has a quick-fix" flag shown there,
  in `custos explain` and in `custos rules --json` reflects what custos offers
  (derived from own fixtures with a changing expected output), not the
  upstream catalogue — e.g. GetDebugTypeCanBeUsed has no fix by design, while
  MagicMethodsValidity, NotOptimalRegularExpressions and
  SuspiciousBinaryOperation do.
- `custos analyse` exits 1 on findings ≥ `--fail-on` (default `warning`);
  formats text, json, checkstyle, github, sarif.
- Baselines: `--generate-baseline` / `--baseline` (or `"baseline"` in
  `custos.json`); entries keyed by path, rule, message and the flagged line's
  text, so moved code stays suppressed.
- `@noinspection <LegacyID>` (PhpStorm-compatible) and `@custos-ignore <ID>`
  both suppress; a comment before the first statement covers the file.
- PHP target: `custos.json` `php`, else composer `config.platform.php` /
  lowest `require.php`, else 8.4. `--php` and `--comparison-style` override.
- LSP: push diagnostics only (not every client supports pull); settings come via
  `initializationOptions` (same shape as `custos.json`); each diagnostic
  carries `data` = {rule, start, end} (byte span) and code actions match
  findings by the request range or by that data when the client sends the
  diagnostics back; lazy
  `codeAction/resolve` when the client supports it, inline edits otherwise;
  background project index when a semantic rule is enabled; the server
  registers a `**/*.php` file watcher (dynamic registration) and updates the
  index on `workspace/didChangeWatchedFiles` (create/change/delete), then
  re-analyses open documents; saved buffers are re-indexed on `didSave`.
- Suppress action (2026-10-07): for each finding in the requested range the
  LSP offers "Suppress <Rule> for this statement", inserting
  `// @custos-ignore <Rule>` (indented) above the innermost statement that
  starts its own line. It is offered only after re-analysing the edited
  buffer shows exactly that finding gone (a shared statement would silence
  neighbours) and never before a file's first statement (that position
  suppresses the rule file-wide); checked eagerly even with lazy resolve,
  so clients never see an action that fails. Listed after the fixes.
- `didSave` re-indexes the buffer and re-analyses every open document (a
  saved declaration can change other files' findings); before, they
  waited for their next edit.
- No on-disk index cache (planned in the migration, declined 2026-10-07):
  a cold project index takes 0.37 s for a 7.6k-source project (incl. vendor)
  and 0.48 s for a 10k-source project; a cache would still stat/hash every
  file and decode the symbols, saving ~0.3 s at best while adding
  invalidation bugs (binary version, PHP target, renamed files). The LSP
  builds it in the background, so first diagnostics are not blocked.

### Coverage pass on the CLI, LSP and I/O packages (2026-10-07)

Bringing cmd/custos, runner, report, baseline, config, safeio, lsp and the
conformance harness to full statement coverage surfaced these bugs (fixed):

- `--fail-on` accepted any value: a typo such as `--fail-on=warn` silently
  meant `never` and turned a CI gate off. Unknown values (and unknown
  `--format` values, previously reported only after the whole analysis) are
  now usage errors (exit 2) checked before analysing.
- `custos fix` aborted at the first unreadable file (after having rewritten
  the files sorted before it) and on the first write failure. It now reports
  each failure, fixes the other files and exits 2 at the end.
- `custos fix --dry-run` counted and diffed symlinked files that a real run
  refuses to write. Non-regular files are now skipped (`not fixed`) before
  being read, in both modes.
- `fix --cpuprofile` was accepted but ignored; it now profiles. A failing
  `pprof.StartCPUProfile` leaked the created profile file.
- `custos rules -h` / `lsp -h` printed "flag: help requested" and exited 2;
  `-h` now exits 0 everywhere. Flag errors go to the command's stderr writer.
- Output errors were ignored by the `github` format (and per line by
  `text`): a closed stdout returned success. They are now returned (exit 2).
- LSP: when the client sent `initializationOptions` (some always do),
  the merged configuration dropped `custos.json`'s `paths`, `exclude` and
  `baseline`, so the project index covered the whole root including
  excluded directories. What to analyse now always stays the project's.
- LSP: a request whose handler panicked (e.g. a crashing quick-fix while
  building code actions) was answered with a `null` success result; it now
  gets an internal error (-32603). The `didChangeWatchedFiles` goroutine is
  guarded the same way (a parser crash there killed the server).
- LSP: file changes arriving between `initialized` and the start of the
  background index build were dropped (the "indexing" flag was set inside
  the goroutine); it is now set before the goroutine starts, so they are
  queued and replayed.
- LSP framing: header and body are written in one `Write` (a failure can no
  longer leave a header without its body).

Removed as unreachable: JSON marshalling error branches for values built
from known types (baseline file, LSP messages/results/params, now
`mustJSON`). Test seams added: `safeio.beforeOpen` (swap a path between
the Stat and the Open, as a concurrent writer could) and, in cmd/custos,
`registry`/`catalogue` (the embedded rule set and catalogue cannot fail
otherwise). The only uncovered statement left is `main()` itself, a
one-line `os.Exit(run(...))` wrapper. Tests that rely on permission bits
(unreadable working directory, chmod 000 files) skip under root.

### Coverage pass on syntax, analysis, util, fix, diff (2026-10-07)

syntax, analysis, analysis/util, fix, diff, phpver and meta are at 100%
statement coverage from their own tests and own fixtures (EA fixtures and
local corpora excluded). Lexer edge cases are checked token by token against
PHP 8.5's `token_get_all` (`TestLexSnippetsMatchPHP`). Grammar cases are
checked against `php -l` of the matching version (`/opt/homebrew/opt/php@X.Y`,
`TestParseCases`). `TestNodeKinds` parses `testdata/cov/allkinds.php`, which
uses every construct, and requires each node kind to appear and to name its
own Go type, so the generated `kinds.go` is covered without an exemption.
Findings (`--all`, JSON) are byte-identical old/new on corpus A src and vendor
and on corpus B (vendor excluded). Bugs fixed:

- Lexer, `readonly (`: it was lexed as a name in every version when `(`
  followed (comments included). Only PHP 8.1 does that, and only across
  whitespace. 8.2+ always emits the keyword, so `public readonly
  (A&B)|null $x` (DNF type) was a false syntax error. The parser now
  accepts `readonly(...)` as a function call.
- Lexer, `private( set )`: inner whitespace was accepted, but PHP lexes only
  `(set)` (any case) as one token. With spaces, PHP reports a syntax error,
  and now custos does too.
- Lexer: `1..2` gave `1` `.` `.2` instead of PHP's `1.` `.2`.
- Lexer: a comment between `->`/`?->` and the member name dropped the
  member-name context. `$o->/*c*/list` lexed `list` as a keyword, and `#[`
  right after `->` started an attribute instead of a comment.
- Parser: an argument-less `new` took postfix operators. `new A->b`,
  `new A::C`, `new A[0]` and `new $a::C` were accepted, though PHP rejects
  them (only `new A()->b` and `new class {}->b` are valid). The fix exposed
  that `new A::$cls` / `new static::$cls` used to parse as
  `(new A)::$cls`. `New.Class` is now the static property fetch.
- Parser: `$x instanceof 1` / `instanceof [1]` was accepted without an
  error, and a reserved word before `::` (`Foreach::x()`) was taken
  silently as a class name. Both are now reported, with the same recovery.
- Parser: `use T { foo as; }` (no visibility or alias) was accepted.
- `Terminates`: `break N` used level 2 for every literal other than `1`.
  So `while (true) { foreach (…) { foreach (…) { break 3; } } }` counted as
  never exiting, and `break 01` / `break 0` were misread. Levels are now
  parsed (non-literal: assumed to leave). A `continue` aimed at a switch
  acts as `break` in PHP, and it now makes the switch fall through.
- diff: an empty hunk range was printed as `-N+1,0` instead of `-N,0`. With
  `patch`, that put insertions at the wrong line (context 0) or warned
  (an empty old file).
- util: casts that differ only in case (`(INT)` vs `(int)`) were not
  equivalent.

Removed as unreachable (proof in a comment at each site): the lexer
dispatcher's fallback (an unknown state now returns without consuming, and
the run loop's never-stall guard is tested by forcing such a state), the
closure parser's zero-width error branch, the progress guards of the class
body and string-interpolation loops, end widening in `SetParents` (a node
ends at the last consumed token, after its children), the negative
checkpoint index in `LineIndex`, the Myers loop's unreachable `return nil`,
and the unused `TokenKind.IsMagicConst`. The util removals are the unused
`UnstableVariable` plus nil and cap re-checks already enforced by the
callers.

## Security (untrusted input)

custos analyses code it does not trust (pull requests in CI, files opened in
an editor), so hostile input must not crash or hang it.

- **Nesting limit:** parser recursion and AST depth are capped at
  `syntax.MaxDepth` = 4000 (≈1,300–4,000 source levels depending on the
  construct; real code stays below ~100). Deeper files are reported once
  ("nesting deeper than 4000 levels; file not analysed") and skipped, instead
  of overflowing the Go stack (a fatal, unrecoverable error) or going
  quadratic in recursive consumers. Found by probing: 200k nested `[`,
  300k `!`, 300k-part `.` chains crashed; 100k nested `if` hung.
- **Doc-type parsing:** a security review (2026-10-07) found
  `types.FromDoc` exponential on nested generics (`array<array<…>>`, ×2 per
  level, ~30 levels hangs) — a single doc comment could hang CI or the LSP
  server. Fixed with linear parsing plus depth/size caps and a fuzz target.
- **Expansion caps (2026-10-07):** a follow-up review flagged a resource
  cap bypass in `types/parse.go` and `infer/generics.go`; both confirmed by
  probes. (1) Alias and template expansions (`"=" + definition` from a
  resolver) were parsed without the `MaxDocTypeLen` check applied to the
  top-level text: a 1 MB `@phpstan-type` definition with few parts, used
  1,000 times in a 2 KB doc type, was rescanned at each use (13.6 s for one
  type). Now an expansion longer than `MaxDocTypeLen` reads as mixed,
  `phpdoc.TypeAliases` records such definitions as "" (mixed) so resolvers
  never copy them, and each `FromDoc` call also has a total byte budget
  (`maxDocBytes` = 64 × 4096 scanned, nested members and expansions
  included; beyond it parts read as mixed). (2) Each `@extends` level
  substituted the bindings into the next class's arguments with a fresh
  parse budget, so arguments using a template several times
  (`@extends B<array{a: T, b: T, c: T, d: list<T>}>`) grew geometrically
  along a chain: 1,000 such classes took 17.7 s for one method call, again
  for each distinct receiver. Now a binding whose text exceeds
  `MaxDocTypeLen` is not used (the template reads as its bound or mixed)
  and `genBindings` visits at most `maxGenAncestors` = 64 classes per
  receiver (17.7 s → 27 ms; 20 distinct receivers on the 1,000-class
  chain in 0.7 s including parsing). Method-template bindings (above) use
  the same length cap. Regression tests: `TestFromDocPathological` (huge
  and long aliases), `TestGenericChainBounded`, `TestScopeDocsParsedOnce`,
  `TestTypeAliasCap`; alias-heavy seeds added to `FuzzFromDoc`.
- **Whole-tool audit (2026-10-07):** probes with generated files (one
  construct repeated N and 4N times: 20k statements, 20k-case chains,
  try/catch, closures, goto labels, …), pathological string literals
  (300 KB of `[`, `{`, `\u{`, `%`, `(`, invalid UTF-8, … passed to the
  string-parsing rules), 15–50 MB files and malformed LSP / config input.
  Findings were unchanged on the reference corpora (corpus A, corpus B,
  EasyAdminBundle). Limits:
  - `syntax.MaxFileSize` = 10 MB: larger sources get one error ("file
    larger than 10 MB; not analysed") and no tree. Memory is linear but
    peaks at up to ~250× the file size on garbage input (a 50 MB unterminated
    string took 11 GB, 20 MB of `;` 13.7 GB and 5 GB of output); a 15 MB
    garbage file now peaks at ~4 GB, so 10 MB keeps one file near 2.5 GB.
    Files are read through `safeio` (`runner.ReadSource`): at most
    MaxFileSize+1 bytes, regular files only, so a `.php` symlink to
    `/dev/zero` or a FIFO fails instead of hanging; `custos fix` never writes
    through a symlink.
  - `syntax.MaxErrors` = 1000 syntax errors per file (then one "more than"
    error; messages beyond it are not even formatted) and
    `analysis.MaxFindingsPerFile` = 10,000 findings (then one "internal"
    warning).
  - LSP: a Content-Length above 128 MB is answered with an error and its
    body skipped (a huge value crashed the server: `makeslice` panic outside
    any recover); header lines are capped at 64 KB, negative lengths and
    framing errors end the server cleanly; a negative fix index in
    `codeAction/resolve` no longer panics.
  - Config: custos.json / composer.json are read with an 8 MB cap and the
    baseline with 256 MB (regular files only); `paths` and `baseline` must
    stay inside the project root, so a repository cannot point custos (or
    `custos fix`) outside it (`"baseline": "../../../../dev/zero"` hung).
    Go's JSON decoder already rejects nesting beyond 10,000 levels; the
    hand-written composer.json parser of SecurityAdvisories caps depth at 512.

  Quadratic algorithms fixed (input, before → after at N = 20k unless
  noted): per-finding suppression lookup re-walking the tree from the top
  (20k findings: 3.4 s, a 40k-statement file did not finish in 120 s;
  now one indexed pass) and tag parsing re-scanning a long comment line per
  tag; `util.Reachable` / `StmtList` scanning statement lists (cached
  `syntax.FirstTerminating`, binary-searched `syntax.StmtIndex`); line /
  rune / UTF-16 columns counted from the line start for every finding
  (minified one-line files; checkpoints every 1 KB on lines > 4 KB);
  value discovery (`PossibleValues*`, `DiscoverValues`,
  `ReachingAssignments`) re-walking the function body per variable use
  (printf on one variable 20k times: 40 s → 0.2 s; per-scope indexes
  cached with `syntax.File.Memo`; more than 512 candidate assignments make
  the result unknown); `VarAccesses` per name in OnlyWritesOnParameter
  (13 s → 0.5 s) and BadExceptionsProcessing (28 s → 0.6 s); per-label
  walks in UnusedGotoLabel (13 s → 0.6 s); chain recomputation in
  CascadeStringReplacement (N = 5k: 12 s → 0.7 s; chains over 256 calls get
  no fix); per-closure body walks in StaticClosureCanBeUsed (48 s → 1 s);
  per-parameter walks in SuspiciousAssignments (12 s → 0.1 s);
  AlterInForeach sibling scans; UnqualifiedReference rescanning top-level
  statements per name. String parsers: `\u{` escapes in
  `util.StringLiteralValue` scanned to the next `}` each (300k unterminated
  escapes hung); NotOptimalRegularExpressions D15 rescanned for `]`/`}` from
  every `[` (now O(1) lookups, checked against the old code by
  `FuzzNorePattern`) and D18 folds groups one regex pass at a time (now
  skipped beyond 64 KB or 256 folds); `phpdoc.Parse` concatenated
  continuation lines (quadratic on long tags) and `SplitType` rescanned
  blank runs. `syntax.TreeDepth` no longer stacks every pending sibling
  (~1 GB of allocations on a 3 MB file). Custom backtracking: none — every
  regexp in custos is a constant compiled with Go's linear-time engine; no
  user-provided pattern is ever compiled.

  Fuzz targets added: `FuzzStringLiterals` (all rules, fuzzed text passed as
  regex/format/date/callable/JSON/doc/suppression literals),
  `FuzzStringLiteralValue`, `FuzzNorePattern`, `FuzzParse` (phpdoc),
  `FuzzParseSuppressionComment`, `FuzzParseJSON`. Regression tests:
  `TestSuppressionsScale`, `TestFindingsCap`, `TestLineIndexLongLines`,
  `TestParseLimits`, `TestReadSourceBounded`, `TestHostile*` (LSP frames and
  requests, config, repository), `*Pathological`, plus equivalence tests of
  the new indexes against the previous per-query code
  (`TestVarAccessesByNameMatches`, `TestReachingAssignmentsIndexMatches`).
  The engine paths this audit reported are fixed below (Engine scaling).
- **Engine scaling (2026-10-07):** fixes for the super-linear engine paths
  found by the whole-tool audit; findings identical on the reference
  corpora (corpus A src/vendor/symfony, corpus B, default and `--all`),
  corpus A vendor timing unchanged (~0.7 s).
  - `index.Ancestors` (every class/member lookup, `IsSubtype`) is cached
    per class and PHP version; each index layer counts its changes and an
    entry is valid while the sum over the layer and its bases is unchanged
    (Add/Remove on the project index or a base invalidate it). At most
    `index.MaxAncestors` = 256 classes are returned (`ParentChain` too).
    Lookups on a 5,000-class chain / cycle: 6.9 s / 13.5 s → 20,000-class
    chain / cycle: 2.0 s each (`TestAncestorsBounded`,
    `TestAncestorsCacheInvalidation`).
  - `guards` indexes each statement list once (`guardIndex`): the keys
    occurring in its expression statements and else-less ifs, the
    assignments resetting each key; a read only visits the statements after
    the last assignment of its key that mention it, and more than
    `maxGuardScan` = 512 of them make the read unknown. The definitions of
    one variable in one scope are capped at `maxVarDefs` = 512 (beyond:
    unknown; element writes: one unknown write), the definition sort is no
    longer an insertion sort, and the T-rules typer collects a scope's
    assignments once instead of walking the scope per read (same cap).
    At N = 20k (`TestLongBodiesBounded`, before → after): `$a['k'] = …`
    121 s → <0.1 s, `if (!isset($a['k'])) return;` 86 s → <0.1 s,
    `$o = new X` 15.6 s, `array_push($a, $x)` 4.1 s, `$s = str_replace(…,
    $s)` 3.7 s → all under 0.5 s together; whole-file analysis of the
    probes: 22 s / 8.7 s / 5.3 s / 89 s → 0.24 / 0.14 / 0.18 / 1.2 s.
  - Untyped property inference finds declarations through a per-class map
    built with the writes (a class with 20k properties: 1.07 s → 0.12 s
    whole-file).
  - Rule-side follow-ups: `util.chainWalk` (`MethodInChain`,
    `PropertyInChain`) and LongInheritanceChain stop after
    `index.MaxAncestors` classes (20k-class chain: 68 s → 4.3 s
    whole-file, now linear); CompactArguments indexes the first `$name`
    token of each scope once instead of rescanning per call (20k
    `compact()` calls: 9 s → 0.2 s).
- **Helper consolidation (2026-10-07, Phase 8):** duplicated helpers merged,
  behaviour unchanged — findings (default and `--all`, JSON incl. messages,
  ranges, fixable) and `fix --all --dry-run --diff` output byte-identical
  before/after on corpus A src/vendor, corpus B, corpus C and corpus D/vendor;
  corpus A vendor timing unchanged. Merged:
  - *Value discovery:* `PossibleValues`, `DiscoverValues`,
    `PossibleValuesComplete` and DynamicCallsToScopeIntrospection's private
    variant (now `util.PossibleValuesReaching`) share one traversal
    (`valueWalk`: parentheses, ternaries, `??`, visited set, unknown flag)
    and the local-variable, file-constant, class-by-short-name and
    class-constant lookups; each variant only supplies its resolver.
  - *AST navigation* moved to `syntax` so `infer` and `util` share it:
    `UnwrapParens`, `IsFuncLike`, `EnclosingFuncLike`, `EnclosingClass`,
    `FuncLikeParams`, `FuncLikeBody`, `IsNullConst` (replacing
    `util.UnwrapParens`, `util.EnclosingFuncLike`, `infer.EnclosingClass`,
    `infer.scopeOf`/`unparen`/`isNullConst`, the T-rules' `scopeParams` and
    ~15 per-rule copies); `names.Resolver.DeclFQN`/`ParentFQN` replace
    `util.ClassDeclFQN`/`ParentFQN` and back `Env.ClassFQN`/`ClassRef`.
  - *Test context:* `analysis.IsTestPath` (behind `ctx.IsTestFile`) replaces
    `util.IsTestPath`; `util.InTestContext` (test file, or the nearest enclosing
    class-like is named and has a test-like FQN) replaces five per-rule
    copies.
  - *Per-rule copies of util helpers:* string/number literal checks
    (`util.IsStringLiteral`, `IsNumberLiteral`, `QuotedStringValue`,
    `QuotedStringContent`), `BoolConst`, `LastNamePart` (was also
    `LastSegment`), `ParentFuncCall`, `CallLastName`, `ArgValues`,
    `IndentBefore`, `IsSpace`, `AsExpr`, `IsStaticPropName`,
    `MentionsVariable`, `IsLogicalOperand`, `ArgBindsByRef` (call-time `&`
    as an option), `DocHasAnnotation`, `GlobalConstName`,
    `ResolvesToGlobalFunction`, `Env.ClassRef`, `QualifiedBuiltin`.
  Kept separate on purpose (comment at each site): the T-rules typer
  (spec-defined partial typing, layered over `Env`; its `ParamTypes`
  differs from `Env.paramType`); `infer.terminates` (shallower than
  `syntax.Terminates`, narrowing depends on it); `infer.plainString`
  (narrower than `util.StringLiteralValue` for shape keys); ForeachInvariants'
  limit discovery (init assignment wins, by-ref chains); the test-class
  checks of CryptographicallySecureAlgorithms and RealpathInStreamContext
  (skip anonymous classes, different FQN source); IsNullFunctionUsage's
  ASCII-only `true`/`false` match; InArrayMissUse's precedence check;
  `IsFuncNamed`/`IsFuncNamedFold` (one implementation, two entry points);
  the class-reference resolvers that differ (`puResolveClassName`,
  `staticCallClass`, `semClassesOf`).

## Clean-room incidents

- The conformance runner's "missing:" lines printed upstream's expected
  message text for EA cases (reported by an implementer agent during the
  class A fixes; nothing copied). EA mismatches now print only severity and
  range.
- Before `HideExpected` existed, a failing fix comparison printed EA's
  expected `.fixed` output. Two implementer agents (ReturnTypeCanBeDeclared,
  CascadeStringReplacement) reported seeing part of it; both state nothing was
  copied, and their fixtures were written from the specs.
- `make cleanroom` flagged PHPUnit API names stored as `Class.method` keys in
  ClassMockingCorrectness (mirroring upstream's key format); restructured as
  class/method pairs.
- Coverage audit (2026-10-07): `make cleanroom` flagged
  a one-line `switch` whose case repeats its subject in a new own fixture
  (ForeachInvariants `shapes.fixed.php`, written from the spec; a generic
  coincidence). The fixture was reworded.
