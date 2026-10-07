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
  `/** @var T $x */` overrides the next assignment; a foreach binding hides
  earlier definitions inside the loop body; element writes (`$a[k] = v`)
  widen the element type of the reads they reach (below).
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
  generic arguments, see Generics); `@phpstan-type`/`@psalm-type`
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
