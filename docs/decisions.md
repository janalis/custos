# Decisions log

Decisions taken during the port. "user decision" marks choices made by the
project owner; everything else is an engineering default that can be revisited.
Private projects used as test corpora are named "corpus A" to "corpus E".
Last updated: 2026-10-08.

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
| DisallowWritingIntoStaticProperties | `disallow-write-into-static-property-default.php` | closures and arrow functions defined in a method of the declaring class are checked like the method (they have its class scope) |
| DuplicateArrayKeys | `duplicate-array-keys.php` | keys are compared as PHP stores them (escapes decoded, '7' == 7, any integer base); duplicate integer keys are reported too |
| ForeachInvariants | `foreach-invariants.php` | upstream's IDE formatter re-spaces an untouched inner for header (custos copies the body verbatim); additionally a limit variable reused by several loops is resolved from the assignment reaching each loop (one extra report); loops whose counter or header-assigned limit is mentioned after the loop are not reported (two fewer reports) |
| GetDebugTypeCanBeUsed | `get_debug_type.php` | reported without a fix: get_debug_type() names differ from gettype() for scalars (int/integer, float/double, null/NULL) |
| InstanceofCanBeUsed | `instanceof-can-be-used.php` | only exact equivalents get a fix (is_a, get_class on a final class, class_implements with an interface literal on an object); get_parent_class/is_subclass_of/class_parents and non-final get_class are reported without a fix |
| InvertedIfElseConstructs | `if-inverted-condition-else-normalization.php` | for a bool operand the fix emits x() instead of false !== x() |
| IsEmptyFunctionUsage | `empty-function.php` | no null-comparison suggestion for int/float/bool\|null subjects (empty() is also true for 0, 0.0, false) nor for possibly unassigned variables; such subjects get the generic report |
| MkdirRaceCondition | `mkdir-race-conditions.php` | the or-form re-check emits || is_dir($concurrentDirectory) (upstream negates it, inverting the logic) |
| NotOptimalIfConditions | `if-instanceof-flaws-false-positives.php` | under && the broader (redundant) instanceof is reported, not the more specific one |
| NotOptimalIfConditions | `if-optimal-conditions.php` | method/static calls are not reordered (they may have side effects); isset($x[...]) && $x is not reported (isset also guards $x itself) |
| PhpUnitDeprecations | `deprecations.phpunit91.php` | fix emits PHPUnit's real method names (assertFileDoesNotExist / assertDirectoryDoesNotExist) instead of upstream's misspelled ones |
| PhpUnitTests | `assert-resource-exists.php` | below PHPUnit 9.1 the fix suggests assertFileNotExists/assertDirectoryNotExists (the DoesNotExist names do not exist yet) |
| ReferencingObjects | `referencing-objects.php` | upstream's IDE formatter re-spaces an unreported `array& $arr` parameter; custos edits only the reported ranges |
| ReturnTypeCanBeDeclared | `return-type-hints.php` | no suggestion when the declaration would be a compile error (": void" with return $x where $x is null/void-documented, "?T" with a bare return) |
| SenselessProxyMethod | `senseless-proxy-signature.php` | an override that calls the parent without return, where the parent returns a value, is not reported (removing it would change return values) |
| SlowArrayOperationsInLoop | `slow-array-operations.for-termination.php` | the generated limit variable avoids names already used in the scope ($iMax1 instead of overwriting $iMax) |
| StaticClosureCanBeUsed | `static-closure-use.php`, `static-closure-use.php74.php` | keyed array values are not reported at file level either (the including code may bind them) |
| SubStrUsedAsArrayAccess | `substr-used-as-index-access.php` | no fix below PHP 7.0 (?? does not parse); ?? '' guard from 7.0; negative offsets other than -1 skipped (strlen($s) - n can go negative) |
| SubStrUsedAsStrPos | `substr-used-as-strpos.php` | 4-argument mb_substr fix emits mb_strpos($h, $n, 0, $enc); case-folded comparisons only against literals already in folded case; loose ==/!= comparisons get a fix only against non-numeric string literals |
| TraitsPropertiesConflicts | `traits-properties-conflicts.php` | an own property incompatible with the trait's (different default, visibility, static, readonly or type) is reported as an error: PHP refuses to compose such a class |
| UnnecessaryCasting | `unnecessary-casting.php`, `unnecessary-casting.php8.php` | an untyped private property without default (not set by the constructor) holds null: casting it is not redundant |
| UnnecessaryAssertion | `unnecessary-assertion.php` | assertInternalType is reported only when the declared return type always satisfies the named type; unknown/contradicting type names are not reported (that call fails, it is not redundant) |
| UsingInclusionOnceReturnValue | `using-inclusion-once-return.php` | success tests (conditions, logical operands, comparison with false) are not reported; no quick-fix (a plain include re-runs the file and redeclares its symbols) |
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

**WordPress / Laravel / Drupal / Nextcloud review (2026-10-08).** First run
on large code custos had never seen, downloaded as untrusted data and only
read: WordPress (1,900 files, PHP 7.4), laravel/framework from a fresh
`laravel/laravel` project (1,762 files, 8.3), Drupal core (10,425 files,
8.5) and Nextcloud server (5,879 files, 8.3), default and `--all`.
Robustness: no internal finding and no syntax error (every file also passes
`php -l` of the target version); `analyse --all` takes 0.5 s (WordPress,
Laravel) to 1.5 s (Drupal); the largest files (1–1.7 MB arrays) scale
linearly. About 1,500 findings were sampled over 150 rules (all findings
below 20, 15–20 otherwise, WordPress weighted). `custos fix --all` on
copies, then `php -l` on every changed file: 569 of 9,138 changed files
were broken, all but three by fixes of *different* rules combining (below);
after the fixes every changed file lints (Nextcloud's `build/stubs`
redeclare builtins and fail before fixing too).

| Rule / area | Was | Now |
|---|---|---|
| Fix engine (`internal/fix`) | An insertion at the start or end of another fix's replacement was not an overlap: `\` (UnqualifiedReference) landed before a rewritten call (`\$x === null`, `\$f($a)`, `\static::$i`), sometimes eating a `;`. 566 broken files. | An insertion touching another edit conflicts; the next iteration re-analyses. `make fixcheck` also applies all fixes of each file together. |
| Parser | `$b = &f() && $c` parsed as `$b = &(f() && $c)` (PHP: `($b = &f()) && $c`); a parenthesising fix then broke the code. | The by-reference value is a single operand. |
| `custos fix` | Files fixed one after another (WordPress 7.1 s). | One worker per CPU, output in input order (1.6 s). `--stats` labels a `--php` version "flag". |
| MagicMethodsValidity | `_set($n)` renamed to `__set($n)` (fatal: must take 2 arguments; callers break); `: never` reported as "got 'never'". | Reported without a fix; `never` accepted (listed divergence). |
| ReturnTypeCanBeDeclared | Unknown returned values dropped when one known type remained (`if ($x) return $x; return '';` → `: string`); PHP 4 constructors below 8.0 and always-throwing `__serialize()` got `: void` (fatal); generators got `?\Generator`. | Unknown values stop the rule unless an `@return` tag exists; those methods skipped; generators get `\Generator`. Drupal 14521 → 14367 (same tree, only this rule changed). |
| StaticInvocationViaThis | Fix `self::m()` dropped late static binding; `$this->createStub()`/`assertTrue()` resolved to a trait's `abstract static` re-declaration (PHPUnit exception lost). | `static::` unless private/final method or final class/enum (listed divergence); an abstract trait method defers to the inherited implementation, unknown when an ancestor does not resolve. corpus B 497 → 19, Nextcloud 672 → 656. |
| TypeUnsafeComparison | `$key != 'streams'` → `!==` although an int key 0 equals `'streams'` below PHP 8.0 (also bools, Stringable objects). | Fix only for operands known to be string/null/array (numbers from 8.0); otherwise a warning without fix (listed divergence). Fixable WordPress 485 → 217. |
| ClassConstantCanBeUsed | `'\A\B'` → `\A\B::class` (string loses its leading `\`). | Reported without a fix (listed divergences). |
| SuspiciousAssignments | "else may be missing" when a later statement of the `if` read the value or the block always exited; "overwritten" when the second right side could read a global/property/by-ref target. | All later statements scanned; exiting blocks and shared targets followed by a call skipped. Nextcloud 45 → 8. |
| MissingIssetImplementation | Interfaces, abstract classes, `#[\AllowDynamicProperties]` and `object\|X` unions reported as "always false" (error). | Only concrete resolved classes without `__isset`/dynamic properties. 101 → 33. |
| UnsupportedStringOffsetOperations | `$p = ['#markup' => 'x']; $p['#attached']['lib'][] = …` reported as fatal. | Nested targets trusted for property/parameter roots only. 19 → 4. |
| MockingMethodsCorrectness, PhpUnitDeprecations, PassingByReferenceCorrectness, ClassConstantUsageCorrectness, ProperNullCoalescingOperatorUsage, CallableParameterUseCaseInTypeContext | Methods of unresolvable parents reported missing; a comparator's own `assertEquals($a, $b, $delta)`; `array_multisort(array_values($a), …)`; `Bag::class` next to `use …\Bag as Alias`; unresolved class on the right of `??`; `func_get_arg()` (`mixed\|false`) taken as bool. | Each skipped (unresolvable hierarchy, foreign receiver, prefer-ref builtin, only imports resolving the written name, D3, `mixed` = unknown). |
| OnlyWritesOnParameter | `global $x; $x = …` reported (95); `++` through a by-ref alias; locals holding objects. | `global` counts as a read; by-ref `++` is a use; object locals skipped. WordPress 88 → 10. |
| ForgottenDebugOutput | `wp_die()` (permission guards) reported as debug output at error severity (563); `dump()` inside methods named `dump`. | `wp_die` dropped from the defaults; methods named like an entry are wrappers. WordPress 597 → 34. |
| AlterInForeach, IssetArgumentExistence, SlowArrayOperationsInLoop | File-scope `unset($v)` after a loop (keeps included files clean); `extract()`/`$$k`/earlier include before `isset($x)` (error); per-iteration targets and read-back accumulators told to merge once (error). | Skipped. 44 → 27 (WordPress), 15 → 3, 217 → 186. |
| ForeachInvariants, MultiAssignmentUsage, SenselessProxyMethod, ReferencingObjects | Loops mutating the iterated array rewritten to `foreach` (snapshot); destructuring moved into the header past `$m[2] ??= null`; `__construct` proxy removed next to a PHP 4 constructor (infinite recursion below 8.0); `&` dropped from overriding/interface methods (fatal signature mismatch) and unions with arrays. | Not reported in those cases. ReferencingObjects Drupal 65 → 22. |
| DisconnectedForeachInstruction, DisallowWritingIntoStaticProperties (disabled) | `$bar->advance();` treated as loop-independent; `static::$x = …` in a closure inside a method of the declaring class. | A discarded method call modifies its receiver; closures use the enclosing method's class (listed divergence). Drupal 126 → 19. |
| UsingInclusionOnceReturnValue | `if (!include_once $f)` reported; the fix to plain `include` re-ran the file (redeclaration fatal). | Success tests not reported; no fix (listed divergence). |
| IsEmptyFunctionUsage | `!empty($n)` with `?int` → `$n !== null` (differs for 0); `$z === null` for a maybe-unassigned `$z` (warning). | Null form only for `resource\|null` and certainly assigned variables (listed divergence). |
| NotOptimalRegularExpressions, SubStrShortHandUsage, ArrayIsListCanBeUsed | `"/a\$/D"` told `/D` is pointless; `return preg_match(…)` rewritten to a bool; `substr($s, 3, strlen($s) - 6)` → `-3` (differs for short strings); `array_keys($a) === range(0, count($a) - 1)` → `array_is_list($a)` (true for `[]`). | Decoded patterns; rewrites only in boolean contexts; d ≥ −2 only; `$a !== [] && array_is_list($a)`. |
| PregQuoteUsage, UnserializeExploits, EncryptionInitializationVectorRandomness, CryptographicallySecureRandomness, SecurityAdvisories | Delimiters preg_quote() escapes itself (`|`, `#`, `{`) reported (error); `unserialize(serialize($o))`; IV assignments after the call; `if ($strong)`; drupal/core-dev metapackages. | Skipped. PregQuoteUsage 40 → 20, SecurityAdvisories Drupal 12 → 0. |
| UnqualifiedReference, EmptyClass, ComparisonOperandsOrder, InvertedIfElseConstructs, OpAssignShortSyntax, NullPointerException | Callback strings in files without a namespace (94 on WordPress); `#[Attribute]` marker classes; `false !== $r = f()` swapped to `$r = f() !== false`; empty `else {}` swapped; `$m[$i++] = $m[$i++] + 1`; `$b ??= $x` not seen as a non-null write. | Fixed (namespace required, attributes skipped, loose operands parenthesised, empty else and side-effect targets skipped, `??=` handled). |

Declined: SuspiciousSemicolon on deliberate empty loops (`for (…; …; $i++);`,
`while (pcntl_waitpid(…) != -1);`) — indistinguishable from the bug the rule
targets; DegradedSwitch, TypeUnsafeArraySearch, UnSafeIsSetOverArray,
AutoloadingIssues (WordPress `class-*.php`), PropertyCanBeStatic,
BadExceptionsProcessing, LongInheritanceChain, EfferentObjectCoupling,
MultipleReturnStatements, ParameterDefaultValueIsNotNull,
ClassConstantCanBeUsed on Composer `autoload_*.php` maps — noise by design;
NullPointerException reports every dereference after the first (spec, and
disabled by default); IncrementDecrementOperationEquivalent `$n -= 1` →
`--$n` differs for null (types are rarely known; upstream behaviour);
InArrayMissUse `in_array($x, array_keys($a), true)` → `array_key_exists`
(numeric-string keys; never seen). Engine-level causes found by the review
(target-version stub return types — `substr()` is `string|false` below 8.0;
assignments used as conditions not narrowed; a call to an unresolvable
function keeps the variable's old type; negation of `@phpstan-assert-if-true`
in the false branch; `$_SERVER['SERVER_PORT']` typed `int`; absent literal
keys typed by the literal's element type; `array_reduce()` with a non-null
initial value typed nullable; class attributes not in the index;
duplicate function declarations; int overflow to float) are left to the
type-engine work.

**phpMyAdmin / Matomo / PrestaShop / Composer / PHPUnit / Doctrine ORM
review (2026-10-08).** Second run on code custos had never seen, shallow
clones plus `composer install --no-scripts --no-plugins` (vendor indexed,
not analysed), only read: phpMyAdmin (1,236 files, PHP 8.2), Matomo (3,769,
8.1), PrestaShop (7,898, 8.1), composer/composer (635, 7.2),
sebastianbergmann/phpunit (3,089, 8.4) and doctrine/orm (1,527, 8.1),
`analyse --all`. Robustness: no internal finding; the only syntax errors
are PHPUnit's own invalid fixtures (`$b = ;`, partial function application
`f(?, '!')`), confirmed by `php -l` of the target version, which reports
nothing else but compile-time errors custos does not model (closures in
constant expressions, PHP 8.4 property hooks and the 8.5 pipe operator in
PHPUnit/ORM test files that target newer versions, `true` as a class
name). `analyse --all` takes 0.4–3.3 s (PrestaShop, 7,898 files, index of
its vendor included); the largest files (Matomo's 825 KB ISO region table,
PrestaShop's 315 KB `Product.php`) analyse in under 0.1 s once the index is
built, linearly. About 700 findings were sampled over 50 rules (all
findings below 20, 12–20 otherwise; error-severity and semantic rules
first), and sampled hunks of the fix diffs of all 90 fixable rules were
read. Fix safety: `custos fix --all` on copies (77,000 edits in 7,187
files), then `php -l` of the target version on every changed file: one
broken file (NestedAssignmentsUsage, below); every individual fix was also
applied alone and re-parsed (`TestFixesKeepCodeParsable` on each project,
`php -l` samples); per-rule diffs were read for semantic changes (seven
fixes changed behaviour, below). After the fixes all changed files lint
(two ORM test models use PHP 8.4 property hooks and fail before fixing).

| Rule / area | Was | Now |
|---|---|---|
| NestedAssignmentsUsage | `$target = $this->files[] = $path` split into `$this->files[] = $path; $target = $this->files[];` (fatal: `[]` for reading); `$m[$i++]`, `$m[g()]` targets read back with side effects. | Fix only when the innermost target is re-readable (plain variable, identifier properties, literal/constant/variable keys); otherwise report only. |
| NullCoalescingOperatorCanBeUsed | `if (isset($m[$k])) { $x = &$m[$k]; } else { $x = null; }` → `$x = $m[$k] ?? null;` (reference lost, Matomo menus no longer edited); if/return in `function &get()`. | Not rewritten. 576 → 572. |
| TypeUnsafeArraySearch | `in_array(Configuration::get('PS_ROUND_TYPE'), [Order::ROUND_ITEM, …])` given `, true` (a config string never equals an int constant). | Reported without a fix: the equivalent cases (same type both sides) are already exempt (listed divergence). Fixable 833 → 0. |
| InArrayMissUse | `in_array($k, array_keys($a)[, true])` → `array_key_exists($k, $a)` (`'5'` is the int key 5; `'1.0' == 1`; `'abc' == 0` before 8.0). | Fix for int needles (strict) and non-numeric string literals (loose, 8.0+) only (listed divergence). Fixable 23 → 10. |
| ForeachInvariants | `for ($a = 0; $a <= count($delays); $a++)` (one more iteration) rewritten to `foreach`. | Only `<`, `>` and `!=`/`!==` conditions (Matomo retry loop). |
| StrEndsWithCanBeUsed, IsEmptyFunctionUsage | `substr($h, -strlen($n)) === $n` → `str_ends_with()` for possibly empty `$n` (`''` flips the result); `empty($xml)` → `$xml === null` for `SimpleXMLElement`/`GMP` (both can be empty). | Fix only for known non-empty needles; those classes excluded from D2b. |
| RealpathInStreamContext, MissingOrEmptyGroupStatement, ObGetCleanCanBeUsed | `realpath(__DIR__ . '/..cache')` → `dirname(__DIR__) . 'cache'`, `. ''` tails; braces inserted as `{\nstmt\n}` at column 0; a blank line with trailing spaces left by `ob_end_clean();`. | Whole `/..` segments only; formatted braces (`if ($a) {` + indented body); the statement takes its leading whitespace. |
| OffsetOperations | 736 reports, mostly `array\|bool` lookups (`Db::getRow()`), parameter defaults (`$tables = false` called with arrays), `$last = false` reassigned in a `foreach`, `bool` keys. | Fallback needs every value (`PossibleValuesComplete`); `bool` next to `array`/`string` is a failure marker; `bool` keys accepted. 736 → 322 (rest mostly loose `@return array<int\|array>` docs). |
| MissingIssetImplementation | PrestaShop importers' dynamic properties (`$entity->$field = $value`, then `isset($product->shop)`) reported as always false (error). | Skipped when the class has no `__set()` and the file writes the property or computed names. 65 → 2. |
| StaticInvocationViaThis | `$db->fetchAll()` with `Db::get(): Tracker\Db\|AdapterInterface\|Db` reported because `Piwik\Db::fetchAll()` is static. | Every class of a union must resolve the method as static. Matomo 421 → 303. |
| CallableParameterUseCaseInTypeContext | `function f($extra = false) { $extra = sprintf(…); }` reported against the default's type. | Untyped, undocumented parameters skipped (listed divergence). 246 → 229. |
| UnusedConstructorDependencies, LoopWhichDoesNotLoop | `#[ORM\Column] private int $id` assigned in the constructor only; `foreach ($user->groups as $g) {}` initialising lazy collections. | Attributes count as annotations; empty bodies over objects/unknown subjects skipped. ORM 26 → 8, 10 → 2. |
| MagicMethodsValidity, PhpUnitTests, PregQuoteUsage, MkdirRaceCondition | `_get()` implementing an abstract `Cache::_get()`; `@covers Composer\X::m` resolved relative to the test's namespace; `'{^' . str_replace('\\*', '.*', preg_quote($p)) . '$}'`; `if (!@mkdir($d)) { … if (is_dir($d)) return; … }`. | Inherited names skipped; tag names tried as fully qualified first; str_replace()/strtr() looked through; re-check in the failure branch accepted. |
| SuspiciousAssignments, PropertyInitializationFlaws, DisconnectedForeachInstruction | `$package['version'] = …; $loader->load($package);` then overwritten after the `if`; `private $removed = [];` removed although `$this->setPackages()` runs first; per-host `fwrite($fd, '# …')` and `mt_rand()` told to move out of the loop. | Reads of the holding array count; O skipped after calls that can run the object's code; stream-write/random/clock calls kept in the loop. |
| CryptographicallySecureRandomness, UsingInclusionOnceReturnValue, OnlyWritesOnParameter (shared walker) | Strength-flag advice on PHP 7.4+ (always true there); `@include_once $f;` (result discarded); `use ($org)` passed only to `new class ($org)` reported unused. | Skipped from 7.4; `@` looked through; anonymous-class arguments belong to the enclosing scope (`util.VarAccessesByName`). |

Deltas on the local corpora (HEAD 6c987f9 → this tree, engine changes
included): corpus A `src/` −1 and corpus B −2 (OffsetOperations `bool`
keys), corpus B +1 UnnecessaryCasting (engine narrowing); corpus C: one
TypeUnsafeArraySearch fix fewer. FuzzRules also found a nil dereference in
IsEmptyFunctionUsage on a recovery tree (`funCtion{{…emptY$s`), guarded
(seed and broken-PHP fixture added). New fixtures shifted FuzzRules' "every
third fixture" seed sample, and one engine branch (`brokenBy`, a ternary's
other branch) was covered only by a sampled fixture; that fixture is now a
permanent seed (`testdata/fuzz/FuzzRules/seed-unsafe-isset-shapes`).

Declined: ForgottenDebugOutput on `error_log()` (upstream default list, 12
of 29 samples deliberate logging; configurable); PotentialMalware on
Composer's `touch($target, filemtime($source))` (exactly the
timestamp-copying pattern the rule targets); RandomApiMigration's
`random_int()` ignores `mt_srand()` seeds (fixtures mix seeding and calls;
`SUGGEST_USING_RANDOM_INT` can be turned off); MagicMethodsValidity
"does not call parent::__construct()" for empty parent constructors (the
index records no body facts — engine request below, since addressed); AmbiguousMethodsCallsInArrayMapping
on stateful calls (`$map[$this->read($s)] = $this->read($s)`, not
detectable); UnnecessaryCasting removing `(int)` casts trusted from PHPDoc
(`@param array<int, int>`) in SQL-building code — the engine does not
record whether a type is native or documented (engine request below,
since addressed);
SuspiciousBinaryOperation D10, OneTimeUseVariables, SlowArrayOperationsInLoop,
SuspiciousLoop, UntrustedInclusion, DynamicInvocationViaScopeResolution,
SenselessProxyMethod, ClassOverridesFieldOfSuperClass, AlterInForeach,
UnnecessaryAssertion, ReturnTypeCanBeDeclared, JsonEncodingApiUsage and
ThrowRawException fixes (deliberate behaviour changes) — by design.
Engine-level causes left to the type-engine work (all addressed in the
Engine section, "Review round 2 engine causes"): `preg_replace()` with a
string subject typed `string|array|null`; `abs(int)` typed `int|float`;
PHPDoc intersections `A&B` read as unions; an early `return` after
`$this->p === false` not narrowing the property; writing `$a[1] = explode()`
types the absent `$a[0]` as `array`; target-version stub returns (`hash()`
is `string|false` on 7.x); `index.Method` without an empty-body flag and
without native-vs-doc provenance of types.

**Magento / Joomla / CakePHP / Yii2 / Laminas / Craft review
(2026-10-08).** Third run on code custos had never seen, shallow clones
plus `composer install --no-scripts --no-plugins` (vendor indexed, not
analysed), only read: magento/magento2 (25,860 files, PHP 8.3), Joomla CMS
(3,260, 8.1), cakephp/cakephp (1,748, 8.2), yiisoft/yii2 (1,078, 7.4),
craftcms/cms (1,708, 8.2) and eight Laminas packages (mvc, router,
servicemanager, eventmanager, view, form, db, validator: 62–383 files
each, 8.2), `analyse --all`. Robustness: no internal finding and no
syntax error; `php -l` of the target version reports nothing either,
except ten Yii test stubs written for PHP 8.0/8.1 (enums, union and
intersection types), which custos parses with its newest-grammar
fallback. Timing (`--stats`, index of vendor included): Magento 6.3 s,
Craft 1.3 s, Joomla 0.7 s (1.0 s once its `libraries/vendor` is indexed,
below), CakePHP and Yii 0.7 s, Laminas 0.2–0.7 s. Magento scales
linearly (1,624 files 1.6 s, 15,245 files 3.1 s, 25,860 files 4.6–6.3 s);
peak RSS is linear too (470 MB, 1.8 GB, 2.5 GB: about 90 KB per analysed
file plus the vendor index — sources are kept for the analysis pass).
The largest file, Craft's 1 MB icon table, analyses in well under a
second (its run time is the vendor index). About 1,100 findings were
sampled over 100 rules (all findings below 20, 15–20 otherwise;
error/warning severity and rules the earlier rounds had not sampled
first), and sampled fix diffs of about 75 fixable rules were read for semantic
changes. Fix safety: `custos fix --all` on copies (Magento 15,112 changed
files, all projects 19,600), then `php -l` of the target version on every
changed file; every fix was also applied alone and re-parsed
(`TestFixesKeepCodeParsable` on each project including its vendor,
`php -l` samples per rule). Before the fixes 6 changed files failed to
lint (UnnecessarySemicolon, SelfClassReferencing) and one vendor fix was
unparsable (MissingOrEmptyGroupStatement); after them every changed file
lints except the three Yii PHP 8.0 stubs that fail before fixing too.

| Rule / area | Was | Now |
|---|---|---|
| UnnecessarySemicolon | `<?= $x;` at the end of a file (no `?>`) told the `;` is stray; the fix was a parse error (Magento, Yii views). | Only before a closing tag. |
| SelfClassReferencing | `Sample&Menu` rewritten to `self&Menu` (compile error: self in an intersection). | Names inside intersection types skipped. |
| MissingOrEmptyGroupStatement | `for (…) // note` + body: the brace landed after the comment (commented out; the fix re-fired ten times). | Brace right after the header, body kept in place. |
| MkdirRaceCondition | `A \|\| !@mkdir($d)` → `A \|\| @mkdir($d) \|\| is_dir($d)` (error branch on success, Joomla thumbnails); statement `mkdir()` followed by `return is_dir($p) && …` turned into a throw. | Form follows the call's polarity (parenthesised, operator kept); later `is_dir()` re-checks accepted. 47 → 44. |
| ForeachInvariants | `$this->n = count($p)` written *after* the loop taken as the limit and deleted (Magento UPS); element writes (`$a[$i] = trim($a[$i]); echo $a[$i];`, `$a[$i + 1] = …`) read stale `$iValue`; `for ($i = 0, $acc = []; …)` lost `$acc = []`. | Property limits only from assignments before the loop, property writes never deleted; overlapping element writes, other header expressions → not reported. 39 → 34. |
| RealpathInStreamContext | `realpath(__DIR__ . '/../')` → `dirname(__DIR__) . '/'` (realpath() drops the separator; `$root . '/app'` became `//app`). | Trailing `/` dropped (listed divergence). |
| SlowArrayOperationsInLoop | Length hoisted into the initialiser although the body pops/pushes/reassigns the subject or calls a method on its object. | Reported without a fix then. |
| SubStrUsedAsArrayAccess | `substr($code, $i, 1)` with `string\|int $code` → `($code[$i] ?? '')` (`''` for an int). | No fix unless the source is string (or null). |
| OffsetOperations | 3,416 Magento findings from one `@return array\|int\|string\|float\|bool` test helper; `DOMNodeList`, `DOMNamedNodeMap`, `ResourceBundle` (native offset access, no `offsetGet()` in the stubs). | Doc unions admitting array/string with unconfirmed scalar members skipped (native type unknown); those classes supported. 3,973 → 455. |
| MagicMethodsValidity | Magento `_construct()` (called by `__construct()`), `_get`/`_set`, Cake `_call()` hooks reported as misspelt magic methods (error). | Skipped when the hierarchy declares the real magic method or the file calls the method by name. Magento 91 → 85. |
| ClassConstantUsageCorrectness | `use X\Filesystem; FileSystem::class` reported as returning the wrong string (PHP yields the import's spelling). | Only wrong-case imports reported (listed divergence). 15 → 3. |
| MissingIssetImplementation | Joomla `Table extends \stdClass`, `SimpleXMLElement` subclasses, classes with unindexed parents (error). | Skipped. 19 → 6. |
| SuspiciousAssignments | `@return array\|bool` narrowed to `array\|true` then destructured; `list($r, $g, $b) = sscanf($c, '#%02x…')`. | true/bool are failure markers next to array; engine: two-argument `sscanf()` is `array\|null`. 110 → 93. |
| PassingByReferenceCorrectness, PregQuoteUsage, UsingInclusionOnceReturnValue | `extract(array_merge(…))` (prefer-ref); `preg_quote('::')`, `preg_quote('\\')` (cannot contain a delimiter); `$found = (bool) include_once $p`. | Skipped (PregQuoteUsage listed divergence). 7 → 2, 51 → 42, 9 → 4. |
| CallableParameterUseCaseInTypeContext | `$path = Yii::getAlias($path)` (`string\|false`), `$key = $this->resize($key)` (`?string`) reported as bool/null: D7c dropped failure markers for plain function calls only. | Method, static and nullsafe calls too. 619 → 561. |
| GetClassUsage | `setObject($o = null) { if (is_object($o)) get_class($o) }`, `$t = $t ?: $this; get_class($t)` (D4b ignored flow). | D4b only when the flow type is unknown; `is_object()`/`is_a()` guards count. 12 → 9. |
| UselessUnset, OnlyWritesOnParameter, IssetArgumentExistence, LoopWhichDoesNotLoop | `extract($d); unset($d); include $tpl;` (Joomla dispatchers); `$spec = $e->getParam('spec'); $spec['x'] = 1;` (ArrayObject); `[$ts, $tz] = …` lazy init in a loop; `foreach ($this as $e) return false;` in Cake's CollectionTrait. | Skipped. 27 → 19 (UselessUnset). |
| UnknownInspection | 19 Craft suppressions of real PhpStorm inspections (`PhpIncompatibleReturnTypeInspection` …). | List extended. 19 → 1. |
| UnnecessaryCasting (T-rules typer) | `$f = $o->x; if ($c) { $f = 'x'; } (string) $f` taken as string (the unknown definition was dropped from the union); the fix changed behaviour. | A reaching definition of unknown type makes the variable unknown outside SpecOnly. 249 → 234. |
| Project index (`runner`) | Joomla's `config.vendor-dir: libraries/vendor` ignored: no dependency indexed. | composer.json `config.vendor-dir` honoured when it stays inside the project. Joomla gains true positives (e.g. `BaseApplication::__construct()` not calling the vendor parent's) and loses unresolved-class noise. |

Deltas on the local corpora (HEAD 3f240eb → this tree, default /
`--all`): corpus A `src/` −1 / −1 (MkdirRaceCondition, `DiagnosticLog`
re-checks with `is_dir()` after the call), corpus B 0 / 0, corpus C −1 /
−2 (SuspiciousAssignments on an `array|bool` reader result,
OffsetOperations on `InputBag::get()`'s documented scalar union).

Declined: NonSecureUniqidUsage's fix changes the identifier format
(23 characters with a `.`; Magento uses one in a temporary table name) —
that change is the advice itself; SuspiciousLoop on reassigned parameters,
ForgottenDebugOutput on deliberate `var_dump()`/`error_log()` in build
scripts, UnSafeIsSetOverArray, AutoloadingIssues (legacy helpers, entry
scripts), NullPointerException, SecurityAdvisories on `composer/composer`
and `yii2-debug` required at runtime, ProperNullCoalescingOperatorUsage
type-mismatch infos, MultipleReturnStatements, UnnecessaryAssertion
(`expects($this->any())`, 10,566 Magento tests) — noise by design;
EncryptionInitializationVectorRandomness on IVs from a helper returning
`openssl_random_pseudo_bytes()` or the previous CBC block (not
detectable); CallableParameterUseCaseInTypeContext on untyped closure
parameters (`$data` in a CakePHP `beforeMarshal` closure — the rule's
core case); OnlyWritesOnParameter element writes on parameters (same);
SenselessProxyMethod removing a PHPUnit test re-declared to change the run
order (deliberate fix, round 2); StrStrUsedAsStrPos with a `"0"` needle
(spec divergence, never seen); RandomApiMigration `rand($hi, $lo)`
(`random_int()` throws; arguments are ordered in every sample);
ReturnTypeCanBeDeclared `?array` where both branches of a final if/else
return (wider but safe, spec heuristic); MissingIssetImplementation's
remaining DOM/legacy cases and laminas-db `getMockForAbstractClass()` on
concrete classes (ClassMockingCorrectness, per spec).
Engine-level causes found by the review (since fixed, Engine section "Review round 3 engine causes"):
*do-while back edge* — `do { $all[] = $e; } while ($e = $e->getPrevious());`
with `\Exception $e`: element type expected `\Exception|\Throwable`, actual
includes null (GetClassUsage, Magento ExceptionHandler); *`array_pop()`
after `!empty(self::$stack)`* on `public static $stack = []`: expected the
element type, actual `…|null` (the static property is not narrowed
non-empty; Yii Widget); *inline `@var` on an unrelated statement* —
`/** @var \yii\base\Widget $class */ $config['model'] = …;` retypes the
string parameter `$class` to `\yii\base\Widget` (expected `string`;
InstanceofCanBeUsed, Yii ActiveField); *narrowing lost when the then
branch writes the guarded variable* — `if (is_int($c)) { $c = []; } else
{ $c['x']; }` with `array|int $c`: else type expected `array`, actual
`array|int` (OffsetOperations, Magento fixture); *builtin stub docs keep
pre-8.0 members* — `substr()` at 8.1 inside a ternary is typed
`string|false` because the stub's `@return string|false` doc is unioned
with the version-resolved native `string` (`TRules.declaredAndDoc`;
CallableParameterUseCaseInTypeContext, Joomla ExtensionManagerTrait);
*`@method static` tags* are stored as ordinary methods
(`index/extract.go`), so Craft's `@method static ActiveQuery hasOne()`
hides Yii's real instance `hasOne()` (StaticInvocationViaThis; needs a
`Method.Magic` flag); *exiting blocks still reach later uses* —
`$x = 5; if ($c) { $x = 'a'; return 1; } return (int) $x;`: expected
`int`, actual `int|string` (`Env.reaching`; UnnecessaryCasting and
ReturnTypeCanBeDeclared).

**Sylius / Shopware / API Platform / Mautic / Kimai / Akeneo review
(2026-10-08).** Fourth run on code custos had never seen, shallow clones
plus `composer install --no-scripts --no-plugins` (vendor indexed, not
analysed), only read: Sylius/Sylius (5,008 files, PHP 8.3),
shopware/shopware (12,222, 8.2), api-platform/core (3,162, 8.2),
mautic/mautic (4,817, 8.2), kimai/kimai (1,909, 8.2) and Akeneo
pim-community-dev (8,442, 8.3), `analyse --all`. Robustness: no internal
finding and no syntax error; `php -l` of the target version on every
analysed file reports nothing but PHP 8.3 deprecations in Akeneo (`${var}`
interpolation, optional-before-required parameters). Timing and peak RSS
(`/usr/bin/time -l`, index of vendor included): Kimai 0.7 s / 440 MB, API
Platform 1.2 s / 490 MB, Sylius 1.5 s / 615 MB, Akeneo 2.7 s / 920 MB,
Mautic 2.7 s / 870 MB, Shopware 3.1 s / 830 MB. The largest file, Mautic's
790 KB emoji table, analyses in 0.06 s (twice its size: 0.05 s). About
1,900 findings were sampled over 140 rules (all below 20, 15–20
otherwise; error/warning rules and rules least sampled by the earlier
rounds first; Symfony/Doctrine idioms — attributes, promoted and readonly
properties, enums, PHPDoc generics — checked on purpose). Fix safety:
`custos fix --all` on copies (82,440 edits in 11,343 files), then `php -l`
of the target version on every changed file: none broken, before and
after the fixes below; every fix was also applied alone and re-parsed on
each project including its vendor (`TestFixesKeepCodeParsable`: 428,000
fixes, 0 broken, `php -l` samples). Sampled fix diffs were read for
semantic changes (three, below).

| Rule / area | Was | Now |
|---|---|---|
| ReturnTypeCanBeDeclared | `: ArrayCollection` for Doctrine collection getters (`@var ArrayCollection` + `new ArrayCollection()` in the constructor), `matching()` results and repositories documented `@return ArrayCollection` but returning arrays: TypeError once Doctrine hydrates a `PersistentCollection`. | Not reported when the types include `ArrayCollection` (`Collection` still suggested). 67 removed; corpus C −64. |
| OffsetOperations | `iterable` taken as unsupported (`private iterable $handlers = []; $this->handlers[$k] = …`); Laravel's `Application $app; $app['config']` (interface); `$counts[$id] ?? 0` on native `array\|int`; index types `mixed\|Struct\|null` and `string\|void`. | `iterable` is array + Traversable; interfaces and non-final abstract classes without offset methods empty S; quiet reads (isset/empty/`??`) skipped when an array/string member exists; `mixed` indexes not checked, `void` dropped. 227 → 106 (error). |
| ClassMockingCorrectness | Final classes in parameters of private helpers of PhpSpec specifications (error, 13/13 FP). | Only `let`, `letGo` and `it…`/`its…` examples are doubled. 13 → 0. |
| SuspiciousAssignments | Switch fall-through that reads the previous case's value before overwriting it (`// no break` installer steps, error); parameter overwritten while `func_get_arg()` reads the passed value. | Targets read before the overwrite leave `W`; `func_get_arg(s)` count as reads. 38 → 28. |
| DisconnectedForeachInstruction | Discarded method chains (`$qb->where()->setParameter('k', $row);`, `$ctx->getConsole()->progressAdvance();`) not seen as modifying their root; statements after a conditional `continue`/`throw`; a variable reset at the top of the body (`$level = 1; if (…) { $level = …; }`). | Chain roots modified (as D8c already did for `$o->m();`); statements after a possible iteration exit and statements writing a shared variable are connected. 20 → 6 (14 FP removed, all TPs kept). |
| ClassOverridesFieldOfSuperClass | `/** @var ReportModel\|null */ protected $model;` over the parent's `@var Model\|null` told to drop the re-declaration (the only way to narrow a documented type). | Re-declarations whose `@var` differs from the inherited documented/declared type skipped. 111 → 65. |
| EmptyClass | `#[ApiResource]`, `#[ORM\Entity]`, `#[Get(…)]` configuration-only classes reported. | Any attribute exempts the class. 291 → 180; corpus B −42 (`#[ORM\Entity]` log classes). |
| CallableParameterUseCaseInTypeContext | `/** @var User $user */ $user = $repo->findOneBy([])` (`object\|null`) reported as type object; `match` of calls not treated like a call for failure markers. | Inline `@var` on the assignment trusted, `object` compatible with class parameters, all-call `match`/ternary drop the markers. 245 → 233. |
| MkdirRaceCondition | `@mkdir($p); $real = realpath($p);` then a test of `$real` turned into a throw (Kimai doctor page). | Later `file_exists()`/`realpath()`/`is_writable()` re-checks accepted; the scan stops at a reassignment of the path. 25 → 23. |
| RealpathInStreamContext (fix) | `$this->webRoot = realpath($r . '/../public'); if (!$this->webRoot) throw …` rewritten to `dirname()`, killing the existence check. | Reported without a fix when the result is tested for failure. |
| SimpleXmlLoadFileUsage | Bug #62577 advice (error, with a rewrite) on PHP 8.x targets. | Only below 8.0 (the trigger, `libxml_disable_entity_loader()`, is deprecated from 8.0). 14 → 0; corpus A −1. |
| DateIntervalSpecification | `$v = $q['v'] ?? ''; if ($v === '') throw …; new DateInterval($v)` reported `''` (error). | A discovered literal only when it is the only discovered value. |
| ThrowRawException, ClassMethodNameMatchesFieldName, OnlyWritesOnParameter | Exceptions whose constructor defaults the message (`AccessDeniedException()`); promoted properties documented by the constructor's `@param` reported as "type unknown"; element writes on a `mixed` local (`ReflectionProperty::getValue()`). | Constructor default message counts as a preset; the constructor `@param` types promoted properties; `mixed` counts as a possible object. 531 → 515, 3 → 0, 37 → 36. |
| UnnecessaryCasting (T-rules typer) | `foreach ($p as $v) { if ($c) { $v = 'x'; } (string) $v; }` typed string (the unknown element dropped); the fix removed a needed cast under strict types. | An unknown foreach binding reaching the read keeps the variable unknown (`infer/trules.go`). 106 → 105. |
| NotOptimalRegularExpressions, `custos fix --diff` | `[^\s]` → `\S` hinted "matches more under /u" (same set in every mode); diff headers `a//abs/path`. | "same result"; leading `/` trimmed. |

Deltas on the local corpora (HEAD 0ae0ff1 → this tree, default /
`--all`): corpus A `src/` −1 / −2 (SimpleXmlLoadFileUsage at 8.4; one
DisconnectedForeachInstruction after a guarded `continue`), corpus B 0 /
−43 (42 EmptyClass `#[ORM\Entity]` log classes, one OffsetOperations on
an `iterable` property, one DisconnectedForeachInstruction after a guarded
`continue`), corpus C −69 / −75 (64 ReturnTypeCanBeDeclared
`: ArrayCollection` — 14 of them repository methods documented
`@return ArrayCollection` but returning `getResult()`/`execute()` arrays, a
TypeError — and six
DisconnectedForeachInstruction `$this->getDoctrine()->getManager()->flush()`
per-iteration flushes, now unreported like `$em->flush();` already was;
four regex messages reworded).

Declined: ForgottenDebugOutput on `Doctrine\Common\Util\Debug::export()`
nested in `print_r(…, true)` (spec default list; three legacy Behat
contexts); TraitsPropertiesConflicts on a trait property re-declared by
constructor promotion (compatible, info, the redundancy is the point);
CascadeStringReplacement `array(…)` in the rebuilt call
(`USE_SHORT_ARRAYS_SYNTAX`); PackedHashtableOptimization on `'0' =>`
keys, CryptographicallySecureRandomness `openssl_random_pseudo_bytes()` →
`random_bytes()`, PregQuoteUsage on delimiter-less regex consumers
(MongoDB, Varnish), UsingInclusionReturnValue on config includes,
NestedTernaryOperator, NotOptimalIfConditions, SuspiciousLoop on reused
parameters, AutoloadingIssues on fixtures and scripts, UnserializeExploits
in `Serializable::unserialize()`, MagicMethodsValidity on decorators that
skip the parent constructor, UnnecessaryCasting `(string)
$_SERVER['X']` (spec typing) — noise by design;
NestedAssignmentsUsage F2 copying a value past a typed-property coercion
(`$z = $m->floatProp = 1`; theoretical, never seen); UnqualifiedReference
breaking Shopware's `eval`-defined namespace function mocks (`IniMock`;
not detectable); DateTimeConstantsUsage `format()` output `+02:00` vs
`+0200` (the rule's intent); JsonEncodingApiUsage, TypeUnsafeArraySearch,
NonSecureUniqidUsage and RandomApiMigration fixes (decided earlier).
Engine-level causes found by the review (since fixed, Engine section "Review round 4 engine causes"): *native
`iterable` refined by docs* — `/** @var array<string, X> */ private
iterable $p` and `@return Collection<Activity>` on `: iterable` stay
`iterable` (expected `array` / the collection class; OffsetOperations);
*partial generic arguments* — Shopware `Collection` (`@template TElement`,
`@template TKey of array-key = array-key`, `@implements
IteratorAggregate<TKey, TElement>`) extended as `Collection<LineItemQuantity>`
types `foreach ($c as $key => $v)` `$key` as `\LineItemQuantity`
(expected `int|string`); *boolean alias narrowing* — `if (($isObject =
is_object($r)) && …) {} if ($isObject) {} else { $r }` with `object|string
$r`: expected `string`, actual `object|string`; *property guard with
`continue`* — `foreach ($m as $ref) { if (!is_string($ref->value))
continue; str_replace('_', '-', $ref->value); }`: result expected
`string`, actual `array|string`; *native `: mixed` return* typed from the
body (`float|array|int|string`) instead of `mixed`; *absent literal keys*
— `$rows = ['total' => count($x)]; $rows['meta']['s'] = 1;` types
`$rows['meta']` as `int` (round-1 item, still open); *`str_replace()` on
an unknown subject* typed `array|string` (expected unknown);
*anonymous classes* — `new class implements TranslatorInterface {…}` typed
`object` (expected to include the interface); *promoted properties* have
no `DocType` from the constructor's `@param`
(`index/extract.go`, constructor promotion; ClassMethodNameMatchesFieldName
works around it locally).

**TYPO3 / MediaWiki / Moodle / phpBB / Flarum / Pimcore review
(2026-10-08).** Fifth run on code custos had never seen, in other styles
(legacy procedural code with `global $CFG, $DB`, includes, generated API
clients, PHP 4-era libraries bundled by Moodle), shallow clones plus
`composer install --no-scripts --no-plugins` (vendor indexed, not
analysed), only read: typo3/typo3 (6,726 files, PHP 8.5),
wikimedia/mediawiki (5,660, 8.3), moodle/moodle (55,457, 8.3), phpbb/phpbb
(1,899, 8.2), flarum/framework (1,777, 8.3) and pimcore/pimcore (1,927,
8.4), `analyse --all`. Robustness: no internal finding and no syntax
error; `php -l` of the target version on every analysed file reports only
PHP 8.4 deprecations in Pimcore tests. One super-linear file: Moodle's
900 KB `tcpdf.php` took 4.8 s (OffsetOperations →
`PossibleValuesComplete` rescanned the whole 25,000-line class for every
`$this->p[…]`); with a per-class index of property declarations and
writes it takes 0.4 s (complexity test added); the other large files
(1.4 MB phpBB CJK table, 1.3 MB AWS data, 870 KB Google client) analyse in
0.1–0.4 s and scale linearly when doubled. Timing and peak RSS
(`/usr/bin/time -l`, vendor index included, final tree): Flarum 0.9 s /
315 MB, Pimcore 1.1 s / 500 MB, phpBB 0.5 s / 400 MB, TYPO3 1.5 s /
740 MB, MediaWiki 2.2 s / 1.2 GB, Moodle 11.9 s / 5.4 GB (24 s before the
tcpdf fix); Moodle is linear (224 files 1.2 s, 2,644 files 2.2 s, 47,051
files 19.5 s and 55,457 files 24 s before the fix; about 90 KB per file).
About 1,700 findings were sampled over 150 rules (four parallel reviews:
error rules all, warning rules 15–20 each, info rules 10–15 each,
semantic rules and legacy idioms first). Fix safety: `custos fix --all`
on copies (508,000 edits in 52,600 files), then `php -l` of the target
version on every changed file: none broken, before and after the fixes
below. Version gating: Moodle, phpBB and MediaWiki fixed with `--php 7.4`
(and Moodle with `--php 5.6`), then every changed file that linted before
was linted with PHP 7.4 (5.6): no new failure. Every fix was also applied
alone and re-parsed on each project including its vendor
(`TestFixesKeepCodeParsable`, `php -l` samples; 680,000 fixes): three
fixes broke parsing on Moodle (adodb, vendored phpxmlrpc), 0 after the
fixes below. Four further fixes changed behaviour (below).

| Rule / area | Was | Now |
|---|---|---|
| OffsetOperations (performance, `util.PossibleValuesComplete`) | 4.8 s on tcpdf.php: property declarations and writes rescanned per lookup. | Per-class index memoised on the file; 0.4 s. |
| IsEmptyFunctionUsage (fix) | `empty($user)` → `$user === null` when `$user` came from a method documented `@return stdClass` that returns `get_record()` (`stdClass\|false`): a failed lookup passed Moodle's access check. | Fix only when native declarations alone give the same types; otherwise reported without a fix. Fixable 363 → 175. |
| RealpathInStreamContext (fix) | `@realpath($this->symlinkToCoreFiles . '/../')` → `dirname(…)`: the parent of the symlink instead of the parent of its target (TYPO3 core updater). | Fix only for `__DIR__`, `__FILE__` and `dirname()` of them. Fixable 21 → 17. |
| StringCaseManipulation (fix) | `strpos(strtolower($name), $query)` → `stripos($name, $query)` (matches queries with capitals, the original never did). | Fix only when both sides are converted the same way or the other side is a literal without opposite-case letters. Fixable 8 → 4. |
| CascadeStringReplacement (fix) | `$s = str_replace('a', 'b', $s); $s = str_replace('x', strlen($s), $s);` merged (second arguments evaluated on the original). | Not linked when the later search/replace mentions the subject. |
| GetTypeMissUse, PrintfScanfArguments (`util.PossibleValuesComplete`) | `gettype($v) !== $this->valueType` → `!is_int($v)` although `FloatDef` redeclares `'double'` (MediaWiki); adodb `$dropIndex` formats checked against the base class's default; `$mode = 'array'` parameter defaults taken as the only value. | `$this->p` complete only for private properties or final/anonymous classes; GetTypeMissUse needs a complete single value (listed divergence for one upstream printf case). 12 → 9, 15 → 8. |
| PrintfScanfArguments | `sscanf($t, 'PT%dH%dM%dS') ?? []` (two-argument form used as a value) reported as missing arguments. | Any value use counts; discarded calls and truth-value uses still reported. |
| MissingIssetImplementation | `isset($fault->errorcode)` on `Exception`, `isset($node->tagName)` on `DOMNode` (subclasses declare them; error, 12/16 FP). | Descendants declaring the property, `__isset()` or dynamic properties exempt the check. 16 → 6. |
| UnusedConstructorDependencies | Moodle `cm_info` properties read by `__get()` returning `$this->$name`. | Computed `$this` reads, `get_object_vars($this)`, `foreach ($this …)`, `(array) $this` exempt the class. 111 → 81. |
| OnlyWritesOnParameter | phpBB's `extract(trigger_event(…, compact($vars)))`; `${…}[] = …` writing by-reference imports. | Non-literal `compact()` and variable variables read every variable. 214 → 199. |
| UselessUnset | `unset($userid)` so a later `!empty($userid)` is false; `unset($config)` in a loop before `$config[$k] = …`. | Not reported when a later read can see the unset. 28 → 23. |
| DisconnectedForeachInstruction | `echo $OUTPUT->box_start()`, progress dots per row; `StringHelper::stringIncrement($column)` (by-reference method parameter). | Output is per-iteration (listed divergence); D8b covers method, static and constructor calls. 33 → 23. |
| ReturnTypeCanBeDeclared | `if ($t) { $r = 'X'; } return $r;` and switch without default → `: string` (null returned). | Null added when the plain-assigned local may be unassigned. |
| ReturnTypeCanBeDeclared (engine interplay) | Engine now types Doctrine `getResult()` as `mixed`; 43 corpus C `@return array` repository methods lost `: array`. | A `mixed` returned value is treated as unknown, so the `@return` tag decides (ArrayCollection exclusion kept). Moodle +14,481 (14,274 in the generated Google API client, `call()` wrappers documented with their class), corpus C +8. |
| CallableParameterUseCaseInTypeContext | `@iconv(…)` not seen as a call; `filemtime()` `int\|false`, `fopen()` `resource\|false` reported as bool; `?callable` assigned an invokable Guzzle handler; `$grade = 1` on `float`. | `@` looked through; false is a marker next to any type; `__invoke()` classes are callable; int fits float. 1,007 → 962. |
| StaticInvocationViaThis | `$this->reader->open($f)` on XMLReader (static in the stubs, opens the instance). | `XMLReader::open()`/`XML()` exempt. |
| SuspiciousBinaryOperation | `wfRandom() == wfRandom()`, `Asset::getById($id) === Asset::getById($id)` "both operands are the same" (error). | Operands that may vary (calls to methods, user functions, random/clock built-ins, `new`, `++`) skipped. |
| PhpUnitTests | `@covers \clean_param` (a global function) reported as an unresolvable class (error). | Function names accepted. 3,091 → 3,041. |
| IssetArgumentExistence | `isset($prev)` in an inner loop with `$prev` set later in the outer loop (tcpdf; error). | Every enclosing loop checked. |
| ClassMockingCorrectness | `$mb = $this->getMockBuilder(Abstract::class); $mb->getMockForAbstractClass();` told to use getMockForAbstractClass(). | Stored builders followed; returned/passed ones skipped. 4 → 2. |
| NotOptimalRegularExpressions | `preg_quote('/test/…', '/')` told the /e flag was removed; `(\s+(?:unsigned\|zerofill))*` reported as `(\s+)*`. | preg_quote() text only gets D20; mandatory inner groups kept as an atom. |
| OffsetOperations | `simplexml_load_string()` (`SimpleXMLElement\|false`) offsets; `@return string[]\|string` indexes (error). | false next to a class is a failure marker; loosely documented index unions skipped. 358 → 276. |
| NestedAssignmentsUsage, MissingOrEmptyGroupStatement (fix, found by `TestFixesKeepCodeParsable` on Moodle) | `if ($c) $n = $_SESSION['k'] = 10; else …` split into two statements (the `else` detached: parse error, adodb); `if ($h) echo "x" ?>` (statement ended by the close tag) got its closing brace after `?>`, in the HTML (phpxmlrpc). | Split statements of a brace-less body are wrapped in braces; the brace goes before the close tag and the statement gets a `;`. |
| MultiAssignmentUsage, ReferencingObjects, PregQuoteUsage | `$q =& $m[6]; $t =& $m[7];`; MediaWiki's `HookRunner` (540 interfaces, beyond the index's 256-ancestor cap) told to drop `&` (fatal signature mismatch); `preg_quote(Packer::PREFIX)` with a delimiter-free constant. | By-reference pairs skipped; capped hierarchies are unknown; constants resolved. 414 → 398, 70 → 66. |

Deltas on the local corpora (HEAD 7c1e498 → this tree, engine changes
included, default / `--all`): corpus A `src/` 0 / 0, corpus B 0 / 0, corpus C
+7 / +7 (eight ReturnTypeCanBeDeclared suggestions from `@return` tags over
`mixed` results — Doctrine repositories and `VacationDays::first()` — and
one CallableParameterUseCaseInTypeContext `mb_ereg_replace()` result
`string|false|null` no longer reported as null).

Declined: UntrustedInclusion and MultipleReturnStatements on Moodle's
`require_once($CFG->dirroot . …)` and legacy functions,
NonSecureUniqidUsage, ForgottenDebugOutput (`error_log`, `xdebug_*`),
SuspiciousLoop on reused parameters, SlowArrayOperationsInLoop,
UnserializeExploits and UnSafeIsSetOverArray, AutoloadingIssues on
Moodle's frankenstyle `classes/` and phpBB scripts, NestedTernaryOperator,
TypeUnsafeComparison fixes trusted from `@param string`, SecurityAdvisories
on `composer/composer` and testing packages, TraitsPropertiesConflicts,
CryptographicallySecureRandomness (`openssl_random_pseudo_bytes`),
ClassMethodNameMatchesFieldName (Flarum fluent setters),
AmbiguousMethodsCallsInArrayMapping on getters, PotentialMalware in dev
tools — noise by design or decided earlier; DeprecatedIniOptions reads of
`mbstring.func_overload` under a `PHP_VERSION_ID < 80000` guard (two
vendored libraries); FixedTimeStartWith overlapping StrStartsWithCanBeUsed
on 8.x targets (disabled by default); UsingInclusionOnceReturnValue on
`array_map(fn ($f) => require_once …)` with a discarded result (one case);
OffsetOperations on wrong legacy PHPDoc (`@var integer` over an `array()`
default — true to the docs); Moodle `*_test.php` files not recognised as
tests (test detection decided earlier). Engine-level causes found by the
review (since fixed, Engine section "Review round 5 engine causes"): *`extract()`,
variable variables and one-argument `parse_str()` keep local types* —
`function g($c, array $v) { $n = (int) $c; extract($v); return (int) $n; }`:
`$n` expected unknown, actual `int` (UnnecessaryCasting, TypeUnsafeComparison
fix; phpBB event dispatch); *possibly undefined variable* — `if ($t) { $r
= 'X'; } return $r;`: expected `string|null`, actual `string` (worked
around in ReturnTypeCanBeDeclared); *256-ancestor cap is silent* —
`index.Ancestors` truncates without a flag (worked around in
ReferencingObjects); *SpecOnly T-rules typer* — `if ($c) { $sql =
array_shift($c); return $sql; } $y = $sql; str_replace('a', 'b', $y)` with
`string $sql`: expected `string`, actual `array|string`; *doc pseudo-type
`number` shadows a class* — `use App\Number; @param array|Number $v`:
expected `array|\App\Number`, actual `array|int|float` (SuspiciousAssignments
on scssphp); *`class_alias()` not indexed* — `class_alias(user::class,
\core_user::class)`: `\core_user` expected to resolve (PhpUnitTests
`@covers \core_user::…`); *elseif instanceof chain* — `@param string|Code
$code; if ($code instanceof Lang) {…} elseif ($code instanceof Code) {
$code = 'x'; }`: after the if expected `string`, actual `string|\Code`;
*`var_export($x, true)` / `print_r($x, true)`*: expected `string`, actual
`string|null` / `string|true` (MagicMethodsValidity `__toString`);
*`assert($b instanceof X)`* apparently not narrowing (unconfirmed).

**Firefly III / Monica / Bagisto / Dolibarr / Roundcube / FreshRSS
review (2026-10-08).** Sixth run on code custos had never seen: three
Laravel applications and three legacy, mostly procedural code bases
(`global $conf, $db, $langs`, templates included by pages, bundled PHP 4-era
libraries), shallow clones plus `composer install --no-scripts
--no-plugins` where a vendor exists (vendor indexed, not analysed), only
read: firefly-iii/firefly-iii (1,743 files, PHP 8.5), monicahq/monica
(1,650, 8.3), bagisto/bagisto (3,053, 8.4), Dolibarr/dolibarr (4,350, 7.1,
its minimum), roundcube/roundcubemail (590, 8.1) and FreshRSS/FreshRSS
(611, 8.1), `analyse --all`. Robustness: no internal finding and no syntax
error; `php -l` of the target version on every analysed file reports
nothing (Dolibarr included, at 7.1). No super-linear file: the slowest
files with the project index loaded are tcpdf.php (0.38 s, 900 KB) and
Dolibarr's 1.6 MB CJK font tables (0.15 s); a synthetic 8,000-branch
elseif chain scales linearly. Timing and peak RSS (`/usr/bin/time -l`,
vendor index included, final tree): Firefly 1.1 s / 560 MB, Monica 0.8 s / 340 MB, Bagisto 1.3 s / 590 MB, Dolibarr 3.4 s / 1.4 GB, Roundcube 0.6 s / 195 MB, FreshRSS 0.3 s / 150 MB (times from the idle first run; under the final load average of 30–40 they were 1.3–1.6× longer). Dolibarr first peaked at
3.0 GB: every finding (273,000 with `--all`) kept its quick-fix closure
and with it the file's syntax tree and type environment until the report;
`custos analyse` now drops the edit closures as each file finishes
(`runner.RunReport`; fix titles kept, output identical), 1.4 GB. About
1,500 findings were sampled over 140 rules (four parallel reviews: error
rules all or 15–20 each, warning rules on the Laravel and on the legacy
projects, info and semantic rules 10–15 each; version gating checked on
Dolibarr's 7.1 target: no suggestion needs a newer PHP). Fix safety:
`custos fix --all` on copies (256,000 edits in 6,550 files, Dolibarr
232,800), then `php -l` of the target version on every changed file: none
broken, before and after the fixes below; every fix was also applied alone
and re-parsed on each project including its vendor
(`TestFixesKeepCodeParsable`, `php -l` samples; 474,000 fixes): one
Dolibarr file broke when all its fixes were combined, 0 after the fix
below. Four fixes changed behaviour and one dropped a type annotation
(below).

| Rule / area | Was | Now |
|---|---|---|
| `custos analyse` memory (`internal/runner`) | Quick-fix closures of every finding retained each file's tree until the report: Dolibarr `--all` 3.0 GB. | Report mode drops the edit closures per file: 1.4 GB, same output. |
| DynamicInvocationViaScopeResolution (fix) | `self::_initTags()` in printipp's `BasicIPP` constructor → `$this->_initTags()`, which runs `CupsPrintIPP`'s override (also `Ancestor::m()` while the class overrides `m`). | Fix only when the name resolves to the called method and no subclass can override it (private/final method, final class or enum, no indexed descendant declaring it); traits never; otherwise reported without fix (listed divergence). |
| NotOptimalRegularExpressions (fix, D22d) | `preg_replace('/__HANDLER__/i', "'" . $db->escape($h) . "'", $sql)` → `str_ireplace(…)`: preg_replace() collapses `\\` and expands `$0`/`\0` in the replacement, str_replace() does not (Dolibarr SQL templates). | Only for literal replacements without `\` and `$`, numbers and int/float casts. 1,340 → 1,259 fixable. |
| MkdirRaceCondition (fix) | `if (!@mkdir($lock_path)) { return true; } … rmdir($lock_path);` (FreshRSS migrator lock) → `&& !is_dir($lock_path)`: every process takes the lock. | A tested mkdir() whose function also rmdir()s the same path is a lock: not reported (an ignored `mkdir($scratch)` still is). |
| OneTimeUseVariables (fix) | `/** @var \Illuminate\Auth\RequestGuard */ $guard = $this->auth->guard('sanctum'); return $guard;` inlined, dropping the type; the doc deletion left an indentation-only line. | A nameless `@var` right before the assignment counts (E4/D8); F1 deletes up to the statement. 302 → 300. |
| MagicMethodsValidity | Restler's `iFilter::__isAllowed()`, `iAuthenticate::__getWWWAuthenticateString()` implementations reported as misusing the reserved prefix (error, 14). | Names imposed by an interface or parent skipped, like single-underscore names. |
| UsingInclusionOnceReturnValue | `$found = @include_once $dir . $f; if ($found) break;`, `$res = include_once $f; if (!$res) die();` (error). | A local only tested for success (or never read inside a function) is a success flag. 36 → 18. |
| ReturnTypeCanBeDeclared + DeprecatedConstructorStyle (combined fixes, found by `TestFixesKeepCodeParsable` on Dolibarr's phan stubs) | At PHP 8.x `function Mail_mime()` got `: void` while DeprecatedConstructorStyle renamed it: `__construct(): void` (fatal). | Methods shaped like PHP 4 constructors are skipped at every version. |
| PrintfScanfArguments | `sprintf("%.0lf", $v)`, `%ld` reported as malformed (PHP accepts and ignores `l`; nusoap). | `l` accepted. |
| NotOptimalRegularExpressions (D13b) | `'/^PhpOffice\\\PhpSpreadsheet\\\/'` (an escaped backslash, then P) told `\P` needs /u (error); `'/\\p/'` (the pattern `\p`) missed. | Escapes searched in the decoded pattern (listed divergence). |
| OffsetOperations | `$folders[0]` on Webklex `FolderCollection` → `PaginatedCollection` → unindexed `Illuminate\Support\Collection` (error). | Unresolvable ancestors empty S like an unresolvable class. 110 → 108. |
| IsEmptyFunctionUsage | `/** @var Conf $conf */ if (empty($conf) \|\| !is_object($conf)) exit;` in Dolibarr templates → `$conf === null` ("Undefined variable" in exactly the guarded case; about 120), and file-scope variables assigned only in an `if`, with a fix. | At file scope the top-level statements must assign the variable first, as in functions. 216 → 67 (fixable 24 → 14). |
| UnusedConstructorDependencies | SimplePie `File` stores `$this->permanentUrlMutable` and re-runs `$this->__construct()` on a redirect, which reads it. | Constructors calling `$this->__construct()` exempt the class. 9 → 7. |
| PropertyCanBeStatic | `protected $options = [...]` with `$this->options['table_prefix'] = $prefix` per connection (Roundcube `rcube_db`) told to become static (shared between instances). | Properties written through `$this` (plain, element, compound, `++`, `unset`) skipped. 122 → 111. |
| DisconnectedForeachInstruction | `var_dump($matches)` / `print_r()` in a loop; `foreach (range(1, 10) as $attempt) { post(…); }` (throttle tests) told to move statements out. | Those calls are output (per iteration); loops that never read their key/value variable repeat their body and are skipped (unless compact/extract/`$$`/include/eval). 16 → 7. |

Deltas on the local corpora (HEAD 6294e7d → this tree's rule changes
alone, default / `--all`): corpus A `src/` 0 / 0, corpus B 0 / −1 and
corpus C 0 / −2 (PropertyCanBeStatic on properties with setters or element
writes: `DatesComponent::$days`/`$months`, a test's `$sections`).

Declined: SecurityAdvisories on a tool-install manifest
(`.ci/php-cs-fixer/composer.json`, one case); NullPointerException in Pest
`*Test.php` closures (disabled by default, every dereference reported by
spec); DisconnectedForeachInstruction on a call that depends on state an
earlier connected statement changed (`$rcube->config->set(…)` then
`invokeMethod($plugin, '_init_driver')`; not detectable without effect
analysis); StaticClosureCanBeUsed on Laravel `macro()` closures (current
`Macroable` survives static closures); StaticInvocationViaThis following a
project's wrong `/** @var Storage $disk */`; OffsetOperations and
UnsupportedStringOffsetOperations on wrong legacy PHPDoc (nusoap
`@return false`, `@var resource`), UntrustedInclusion,
MultipleReturnStatements, NonSecureUniqidUsage, ForgottenDebugOutput
(`error_log`), UnserializeExploits, SuspiciousLoop, TypeUnsafeComparison
fixes trusted from `@return string` (Dolibarr's GETPOST), AutoloadingIssues
on FreshRSS controllers and Roundcube actions, FixedTimeStartWith,
UnSafeIsSetOverArray, NestedTernaryOperator — decided earlier or noise by
design. Engine-level causes found by the review (since fixed, Engine
section "Review round 6 engine causes"): *index body facts* — MagicMethodsValidity "does not
call parent::__construct()" on parents whose constructor only stores its
parameters (`$this->db = $db;`) or promotes them in an empty body, when
the child does the same (about 280 of Dolibarr's 424 errors: DolibarrModules,
CommonDocGenerator, ModeleBoxes, CommonObjectLine; Bagisto `AbstractType`):
`index.Method` needs a summary such as "body only assigns parameters to
same-named properties" and an empty-statements flag independent of
promotion; and EncryptionInitializationVectorRandomness's wrapper
exemption works only in the same file (`rcube_utils::random_bytes()`
calling `random_bytes()`, declared elsewhere) — needs a "returns a CSPRNG
call" fact; *`Safe\` functions* — `use function Safe\preg_replace;
$s = preg_replace('/\s+/', '', $s)` with `string $s`: expected `string`,
actual `array|string` (docblock `string|array|string[]`; 6
CallableParameterUseCaseInTypeContext on Firefly/Monica; type them like the
builtin minus false/null); *`include`/`require` keep local types* —
`$total = '0'; include 'x.php'; strlen($total)`: expected unknown, actual
`string` (UnnecessaryCasting fixes removing `(int)` in Dolibarr SQL after
`require '../../main.inc.php'`, Roundcube `$config = []; require $file;
(array) $config`); *T-rules typer loop back edges* — `$t = '0'; foreach
($rows as $r) { bcadd((string) $t, '1'); if ($r) { $t = null; } }`:
expected `string|null` at the cast, actual `string` (Firefly
`PiggyBankEnrichment`, a TypeError under strict_types once the cast is
removed; the main typer is right); *`pathinfo($f, PATHINFO_EXTENSION)`*:
expected `string`, actual `array|string` (5 OffsetOperations "index of
type array"); *`gettimeofday()`* without argument or with `false`:
expected `array`, actual `float|array` (8 OffsetOperations); *`??=` on an
absent key* — `$r = ['count' => 0]; $r[$id] ??= []; $r[$id]['a'] = 1;`:
`$r[$id]` expected `array`, actual `array|int` (3 on Firefly); *negated
`is_object()` on a doc-only class* — `/** @param Translate $langs */
!is_object($langs) && $langs == 'es_MX'`: expected unknown/empty, actual
`\Translate` (TypeUnsafeComparison error); *optional regex groups* —
`preg_match('/(a)(.+)*?(b)?/', $s, $m)`: `$m[3]` expected `string|null`
(may be missing), actual `string` (UnnecessaryCasting on PHPMailer).

**SuiteCRM / EspoCRM / Kanboard / Grav / October CMS / Koel review
(2026-10-08).** Seventh run on code custos had never seen: a very large
legacy SugarCRM descendant, two CRMs/project tools with bundled libraries,
a flat-file CMS and two Laravel applications, shallow clones plus
`composer install --no-scripts --no-plugins --no-dev` (vendor indexed, not
analysed; Kanboard commits its vendor), only read: salesagility/SuiteCRM
(4,653 files, PHP 8.1 — its composer platform; the 7.x line now requires
8.1), espocrm/espocrm (3,395, 8.3), kanboard/kanboard (1,671, 8.1),
getgrav/grav (705, 8.3), octobercms/october (1,682, 8.2) plus its vendored
october/rain library analysed as a project of its own (438, 8.2, October's
vendor indexed) and koel/koel (1,573, 8.3), `analyse --all`. Robustness:
no internal finding except the 10,000-findings cap on SuiteCRM's
`install/demoData.en_us.php` (UnNecessaryDoubleQuotes, by design); the
only syntax errors are SuiteCRM's two code-generator templates
(`class <module_name>Dashlet …`), which `php -l` rejects too; `php -l` of
the target version on every analysed file reports nothing else. One
super-linear file: SuiteCRM's vendored `google/apiclient/src/Model.php`
(10 KB) took 0.34 s — ReturnTypeCanBeDeclared, DynamicInvocationViaScopeResolution
and ReferencingObjects walked all descendants of the class once per
method, and the Google API services hold thousands of generated
subclasses; a per-class memo of descendant methods
(`util.DescendantMethods`) makes it linear (complexity test added;
ReferencingObjects also loses its 1,000-step cap, which answered "not
overridden" on large hierarchies). The other large files (tcpdf.php 0.25 s,
1.6 MB CJK font tables 0.13 s, nusoap, pChart) scale linearly. Timing and
peak RSS (`/usr/bin/time -l`, vendor index included): SuiteCRM 2.4 s /
1.6 GB (77,500 findings), EspoCRM 1.2 s / 400 MB, Kanboard 0.4 s / 140 MB,
Grav 0.4 s / 155 MB, October 1.0 s / 280 MB, rain 0.75 s / 215 MB, Koel
1.35 s / 375 MB. About 1,700 findings were sampled over 150 rules (five
parallel reviews: error rules all or 15–20, warning rules in two halves,
never-sampled info rules 10–15, the other info rules 10–15; verdicts on
the target versions, no suggestion needs a newer PHP), and the fix diffs of
all 91 fixable rules were read for semantic changes (a sixth review: all
hunks when few, else 15–60 plus pattern scans; UnNecessaryDoubleQuotes'
56,748 changed literals decoded and compared). Fix safety: `custos fix
--all` on copies (102,200 edits in 6,290 files), then `php -l` of the
target version on every changed file: none broken, before and after the
fixes below; every fix was also applied alone and re-parsed on each
project including its vendor (`TestFixesKeepCodeParsable`, `php -l`
samples; 0 broken). Fourteen fixes changed behaviour (ten on the corpus).

| Rule / area | Was | Now |
|---|---|---|
| Descendant walks (ReturnTypeCanBeDeclared, DynamicInvocationViaScopeResolution, ReferencingObjects) | One walk over all indexed descendants per method: n methods × n subclasses (Google API `Model.php` 0.34 s). | `util.DescendantMethods`, memoised per class and file; linear (test at 4× size: ×4 instead of ×16). |
| StaticClosureCanBeUsed (fix) | Keyed closures in file-level arrays made static: Grav's `updates/*.php` return `['postflight' => function () {…}]`, run with `$closure->call($this)` (a static closure only warns: 21 migrations would silently stop); closures passed to facades (`Cache::extend('x', fn …)` binds them). | Keyed array values unsafe at any level; a static call is safe only for a declared static method (magic `@method static` or a class with `__callStatic()` count as unknown targets). 580 → 326 (listed divergence: two EA fixtures). |
| PhpUnitTests / PhpUnitDeprecations (fix) | Without an indexed PHPUnit the version fell back to 8.0: EspoCRM (^11.5) got `assertContains()` for `in_array()` (strict from 9.0), Grav `assertRegExp()` (removed in 10); 29 fixes. | The CLI passes the lowest version allowed by composer.json's `phpunit/phpunit` (require-dev, else require) as a fallback. |
| ReturnTypeCanBeDeclared (fix) | `new static` / `clone $this` in a trait typed as the trait: `instance(): Singleton` on October's Singleton trait (every call throws; 9 fixes); fixes resting on a callee's `@return`, a `@param` or a property's `@var` (Kanboard `getProgress(): int` over `round(…, 1)`; `/** @param string $v */ f($v = null)`). | Traits are never suggested; the fix is offered only when the native typing (PHPDoc ignored) gives the same type for every returned value — the 2026-10-08 "@return alone" policy extended to every PHPDoc source. Reports unchanged; fixable 2,515 fewer on the corpus, corpus C 1,569 → 1,091. |
| PreloadingUsageCorrectness (fix) | EspoCRM `preload.php`: `include "bootstrap.php";` (registers the autoloader) → `opcache_compile_file()`, then `(new Application())->run(…)` fails. | E4: inclusions followed by `new`, method/static calls or non-builtin function calls are not reported. |
| SwitchContinuationInLoop (fix) | `continue` → `continue 2` also skipped `++$i;` after the switch (SuiteCRM's SQL parser position counter; 4). | No fix when statements follow the switch inside the loop. |
| StrlenInEmptyStringCheckContext (fix) | `!strlen($this->originalFileName)` → `=== ''` for an untyped `@var string` property without default (null on a new object: October deleted the file just saved; 6). | The cast is dropped only for a natively known string. |
| NullCoalescingOperatorCanBeUsed (fix) | `$type = $parts[1]; if (isset($mimeMap[$type])) $type = $mimeMap[$type];` → `$type = $mimeMap[$type] ?? $parts[1]` (looks up the old `$type`; SuiteCRM SVG uploads). | S3 requires a probe and value that do not read the target. |
| CascadeStringReplacement (fix) | `str_replace($bad, ' ', $n); str_replace('\'', '', $n)` → `[...array_values($bad), '\'']` / `[' ', '']` (EspoCRM sheet names lost the spaces). | Spread searches need aligned replacements (same literal everywhere, or the latest call's array pair). |
| SubStrUsedAsStrPos (fix) | `substr($n, 0, 2) == '00'` → `strpos(…) === 0` (`"0 " == "00"`; SuiteCRM `skype_formatted()`). | Loose comparisons fixed only against non-numeric literals (listed divergence updated). |
| ArgumentUnpackingCanBeUsed (fix) | `call_user_func_array('array_multisort', $p)` → `array_multisort(...$p)` (now really sorts). | No fix for callees with by-reference parameters. |
| SelfClassReferencing (fix) | `Model::make()` → `self::make()` changes late static binding (Grav `Uri`). | Static calls in non-final classes with indexed subclasses reported without a fix (39 fewer fixes). |
| RedundantElseClause, MkdirRaceCondition, CompactCanBeUsed, StringCaseManipulation, IncrementDecrementOperationEquivalent (fix; reproduced on own code, not seen in the corpus) | Code moved out of an unbraced `if` body / a hoisted `function` in the moved `else`; the thrown re-check `if` capturing an outer `else`; `fn () => compact('a')` (arrow functions capture only named variables); `strpos(mb_strtolower($s), 'é')` → `stripos()`; `$s = $s + 1` → `++$s` for strings and bools. | No fix / braced replacement / arrow functions need their own parameters / same `mb_`/byte family / string and bool operands skipped. |
| UnknownInspection | `SpellCheckingInspection`, PhpStorm naming/plural/condition inspections and unported Php Inspections names reported (33 of 33). | Added to the embedded list. |
| CallableParameterUseCaseInTypeContext | `@param unknown_type $x` / misspelt classes made every assignment incompatible; `$w = preg_replace(…, $w)` on a subject of unknown type typed `array|string` (SuiteCRM `SugarBean` after a `require`). | Parameters naming an unresolvable class skipped; replacements on unknown subjects not checked. 141 → 106. |
| OffsetOperations | Float keys (`$units[floor($v)]`, `$a[$i / 2]`; pChart, 19 errors). | Accepted where int is, like bool. |
| StaticInvocationViaThis | Magic `@method static` (Eloquent, facades) reported, fix to `static::` (a fresh instance). | Magic methods skipped. |
| MagicMethodsValidity | Swiftmailer's `call_user_func_array('Swift_Mime_SimpleMimeEntity::__construct', …)` not seen as a parent call (10 errors). | Callables ending in `::<method>` count. |
| PHPDoc (`internal/phpdoc`, table entry) | `@param $x The extension name` / `@return The item unserialized` read as class `\The` (about 60 SuiteCRM tags: OffsetOperations, IsEmptyFunctionUsage `!== null` advice, CallableParameterUseCaseInTypeContext). | Determiners and similar words followed by more text are a description. |
| MissingArrayInitialization, GetClassUsage, ClassConstantUsageCorrectness, HostnameSubstitution, MkdirRaceCondition, DisconnectedForeachInstruction, TypeUnsafeComparison | `$_SESSION['k'][] =` reported; `get_class($m)` after `$m->methodExists()`; `class_alias()` names; `!empty($_SERVER['SERVER_NAME']) ? $_SERVER['SERVER_NAME'] : …` reported twice; `while (!is_dir($d) && !@mkdir($d))` retry loops; `$progress && $progress([…])` ticks; `Money == Money` told to use `===`. | Superglobals skipped; an earlier method call counts as a null check; alias names skipped; fetches inside isset/empty ignored; retry loops skipped; callbacks are per-iteration; object pairs not reported. |

Deltas on the local corpora (HEAD 1efe7fd → this tree, default /
`--all`): corpus A `src/` 0 / −2 (CompactCanBeUsed in arrow functions),
corpus B 0 / 0, corpus C 0 / 0 findings, 478 ReturnTypeCanBeDeclared fixes
withheld (doc-typed getters and callee chains such as
`getRessourceRapport()->getType()`, whose documented non-null return can
be null).

Declined: UntrustedInclusion, MultipleReturnStatements,
UnSafeIsSetOverArray, AutoloadingIssues, NullPointerException,
LongInheritanceChain, NotOptimalIfConditions, ClassOverridesFieldOfSuperClass,
AlterInForeach, FixedTimeStartWith, NestedTernaryOperator, SlowArrayOperationsInLoop,
NonSecureUniqidUsage, ForgottenDebugOutput (`error_log`, debug views),
UnserializeExploits (also October's `tests/benchmarks/`, outside the spec's
test paths), UnnecessaryAssertion (`expects($this->any())`),
UsingInclusionReturnValue on config includes, PackedHashtableOptimization
`'0' =>` keys, UnnecessaryCasting `$_SERVER` typing, PropertyCanBeStatic,
SecurityAdvisories on runtime `filp/whoops` / `composer/composer`,
StaticInvocationViaThis and OffsetOperations following wrong project
PHPDoc — decided earlier or noise by design; SuspiciousSemicolon on an
intended empty loop; DisconnectedForeachInstruction on `fseek()` before a
connected `fread()` (needs effect analysis, as decided in round 6);
CallableParameterUseCaseInTypeContext null leaking through `$f = @realpath($f) ?: $f`
after a `preg_replace()` (needs failure markers to flow through
variables); SubStrUsedAsArrayAccess's message for `int|string` subjects
(no fix, never seen); MagicMethodsValidity "`__get` needs `__set`" on
read-only getters (spec, error severity upstream);
UnsupportedStringOffsetOperations one report per line (pChart, correct);
ReturnTypeCanBeDeclared fixes on public methods of a library (October's
rain) overridden outside the project (no project-type option);
NestedAssignmentsUsage evaluation order against an outer target
(`$l[$v.''] = $v = trim($v)`), UselessReturn on reference-bound variables,
IfReturnReturnSimplification for NAN/array operands, ObGetCleanCanBeUsed
with other output in the statement, CompactCanBeUsed on possibly undefined
variables and PhpUnitTests `assertTrue(!$x)` → `assertNotTrue($x)` (exact
only for bools) — reproduced on own code only, never seen; message
wording (multi-line expressions in IfReturnReturnSimplification /
ArrayPushMissUse messages, StringCaseManipulation and ArrayIsListCanBeUsed
message text, EmptyClass on traits); MissingIssetImplementation on
`\OAuthProvider` (stub gap, below). Engine-level causes found by the
review (since fixed, Engine section "Review round 7 engine causes"; the
remark on PHPDoc contradicted by every `return` stays a decision of earlier
rounds): *`.=` keeps the old type* — `$h =
file_get_contents($f); $h .= 'x';`: expected `string`, actual
`false|string`; `$n = null; $n .= 'x';`: expected `string`, actual
`null|string` (MagicMethodsValidity "`__toString` must return string",
SuiteCRM `DotListWizardMenu`); *`$this` in a trait* — `trait T { function
f() { $this; new static(); clone $this; } }`: expected `static` (or
unknown), actual `\T` (OffsetOperations on rain's `Sluggable`,
ReturnTypeCanBeDeclared, worked around in the rule); *negated
`is_numeric()` on a native string* — `function e(string $v) { if
(is_numeric($v)) { return $v; } $v; }`: expected `string`, actual unknown;
*destructuring in the casting typer* — `list($h, $m) = explode(':', $t);
if (!$h) { $h = 0; } (int) $h`: T-rules type expected `int|string` (or
unknown), actual `int` (UnnecessaryCasting removed `(int)` before
`setTime()` in EspoCRM `ClosestType`, a TypeError for `"9:30pm"`); *property
read after an unknown write* — `/** @var string */ protected $p; …
$this->p = isset($o['k']) ? $o['k'] : null; return $this->p;`: expected
`?unknown` (or `mixed|null`), actual `string` (ReturnTypeCanBeDeclared
`: ?string`; no longer fixable with the native-type policy); *`@return
void` over a body returning a value* — `/** @return void */ function
pair($t) { return [$t, 1]; }`: call typed `void`, expected the body's
`array` (4 SuspiciousAssignments errors on rain's `FieldParser`); *stub
gap* — `\OAuthProvider` declares none of the properties the extension
sets (`nonce`, `timestamp`, `consumer_key`, …; 2 MissingIssetImplementation
errors). PHPDoc contradicted by every `return` of an untyped body
(`@var EmailAddress[]` holding arrays, `@return self` returning arrays)
accounts for most remaining OffsetOperations errors on SuiteCRM; an engine
rule ignoring such docs would remove them (earlier rounds decided to
trust the doc).

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
the only exemption is the body of a one-statement `func main()` in package
main — each command's `os.Exit(run(...))` wrapper — which `tools/covercheck`
finds by parsing the source; a line-number list in the Makefile broke on
every edit above `main()`). The generators under `tools/` keep their logic in `run` functions
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
  `while`/`for` bodies, match arms, switch cases and after early-exit guards
  (see Narrowing completeness below).
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
  types `(T is X ? A : B)` give `A|B` (a return type is resolved per call,
  see Deeper inference below). A doc intersection refining an object
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
- **Calls:** first-class callables (`f(...)`) are `\Closure` (without a
  return type); named arguments bound in builtin return-type overrides.
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
- **Deeper inference (2026-10-08):** five additions, each bounded by the
  existing caps; findings measured old vs new engine built from the same
  tree (rules unchanged) on corpus A `src`, corpus A `vendor/symfony`,
  corpus B and corpus C (vendor excluded), default and `--all`: no change
  except +4 / +5 on Symfony (below). corpus A vendor `analyse --all` timing
  unchanged (alternating runs, within noise); `BenchmarkTypeOfCallables`
  +36 % allocations, the other infer benchmarks +1 %.
  - *Closures and callables* (`infer/callables.go`): a closure or arrow
    function is `\Closure` carrying its return type as the atom's generic
    argument (`types.CallableReturn`, like class generics it never changes
    the atom set): declared type, refined by an `@return` doc on the
    closure, else its body inferred like a function's (`void` reads as
    null; generators, recursion and nesting beyond `maxBodyDepth` give
    unknown). `callable(…): R` and `Closure(…): R` doc types keep R the same
    way (parameters are dropped; `DocString` writes `callable(): (R)`, which
    round-trips; a mixed R is not recorded). Calling a value — `$f()`,
    `(fn() => …)()`, `call_user_func[_array]($f)`, an object with
    `__invoke()` — gives that type (null/false callees ignored, any other
    member unknown). `array_map($cb, …)` is `R[]` (callback a closure, a
    typed callable, an invokable or a string naming a global function;
    a null, mixed or unknown R keeps the stub type); `array_filter()` keeps
    the element type (a sealed shape's values) and, without callback, drops
    null and false; `usort()` & co. keep the element type (only the shape
    goes, as for every by-reference argument). Method templates bind from a
    callback's return type (`@param callable(): T`, `Closure(int): U`), so
    `array_reduce()` and `Ds\*::map()` stubs bind too.
  - *Out parameters* (`infer/outparams.go`): a variable passed to a
    by-reference parameter documenting `@param-out T` (phpstan-/psalm-
    variants; `index.Param.Out`) or to a builtin output (`builtinOut`:
    preg_match `$matches` → `string[]`, `string[][]` for preg_match_all,
    plain `array` with flags; exec `$output` appends `string` to the
    elements it had; parse_str, count arguments, error codes/messages of
    fsockopen/stream_socket_*, …) is defined by the call. The definition
    kills earlier ones like an assignment when the call always runs with a
    statement directly in a block (expression statement, return, echo,
    condition of `if`/`while`/`switch`, not under the right operand of
    `&&`/`||`/`??` nor a ternary branch). Callees are resolved without type
    inference (definitions are collected before any variable is typed):
    named functions, `Name::`/`self::`/`parent::`/`static::`, `$this->`
    and `new Name`; other receivers keep the previous behaviour.
  - *Conditional return types* (`types/cond.go`, `infer/condreturn.go`):
    the index stores the conditional of `@phpstan-return` / `@psalm-return`
    / `@return` in a canonical form (`CondReturn`, names resolved; a
    template subject stands for the parameter documented as exactly that
    template). A call takes the branch its argument decides: literal
    targets (`'a'`, `1`) compare the literal argument, class-constant
    targets (`PDO::FETCH_ASSOC`) the constant or its literal value, type
    targets decide "then" only when the target is a plain type or class
    name list (`string`, `?Foo`; refined types such as `non-empty-string`
    or `array<int>` only decide "else") and the argument's atoms all belong
    to it, "else" when no value category overlaps; a missing argument uses
    the parameter's literal default. Nested branches resolve recursively;
    an undecided call keeps the previous type (the flattened union), or the
    union of the branches when only a `@phpstan-return` holds the
    conditional. The result must agree with the declared return type
    (`withDeclared`). Separators of a conditional need blanks around them,
    so `($x is ?int ? callable(): int : T)` parses (the old split broke
    on `?int` and on `callable(): int`). Parsing is capped like doc types
    (`MaxDocTypeLen`, 32 levels) and cached per Env; `FuzzFromDoc` also
    round-trips the canonical form. The stubs were regenerated: PDOStatement
    fetch/fetchAll and iterator_to_array gained conditionals, 47 stub
    callables gained signatures.
  - *Property writes* (`afterPropertyWrite`): a `$this->prop` read
    preceded, in an enclosing block, by `$this->prop = v;` / `??= v;` with
    a non-null `v` loses null unless something that may reset the property
    lies between (assignment, reference, any non-builtin call, mutation in
    a loop entered after the write: `nonEmptyBroken`). The existing
    `if ($this->p === null) { $this->p = …; }` guard now applies the same
    check (it ignored intervening calls). Non-null `$this->prop` reads
    (all narrowing): corpus A src 330 → 346, Symfony 245 → 253, corpus B
    132 → 148, corpus C 3 → 6.
  - *Element writes into null*: a variable whose type is an array and/or
    null gains an array of the written values for its null member
    (`$n = null; $n[] = 'x';` is `string[]`), and loses null when a write
    `$n[k] = v;` sits directly in a block enclosing the read after every
    reaching definition with no mutation in between (`writeDominates`).
    Other non-array members (strings: offset writes) are left alone.
  - *Bounds:* `brokenBy` (mutation scan between a fact and its use) now
    binary-searches the source-ordered mutation lists and examines at most
    `maxMutScan` = 512 mutations (beyond: broken); builtin-call
    resolution is cached per Env. Without it, 20k `$this->p = new X; $y =
    $this->p;` statements took 23 s; all probes of
    `TestCallableInferenceBounded` (20k repeated constructs, a 20k closure
    chain, conditionals at the length cap, 1,000 nested closures or arrow
    functions) now take under 0.25 s each.
  - *Symfony deltas* (all verified against the source): −1
    TypeUnsafeArraySearch (`array_search(strtolower($t), array_map('strtolower',
    …))`: string needle in a `string[]` haystack, spec E2); +1
    UnnecessaryCasting (`(string) $matches[$ofs]`, `$matches` from a
    `callable(string): list<string>` callback); +1 SubStrUsedAsArrayAccess
    (`substr($m[1], 0, 1)` after `preg_match(…, $m)`); +1
    ReturnTypeCanBeDeclared (`?int` from preg_replace's count); +2
    CallableParameterUseCaseInTypeContext (`string $s1` reassigned the
    `string[]` matches of preg_match_all); +1 OffsetOperations in `--all`
    (ParameterBag: `resolveValue()` documented `(TValue is scalar ?
    array|scalar : …)` used as a key — bool/float keys are what the rule
    reports, but `array` is listed only because the guard `if
    (!is_scalar($k) && !$k instanceof \Stringable) throw` is not
    understood: narrowing does not split a falsy `&&`).
- **Real-world review follow-up (2026-10-08):** engine root causes of
  false positives found on WordPress, Drupal, Laravel and Nextcloud.
  - *Version-specific builtin returns:* phpstorm-stubs give many builtins a
    `#[LanguageLevelTypeAware(["8.0" => "string"], default: "string|false")]`
    return type; the index kept only the newest. Functions and methods now
    carry the older versions' types (`RetVer`, `VerType{Until, Type}`), and
    `Index.Function`/`FindMethod`/`FunctionDecls` return a copy with
    `Return` resolved for the requested version (cached per declaration
    and version, so identity comparisons hold). `substr()` at 7.4 is
    `string|false`, so StrlenInEmptyStringCheckContext suggests
    `(string)$x !== ''` there. To keep the stricter types from making reads
    unknown: an element read of `T[]|false` gives T (as for `|null`), an
    offset read of `string|false|null` gives string, str_replace/preg_*
    overrides accept `true`/`false`/`null` subjects, and `explode()` with a
    non-empty literal separator is `string[]` (false only came from an
    empty separator). Stubs regenerated.
  - *Assignment as a condition:* `while ($job = $q->next())`,
    `if (!$x = f())` narrow the assigned variable like `$x` itself.
  - *Casting typer kills:* outside SpecOnly the T-rules only use the
    definitions reaching a read (Env's reaching definitions), so `$s = 0;
    $s = undeclared($s);` is unknown instead of `int` (UnnecessaryCasting);
    an inline `@var` on an assignment keeps that assignment's value, a
    standalone one contributes its type.
  - *Negated conditional assertions:* where a call is false its
    `-assert-if-true` assertions are applied negated (and `-if-false` ones
    where it is true): after `if (is_wp_error($m)) return;` `$m` is no
    `WP_Error`.
  - *`$_SERVER` ports* are strings (web SAPIs); `argc`, `REQUEST_TIME`
    (int), `REQUEST_TIME_FLOAT` (float), `argv` (array) stay. Divergence
    recorded in the UnnecessaryCasting spec.
  - *Absent literal keys:* reading a key a sealed shape does not list (and
    no write adds) is unknown instead of the other elements' type.
  - *`array_reduce()`*: the initial value's type (null when omitted) with
    the callback's return type; unknown when either is (the stub's
    `TCarry|null` added null and ignored the initial value).
  - *Class attributes* are indexed (`Class.Attrs`, FQNs; `HasAttr`), so
    MissingIssetImplementation sees `#[\AllowDynamicProperties]` declared
    in another file (Drupal: 26 → 0 findings).
  - *Duplicate function declarations:* a call to a function the project
    declares several times (WordPress `apply_filters()` in plugin.php and a
    no-op in noop.php) is the union of every declaration's type
    (`Index.FunctionDecls`), unknown beyond `maxFuncDecls` = 16.
  - *Not done:* `int * int` overflowing to float is still typed int; no
    finding acted on it in the review, and `int|float` would make every
    product non-int for the casting rules.
  - *Deltas* (old = HEAD 7e1530f, default / `--all`): corpus A src −1 / −1
    (CPUCITC on `$path = ($r = realpath($path)) ? $r : $path`, fixed by the
    assignment-condition narrowing), Symfony −1 / −1 (CPUCITC, `if ($c =
    \Closure::bind(…))`), corpus B 0 / −1 (OffsetOperations, `if ($key =
    array_search(…))`), corpus C +3 / +3: UnnecessaryCasting on `(string)
    $value` after `null !== $value` where `$value` is a preg_replace chain
    (now `string|null` through nullable subjects; true positive), and two
    ReturnTypeCanBeDeclared `: ?string` on preg_replace chains (true: the
    functions return preg_replace's `string|null`). WordPress (`--php
    7.4`): 13,613 → 13,480 default, 30,687 → 30,561 `--all`: −120
    ReturnTypeCanBeDeclared and most UnnecessaryCasting removals come from
    `apply_filters()` now unknown and `substr()`'s `false`; −5
    SuspiciousAssignments / −3 CPUCITC from `is_wp_error()` negation; 6
    StrlenInEmptyStringCheckContext messages now add the `(string)` cast
    and 138 TypeUnsafeComparison messages no longer call the other operand
    a non-numeric string (it may be false);
    +7 OffsetOperations (`--all`) are `substr()` results used as keys
    (may be false before 8.0); new UnnecessaryCasting are casts made
    redundant by negated assertions or killed definitions (e.g. `(int)
    $post_author` after `$post_author = $post_author->ID`). corpus A vendor
    timing unchanged (0.69 s both).
- **Narrowing completeness (2026-10-08):** conditions narrow in every
  place they decide control flow, measured old vs new engine built from
  one tree (rules unchanged).
  - *Negated compound conditions* (De Morgan): where `A && B` is false the
    type is the union of "A false" and "A true, then B false"; where
    `A || B` is true, the union of "A true" and "A false, then B true"
    (`applyCondB`; an emptied side is an impossible path and adds
    nothing). It applies wherever conditions do: guards, else branches,
    ternaries, loops. After `if (!is_scalar($k) && !$k instanceof
    \Stringable) throw …;` `$k` is `scalar|\Stringable` (Symfony
    ParameterBag). Both polarities of A are evaluated, so alternating
    nestings would cost 3^depth: one condition evaluation examines at most
    `maxCondSteps` = 1024 nodes, the rest narrows nothing (sound).
  - *Branch chains:* an elseif body sees the `if` and earlier elseif
    conditions as false, an else after elseifs all of them; `for` bodies
    use the last condition expression. `match` arms see the earlier arms
    as failed and their own values as `subject === value` (several values:
    their `||`), `default` all others failed; `switch` cases likewise with
    `==`, the empty cases falling into a case added to its values, nothing
    when a non-empty case may fall through (its last statement does not
    break/continue/return/throw/exit; goto does not count). The arm and
    case comparisons are synthetic `Binary` nodes handed to the ordinary
    condition code, so `match (true) { is_string($x) => … }` narrows like
    `if`, and `match ($x) { null => …, default => … }` like `=== null`.
    At most `maxBranchScan` = 256 earlier branches are applied (beyond:
    none, so a 10k-arm match stays linear).
  - *Comparisons:* `$x === 'a'` / `=== 1` / `=== K::C` / `=== Enum::A`
    narrow to the constant's type (members of the current type belonging
    to it, as for assertions); `true/false ===/== cond` narrows by cond
    (`!==` only when cond is boolean syntax: `match (true)` arms are
    compared strictly); `gettype($x)` (closed resources read "resource
    (closed)", so `!== 'resource'` keeps resources), `get_debug_type($x)`
    names and classes, `get_class($x)` / `$x::class` against `Foo::class`
    or a class-name literal (exact class: inequality removes nothing).
  - *Chains:* `$x?->m()` / `$x->p` truthy, `isset($x->p)`, `!empty($x->p)`,
    `$x?->p !== null` (or strictly equal to a non-null constant), `$x->p
    instanceof Foo` make `$x` non-null (an element chain `$x[…]`: not
    false either, non-empty).
  - *Type guards:* `is_a($x, Foo::class)` (with `allow_string` a string
    member stays), `is_subclass_of()` (strings allowed by default; false
    removes only strict subclasses), strict `in_array($x, list, true)`
    narrows to the list's element type (literal lists: the union of their
    values); `count($x) > 0` also drops null (count(null) is 0 or a
    TypeError); truthiness makes `bool` `true` and falsiness `false`
    (removing `true`). `instanceof` on `$this->prop` already narrowed.
  - *Fixed on the way:* a definition in a switch case ending with `break`
    (and the like) reached the later cases' reads (`dropOtherCases`; an
    enclosing loop's back edge still brings it); `bool` minus `true`/`false`
    became unknown when bool was the only member (now `false`/`true`, also
    in negative assertions).
  - *Cost:* `types.normalizeAtom` no longer lower-cases class names on
    every `Of`/`Has`/`Without` (case-folded map lookups on a stack buffer)
    and `Without` normalizes its arguments once: `BenchmarkTypeOfCallables`
    −10 % allocations; the new `BenchmarkTypeOfConditions` (compound
    guards, chains, match, switch) is ~+40 % time / +25 % allocations, the
    price of the extra narrowing; corpus A vendor `analyse --all` unchanged
    (8 alternating runs on a loaded machine: median real 1.6 s, user
    5.9 s for both engines). Probes
    (`TestNarrowingBounded`, n = 10k): 1,500-operand `&&`/`||` chains
    0.31 s, a 140-level alternating `&&`/`||` guard followed by 10k reads
    0.35 s, 10k elseifs 0.09 s, a 10k-arm `match ($x)` 0.04 s and `match
    (true)` 0.13 s, a 10k-case `switch (true)` 0.14 s.
  - *Deltas* (old = HEAD 6c987f9 engine, default / `--all`): corpus A src
    0 / 0, corpus C 0 / 0. Symfony −1 / 0: −1 UnnecessaryCasting
    (console Application: `(int) $exitCode` from `$e->getCode()` after
    `if ($e instanceof \Exception && !…) throw` guards — `$e` is now
    `\Exception|\Throwable` and `Exception::getCode()` is not int; true
    removal, PDOException codes are strings); `--all` also +1
    OffsetOperations (SymfonyQuestionHelper `$choices[$default]`: the
    failed `case null === $default:` removes null and the previous case's
    `$default = explode(…)` no longer reaches, leaving the documented
    `float` member, reported like other float keys) and the ParameterBag
    message now lists `\Stringable|float` instead of `array|float`
    (objects, Stringable included, are illegal array keys: correct).
    corpus B +1 / +1 UnnecessaryCasting (`(int) $workEditionId` where the
    id comes from `match (true) { $e instanceof WorkEdition => $e->id, …}`
    after a null check: true positive).
- **Review round 2 engine causes (2026-10-08):** the engine requests of
  the phpMyAdmin/Matomo/PrestaShop/Composer/PHPUnit/Doctrine review.
  - *Replacement functions:* `str_replace()`, `str_ireplace()`,
    `preg_replace()`, `preg_replace_callback[_array]()`, `preg_filter()`,
    `substr_replace()` follow the subject member by member (`replaceResult`):
    scalar and object members give string (converted), arrays give array,
    plus null for `preg_*`; mixed/iterable subjects keep the stub type. The
    T-rules typer does the same; in SpecOnly mode a subject the T-rules
    leave unknown falls back to the engine's type (the review's
    `string|array|null` came from such subjects). An int subject is now a
    string (was unknown in the T-rules).
  - *max()/min():* the union of the arguments (two or more), or the element
    type of a single `T[]` / sealed-shape argument (`|false` before 8.0
    unless non-empty); unknown/mixed members keep `mixed`. `abs(int)`,
    `round/floor/ceil` (float) and `intdiv` (int) were already right in
    both typers; the review's `int|float` came from untyped operands.
  - *Intersections:* `A&B` (PHPDoc, native, `(A&B)|null`) keeps its atoms
    and records them as one intersection (`types.Type.Intersection`, an
    8-byte pointer field). A union keeps it only while no member brings
    another class or a different intersection, `Without` drops it when a
    side goes; `String`/`DocString` print `\A&\B`, `(\A&\B)|null`, which
    round-trip. Method/property lookups (`memberClasses`: method calls,
    property fetches, assertions) use the first side declaring the member,
    and iteration the side that binds Traversable (Doctrine
    `matching()`: `AbstractLazyCollection<K, V>&Selectable<K, V>`).
  - *Property guards:* `if ($this->keyStore === false) { return; }` already
    narrowed the property (`string|true`), also on the Matomo file; the
    review's type came from an older engine. Regression test added.
  - *Per-key writes:* a read of literal key k on a `T[]` array only widens
    by writes to k, to computed keys, and appends when k is an integer
    (`writesForKey`): `$row[1] = explode(…)` no longer types `$row[0]`
    (phpMyAdmin `fetchRow()` loops; `$row['privs'] = …` / `$row['User']`).
  - *Pre-8.0 failure returns:* phpstorm-stubs give `hash()` & co. only their
    8.0 type (no LanguageLevelTypeAware map). `stubs.pre80Failure` adds the
    member below 8.0 for the builtins that returned false/null on invalid
    arguments and throw since 8.0: hash, hash_hmac, hash_pbkdf2, array_fill,
    chunk_split, wordwrap, substr_count, count_chars, str_word_count (false),
    array_chunk, array_rand (null). Audit of the others: hash_hkdf,
    mb_strlen, mb_str_split, password_hash, str_split, array_combine already
    vary in the stubs. A version-resolved declaration also adds the older
    declared members to its doc return (`docAt`), so `array_chunk()`'s doc
    `array` no longer hides null on 7.x.
  - *Index facts:* `Method.EmptyBody` (no statement, no promoted parameter;
    never for stubs) and `Builtin` on stub functions, methods and
    properties (set when the stubs load). Native-vs-PHPDoc provenance was
    already in the index (`Return`/`DocReturn`, `Type`/`DocType`); the
    engine side is `Env.Native()`: a twin Env (cached) whose types ignore
    user PHPDoc (param/property/return docs, inline `@var`, templates,
    conditional returns, assertions, `@param-out`, closure `@return`,
    index-time inferred types) but keep builtin stub docs; the T-rules typer
    over it drops them too. Wired: MagicMethodsValidity no longer asks for
    `parent::__construct()` (or `__clone`, `__destruct`) when the parent's
    body is empty; UnnecessaryCasting requires the target type without
    PHPDoc as well (listed EA divergences for the two MagicMethodsValidity
    cases; spec Divergences in both rules).
  - *Coverage without fuzz seeds:* `go test -skip '^Fuzz'` left 21
    statements of infer/types covered only by FuzzRules/FuzzFromDoc seeds
    (including the `brokenBy` ternary branch of shapes.go); unit tests now
    cover them (`TestInferEdgesUnitOnly`, `TestDocEdgesUnitOnly`): the six
    engine packages are at 100% with and without the fuzz seeds.
  - *Deltas* (old = HEAD 337502b, default / `--all`): corpus A src −2 / −2,
    Symfony −1 / 0, corpus B −55 / −55, corpus C −3 / −3. Removed: 63
    UnnecessaryCasting whose operand is typed only by PHPDoc (inline `@var`
    shapes, `@return int[]|null`, `@var` on ORM properties, doc array
    shapes, `callable(): list<string>` docs; sampled, all doc-only — by
    design). Added: 2 UnnecessaryCasting on Symfony `(int) max(0, $a - $b)`
    with native ints (true positives, max() now typed) and, in `--all`, 1
    OffsetOperations on `$units[$power]` with `$power = min(floor(…), 3)`
    (float key, reported like the others).
  - *Cost:* infer benchmarks within ±2 % allocations, +3 % bytes (the
    pointer field); corpus A vendor `analyse --all` user time unchanged
    (12 alternating runs, medians 5.9 s / 6.0 s on a loaded machine).
- **Review round 3 engine causes (2026-10-08):** the seven engine
  requests of the Magento/Joomla/CakePHP/Yii2/Laminas/Craft review.
  - *Loop back edges:* a back-edge definition sitting in a do-while
    condition, or in a for step/condition, is narrowed by the loop
    condition being true (`loopCondNarrow`): `do { $all[] = $e; } while ($e
    = $e->getPrevious());` gives `\Exception|\Throwable` in the body.
    A for step (`for (…; $r !== null; $r = next())`) written before the
    body in the source is treated as a back edge for body reads (it no
    longer suppresses the loop condition's narrowing).
  - *Static properties* (`self::$p`, `static::$p` (same key), `Name::$p`)
    have narrowing keys and narrow like `$this->p`: conditions, early-exit
    guards, `!empty()` non-emptiness (so `array_pop(self::$stack)` loses
    null), non-null after a write; any non-builtin call or write breaks
    the facts the way it does for `$this->p` (guards survive calls, as for
    `$this->p`).
  - *Inline `@var` class-name idiom:* a standalone `/** @var Widget $class
    */` (not attached to an assignment of `$class`) whose type is classes
    only, over a variable whose reaching definitions are all strings, is
    ignored: it describes the class a class-name string names (PhpStorm
    completion for `$class::widget()`), not the value. Other standalone
    hints keep overriding (`/** @var string $n */` after `$n = 1`), and
    declarations of otherwise undefined variables still apply.
  - *Exclusive branches:* definitions in another if/elseif/else body, the
    other ternary branch or another match arm no longer reach a read
    (`dropExclusive`), when no loop encloses the branching (a later
    iteration may run the other branch); so `if (is_int($c)) { $c = []; }
    else { … }` narrows `$c` to array in the else branch. Chains of more
    than `maxBranchScan` branches are not split (sound).
  - *Exits in reaching definitions:* per scope, each statement list's
    prefix up to its first always-leaving statement (return, throw, exit;
    break; continue; an if/else whose branches all leave) is an exit region
    (`exitRegions`, nested regions linked to their parent). A definition in
    a region reaches no read outside the region's statement list within
    the region's limit: the function for return/throw/exit, the loop or
    switch a break leaves (no region when a loop encloses that construct,
    or when a try encloses the statement: catch/finally resume). Code
    after the exit in the same list is dead and keeps its definitions.
    After continue, a definition reaches the reads of its loop only through
    the loop head: it joins the back-edge definitions unless a definition
    inside the loop precedes the read; a continue in a switch acts as
    break. Goto and multi-level break/continue end the scan of a list
    (no region). The dropped-case rule for switches (`dropOtherCases`) now
    also stands back inside loops. Probe `TestExitRegionsBounded` (20k
    guarded returns and if/else pairs: 0.2 s).
  - *Builtin docs never widen:* a builtin's version-resolved native return
    is authoritative; the stub doc only refines its members
    (`builtinMemberType`/`refining`: `string[]` for `array`, a class for
    `object`), in the engine and in the T-rules typer: `substr()` and
    `date()` at 8.x are `string`, `str_split()` `string[]`, `fgetcsv()`
    `array|false`.
  - *`@method` tags:* stored with `Method.Magic`; `Index.FindMethod`
    returns a real declaration from the class or any ancestor before a tag
    (Craft's `@method static ActiveQuery hasOne()` no longer hides Yii's
    instance `hasOne()`); `@method static` keeps the static flag.
  - *Deltas* (old = HEAD 2b80eec, default / `--all`): corpus A src −1 / −1,
    Symfony −1 / −1, corpus B 0 / −1, corpus C −1 / −2. Removed: 9
    CallableParameterUseCaseInTypeContext "type bool" on `substr()`/`date()`
    results at 8.x (true removals: no false since 8.0), 1 NullPointerException
    on corpus B `self::$inflector ??= …; $inflector = self::$inflector;`
    (static property non-null after the write; true removal). Added: 3
    UnnecessaryCasting in symfony/polyfill-mbstring `(string) substr(…)`
    (redundant at the configured 8.4 target, which the polyfill does not
    run on), 2 UnnecessaryCasting `(int) max(0, $a - $b)` with native ints
    (true positives), 1 ReturnTypeCanBeDeclared `: QueryBuilder` on corpus C
    `addSearch()` (every return is the QueryBuilder parameter or a
    `@return static` call on it; true positive).
  - *Cost:* corpus A vendor `analyse --all` unchanged (10 alternating runs,
    medians 1.17 s / 1.16 s real); infer benchmarks within 1 % allocations.
- **Review round 4 (2026-10-08):** one engine fix, in the T-rules typer:
  an enclosing `foreach` binding of unknown element type that reaches the
  read now makes the variable unknown, like an unknown reaching
  assignment (`foreach ($p as $v) { if ($c) { $v = 'a'; } (string) $v; }`
  was typed `string`; the binding is ignored when an unconditional
  reassignment hides it). Test `TestTRulesUnknownReachingDefinition`. The
  other engine causes of that review are listed with it (rule-level
  section) and left open.
- **Review round 4 engine causes (2026-10-08):** the nine engine requests
  of the Sylius/Shopware/API Platform/Mautic/Kimai/Akeneo review.
  - *`iterable` refined by docs:* a native `iterable` (parameter, property,
    return) takes its doc type when one is given, as `array` does
    (`@param Foo[]`, `@var array<string, X>`, `@return Collection<A>`).
  - *Partial generic arguments:* `@template` defaults (`@template TKey of
    array-key = array-key`) are parsed (`phpdoc.TemplateParam.Default`,
    `index.Template.Default`). With fewer arguments than templates, the
    arguments bind, in order, the templates without bound or default when
    their count matches (Shopware `Collection<LineItem>`: TElement); else
    the trailing templates when the leading ones are all bounded or
    defaulted; else a single argument to a Traversable class binds the
    value template (as before); unbound templates take their default.
  - *Boolean aliases:* `$isObject = is_object($r); if ($isObject)` narrows
    `$r`: a variable used as a condition whose only reaching definition is
    `$v = <boolean expression>` stands for that expression (`aliasCond`),
    unless the narrowed variable was written in between; one level only (an
    alias of an alias is not followed, `condBudget.inAlias`), plain
    variables only. Early-exit guards on aliases are found through an
    `aliasGuardKey` index entry for ifs whose condition is made of bare
    variables. Guard candidates are now cut by binary search per list
    before merging (a read with thousands of such ifs stays linear).
  - *Properties of variables:* `$v->p` and chains (`$param->var->name`,
    `$this->a->b`) have narrowing keys and narrow like `$this->p`
    (conditions, continue/return guards, non-null writes); a fact ends when
    the variable, or a property along the chain, is written in between
    (`chainBroken`, also for elements `$v->p['k']`); an assignment to `$v`
    resets its guards (and no longer counts as a write of `$v->p`).
  - *Inherited return types:* a method without declared or documented
    return type keeps the one of the nearest ancestor method it overrides
    (interfaces included, builtin ones through the stub rules); the body is
    inferred only when no ancestor declares one. `: mixed` declared on the
    method itself always won.
  - *Absent literal keys:* reading a key a sealed literal does not list is
    typed only by the reaching writes that may store into it (the same
    literal key, or a computed one, `mayWriteKey`); nested, appending (for
    integer keys) and destructuring writes, or writes to other keys only,
    leave it unknown, never typed as the other elements (`$rows['meta']`
    after `$rows['meta']['s'] = 1`).
  - *`str_replace()` & co. on an unknown subject* are unknown in the engine
    and in the T-rules typer outside SpecOnly (CallableParameterUseCaseInTypeContext's
    spec still types them `string|array`).
  - *Anonymous classes* are typed as the intersection of their parent and
    interfaces (`types.Intersect`; `\P&\I`, one class alone, `object`
    without any): members of either side are found; their own extra methods
    stay unknown. A synthetic indexed class was tried and dropped: its name
    leaked into rule output (ReturnTypeCanBeDeclared offered `:
    \class@anonymous…`) and anonymous-class scoping in rules.
  - *Promoted properties* take the constructor's `@param` type as their
    doc type; ClassMethodNameMatchesFieldName's local workaround was removed
    (fixture unchanged).
  - *Deltas* (old = HEAD 7c1e498 engine, rules unchanged; default /
    `--all`): corpus A src 0 / 0, corpus B 0 / 0, Symfony 0 / −3, corpus C
    −41 / −41. Removed: 3 OffsetOperations "index of type array" on
    `str_replace()` results of untyped values (Symfony config; true
    removals) and 43 ReturnTypeCanBeDeclared `: array` on corpus C
    repositories whose body returns Doctrine `getResult()`: inheriting
    `EntityManagerInterface::createQuery(): Query` types the body as
    `mixed`, so the rule no longer falls back on the `@return array` doc
    (the suggestions were doc-derived; the rule could prefer the doc for a
    `mixed` body — rule-side). Added: 2 ReturnTypeCanBeDeclared (`: Query`
    on `createQuery()`, `: ObjectRepository` on `getRepository()`, both
    per the declared contracts).
  - *Cost:* corpus A vendor `analyse --all` unchanged (10 alternating runs on
    a loaded machine, medians 1.7 s / 1.8 s real, 7.1 s user both); infer
    benchmarks within 1 % allocations except `BenchmarkTypeOfConditions`
    (+4 %). Probe `TestRound4Bounded` (20k variable-property guards and
    boolean aliases): 1.1 s.
- **Review round 5 engine causes (2026-10-08):** the nine engine requests
  of the TYPO3/MediaWiki/Moodle/phpBB/Flarum/Pimcore review.
  - *Dynamic writes:* `extract()` (any flags, conservatively),
    one-argument `parse_str()` and `$$name = …` may set any local. A read
    of a variable defined before one of them (it lies between a reaching
    definition and the read, or in a loop around the read) is unknown, in
    the engine and the T-rules typer (`noteDynamic`, `clobbered`; the
    positions are binary-searched).
  - *Possibly undefined variables:* a variable that only plain `$v = …`
    assignments define (no parameter, import, binding, reference, element
    write, out argument or inline @var; no dynamic construct or include in
    the scope), read where no unconditional assignment precedes it in an
    enclosing block and no statement or condition on the way always
    assigns it, also holds null (`if ($t) { $r = 'X'; } return $r;` is
    `string|null`). "Always assigns" is structural (`stmtAssigns`): an
    if/elseif/else whose branches all assign or return/throw/exit, a switch
    with default whose non-empty cases all assign, a try whose body and
    catches (or finally) assign, a do-while, assignments in conditions
    evaluated before the read; continue/break end a list without assigning.
    At most `maxAssignScan` = 64 statements are examined per read (beyond:
    taken as assigned, no null). ReturnTypeCanBeDeclared's local
    `rtdMaybeUndefined` is removed (its fixture passes; the extract/`$$`/
    parse_str cases are now unknown instead of `string`).
  - *Ancestor cap visible:* `Index.AncestorsComplete` reports a hierarchy
    cut by `MaxAncestors`; ReferencingObjects uses it instead of comparing
    the length.
  - *SpecOnly T-rules exits:* assignments followed by return/throw/exit/
    break (exit regions) no longer reach later reads in the SpecOnly typer
    either (`Env.exited`).
  - *Pseudo-types vs classes:* a pseudo-type name (`number`, `scalar`,
    `numeric`…) imported as a class (`use App\Number;`, case-insensitive)
    or declared in the current namespace is that class (`!name` resolver
    probe, `Env.className`); otherwise the pseudo-type.
  - *`class_alias()`:* calls with class constants or string literals are
    indexed (`FileSymbols.ClassAliases`); `Index.Class` resolves an alias
    to the original (at most `maxAliasHops` = 8 hops, cycles end there).
  - *Chains joining:* past an if/elseif(/else) chain, a variable is the
    union of what each path leaves: a branch's last assignment, or the
    incoming type narrowed by the conditions of that path (earlier false,
    own true; all false without else); always-leaving branches add nothing;
    another write in a branch keeps the type (`chainJoin`). `instanceof`
    now keeps the members of the tested type that are the class or its
    subtypes, adds the class only when another member may hold an instance
    of it (a parent, an interface, an unknown class, mixed/object/iterable/
    callable), and makes an impossible test (a string against a class) an
    unknown branch (`instanceOf`).
  - *`var_export()` / `print_r()`* return a string with a literal `true`
    `$return`, null / true without it or with `false` (both typers).
  - *`assert(cond)`* (the global function) narrows the following
    statements like an early-exit guard.
  - *Deltas* (old = HEAD 6294e7d, default / `--all`): no finding added or
    removed on corpus A src, Symfony, corpus B and corpus C; one message
    change: ParameterBag's OffsetOperations index type is now `float`
    instead of `\Stringable|float` (`resolveValue()` is documented
    `array|scalar`, which cannot be Stringable).
  - *Cost:* corpus A vendor `analyse --all` unchanged (10 alternating runs,
    medians 1.40 s / 1.47 s real, 6.5 s / 6.6 s user);
    `BenchmarkTypeOfConditions` +15 % allocations (chain joins), the other
    infer benchmarks within 1 %. Probe `TestPossiblyUndefinedBounded`
    (20k reads after 20k guarded ifs, 20k `extract()` calls): 0.4 s.
- **Review round 6 engine causes (2026-10-08):** the seven engine requests
  of the Firefly III/Monica/Bagisto/Dolibarr/Roundcube/FreshRSS review.
  - *Parameter-storing constructors:* `Method.StoresParams` / `Stores`: a
    project constructor whose every statement is `$this->prop = $param;`
    (plus promoted parameters) and the properties it sets. MagicMethodsValidity
    no longer asks for `parent::__construct()` when the parent constructor
    only stores its parameters and the child sets every one of those
    properties itself (assignment anywhere in its body, or promotion);
    spec Divergence, fixture `storing-parent.php`.
  - *CSPRNG wrappers across files:* `Function.CSPRNG` / `Method.CSPRNG`
    record a body calling random_bytes, openssl_random_pseudo_bytes or
    mcrypt_create_iv; EncryptionInitializationVectorRandomness accepts such
    a wrapper from another file (companion fixture, spec Divergence).
  - *thecodingmachine/safe:* a call to `Safe\X` with a builtin X is typed
    like X for these arguments, without false (Safe throws instead) and,
    for the preg_ functions or when the Safe declaration excludes it, null;
    a narrower documented Safe type wins; an undocumented body is ignored
    (both typers).
  - *include/require* count as dynamic writes: a later read of a local
    defined before them (or in a loop holding them) is unknown.
  - *Casting typer back edges:* outside SpecOnly, the T-rules typer adds the
    definitions later in a loop around the read (Env's back-edge
    definitions), so `foreach (…) { (int) $n; $n = (string) $r; }` sees
    `int|string`.
  - *Builtins and narrowing:* `pathinfo($f, PATHINFO_*)` is a string
    (array without flag), `gettimeofday()` an array (`true`: float);
    `$x[$k] ??= v` on an empty literal no write filled yet is v, and a
    computed-key read of an empty literal is typed by the values written
    since (non-empty literals stay unknown: a union of their values reads
    as noise); a type check that every known member fails (`!is_object($o)`
    on a `@var Foo`) leaves the branch unknown instead of the refuted type;
    preg_match()'s `$matches` is `string[]` without flags or with `0`
    (unmatched groups are `''`, trailing ones absent), `(string|null)[]`
    with `PREG_UNMATCHED_AS_NULL` (preg_match_all likewise, one level
    deeper), plain `array` with other flags.
  - *Definitions dominating a read inside a condition:* a definition in the
    left operand of && / || (or the condition of the if/while whose body
    holds the read) that always runs when the read runs — the && operands
    when the condition is true, the || ones when false, never a ?? right
    side, ternary branch or match arm — hides the definitions made before
    (`dropDominated`: `preg_match(…, $m, 0) && isset($t[$m[1]])` no longer
    unions an earlier preg_match_all() `$m`).
  - *Reaching definitions performance:* the kill of earlier definitions in
    the same block popped a suffix of the reaching list instead of
    filtering it per definition (quadratic: 400 definitions per variable
    read 20k times took 10 s, now 5 s, the rest being the union of up to
    400 reaching types per read); `TestReachingScales` checks that 4x the
    reads costs under 8x (measured ~4.7x).
  - *Deltas* (old = HEAD a031675, default / `--all`): corpus A src, corpus B
    and corpus C unchanged; Symfony −2 / −6: 4 OffsetOperations float-index
    reports in AbstractUnicodeString removed because a `require` of a data
    file sits in the loop around them (conservative: an include may reset
    locals), 2 UnnecessaryCasting `(int) max(…)` in AnsiUtils removed
    because the casting typer now sees the loop's later `+=` writes of
    unknown type (conservative).
  - *Cost:* corpus A vendor `analyse --all` real time unchanged (medians 1.18
    s both), user +2 %; infer benchmarks within 1 % allocations except
    `BenchmarkTypeOfConditions` −6 %.
- **Review round 7 engine causes (2026-10-08):** the seven engine requests
  of the SuiteCRM/EspoCRM/Kanboard/Grav/October/Koel review.
  - *Every assignment replaces the value:* an assignment to a plain
    variable with any operator (not by reference) hides the earlier
    definitions in its block like `=` does, so `$h = file_get_contents($f);
    $h .= 'x';` and `$n = null; $n .= 'x';` are `string`.
  - *Traits:* `$this`, `self`, `static`, `new static`, `clone $this` and
    `self::`/`static::` calls inside a trait are the using class, unknown
    when the trait is analysed (`Env.selfClass`; no object is an instance
    of a trait). StaticInvocationViaThis and DynamicInvocationViaScopeResolution
    keep resolving a trait's own methods through the trait (their only
    consumers of the old typing); ReturnTypeCanBeDeclared keeps its trait
    guard for doc types naming a trait (new fixture cases).
  - *Negated `is_numeric()`* removes only int and float: a string failing
    it is a non-numeric string (`if (is_numeric($v)) return; $v` is
    `string`).
  - *Casting typer and unmodelled definitions:* outside SpecOnly, a
    variable some reaching definition of which the T-rules do not type
    (destructuring, out arguments, catch, global/static, a foreach binding
    read after its loop, by-reference imports) is unknown instead of the
    type of the other definitions (`list($h, $m) = explode(…); if (!$h) $h
    = 0; (int) $h` keeps its cast).
  - *Unknown property writes:* a property read right after storing an
    unknown (or mixed) value in it (the existing property-write guard: same
    block, nothing that may reset it in between) is unknown, unless the
    property's native type is enforced on every receiver class; a doc type
    alone no longer survives the write.
  - *`@return void` over a returning body:* the index drops a documented
    `void` return when the function's own body has `return value;`, so the
    body's type (same file, or inferred at index time) applies.
  - *Stub gaps:* phpstorm-stubs does not declare the public properties
    pecl/oauth fills on `OAuthProvider` (consumer_key, nonce, timestamp,
    token, …); there is no extraction bug (the stub file has none, nor the
    reflection caches, oauth being PECL). `stubs.missingProps` declares them
    at load time (no regeneration needed), a table for any further
    extension class found incomplete.
  - *Deltas* (old = HEAD 2b89e41, default / `--all`): unchanged except
    corpus C +1 / +1 ReturnTypeCanBeDeclared `: ?string` on
    a Twig extension's `getAmount()` (the wrapped `amount()` returns `.=`
    results or `null`; true positive).
  - *Cost:* corpus A vendor `analyse --all` unchanged (medians 1.13 s real,
    6.47 s / 6.50 s user); infer benchmarks unchanged.
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
- File reading (2026-10-08): profiles showed ~45% of CPU in `open`/`stat`
  (kernel contention with one reader per core, every file read twice).
  `analyse` reuses the bytes the index pass read (`BuildIndexKeep` →
  `RunSources`), at most 6 files are read at once, and `safeio` sizes its
  buffer from the file. Vendor corpus `--all`: 0.78–0.93 s → 0.73–0.78 s,
  system CPU 2–4 s → 1.2 s; findings byte-identical. `safeio` still checks
  the path (`stat`) before opening: an intermediate version opened first
  and checked the descriptor, which a security review flagged — opening a
  device can have side effects (tape rewind, serial line toggling, a
  terminal becoming the controlling tty). The open adds `O_NONBLOCK` and
  `O_NOCTTY` for a path swapped between check and open, and the type is
  re-checked on the descriptor; with reads capped, the extra `stat` costs
  nothing measurable.
- ReturnTypeCanBeDeclared doc trust (2026-10-08, user decision): a
  suggestion that rests on the `@return` tag alone (a returned value of
  unknown or `mixed` type) is reported without a quick-fix. Restoring the
  doc fallback for `mixed` bodies (Doctrine `getResult()`) brought back
  43 corpus C suggestions but added ~14.5k on Moodle (mostly a generated
  API client); applying them blindly would turn any wrong doc into a
  TypeError. Same policy as UnnecessaryCasting's PHPDoc-only types.
  Extended in review round 7 (2026-10-08, user decision to keep it):
  the fix is offered only when the native typing (PHPDoc
  ignored, `Env.Native`) gives the same type for every returned value, so
  a callee's `@return`, a `@param` or a property's `@var` no longer
  suffices either (Kanboard `getProgress(): int` over a `round(…, 1)`
  documented `@return integer`). Reports are unchanged; corpus C keeps
  1,091 of 1,569 fixes.
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
- **Condition narrowing caps (2026-10-08):** negating `A && B` narrows A
  in both polarities, so an alternating `&&`/`||` nesting was exponential;
  one condition evaluation examines at most `infer.maxCondSteps` = 1024
  nodes, and elseif chains, match arms and switch cases apply at most
  `infer.maxBranchScan` = 256 earlier branches (beyond: no narrowing from
  them). Probes in `TestNarrowingBounded` (Engine: Narrowing
  completeness).
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
