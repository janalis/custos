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
| ForeachInvariants | `foreach-invariants.php` | upstream's IDE formatter re-spaces an untouched inner for header (custos copies the body verbatim); additionally a limit variable reused by several loops is resolved from the assignment reaching each loop (one extra report) |
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
| Test-file detection | Spec definition everywhere: path ends with `Test.php`, `Spec.php`, `.phpt` or contains `/Fixtures/` (case-sensitive) — `ctx.IsTestFile()` / `util.IsTestPath`; rules add their own test-class checks per spec. |
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
  `/** @var T $x */` overrides the next assignment; element writes
  (`$a[k] = v`) widen the element type.
- **Narrowing:** `is_*()`, `instanceof`, null/true/false comparisons, isset,
  truthiness — in ternary branches, if/elseif/else bodies, `&&`/`||` operands,
  `while` bodies and after early-exit guards.
- **PHPDoc:** `@template` names map to `mixed`; `@phpstan-type`/`@psalm-type`
  aliases expand; `@phpstan-import-type` → `mixed`; nested generics parse. Conditional
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
