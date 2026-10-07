---
id: PhpUnitTests
group: PHPUnit
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# PhpUnitTests

## Summary
A grab-bag of PHPUnit hygiene checks. On test method docblocks it validates the
targets of `@covers`, `@depends` and `@dataProvider`, flags a redundant `@test`
and (optionally) unnamed datasets. On assertion and mock-builder calls it
points at generic assertions that have a dedicated, more expressive PHPUnit
counterpart (`assertTrue(!$x)` → `assertNotTrue($x)`,
`assertSame(1, count($l))` → `assertCount(1, $l)`, `->will($this->returnValue(…))`
→ `->willReturn(…)`, …) and rewrites them.

The rule has two independent halves:
- **Part A** — docblock tags on methods (D1–D12).
- **Part B** — method calls whose name starts with `assert`, or is `expects`
  or `will` (D13–D30).

Neither half checks that the class extends a PHPUnit test case: every method
of every class-like and every matching method call is examined.

## Detection

### Part A — method docblock tags

- **D1** Visit every method declaration that (a) is declared inside a
  class-like (class, abstract class, interface, trait), (b) has a name, and
  (c) has a docblock (`/** … */`). Top-level functions and closures are never
  examined.
- **D2** Iterate over every tag of that docblock (in source order); each tag
  is handled independently, so one method may be reported several times. Tag
  names are compared case-sensitively and exactly: `@dataProvider`,
  `@depends`, `@covers`, `@test` (`@coversNothing`, `@coversDefaultClass`,
  `@testWith` … are other tags and ignored).
- **D3 — "annotation position".** A tag only counts when it starts a docblock
  line: the token immediately before the tag (skipping at most one whitespace
  run) is the docblock opener `/**` or a leading `*` of a docblock line. A tag
  mentioned in the middle of prose (`/** see @covers below */`,
  ` * text @covers X`) is ignored. `@test` (D11) also requires this.
- **D4 — tag value.** For `@dataProvider`, `@depends`, `@covers` the tag's
  first element after the tag name must be a reference-like word
  (identifier, qualified name, `Name::member`, `::member`, optionally followed
  by `()`); otherwise the tag is ignored (e.g. empty tag).

Reference resolution used below: a reference is the word from D4. It may
contain a class part and a member part separated by `::`.
- Class part: resolved like a class name in code at that position (current
  namespace, `use` imports/aliases, leading `\` = fully qualified), against
  project classes and built-in stubs (classes, interfaces, traits).
- Member part: a method looked up in the resolved class **including
  inherited methods** (parent classes, also stub classes such as
  `\SplObjectStorage`, `\ArrayIterator`; traits; interfaces).
- A bare word without `::` in `@depends`/`@dataProvider` is a method name
  looked up in the method's containing class (including inherited methods).
- `::name` (empty class part) in `@covers` names a global function, resolved
  against project functions and built-in stubs.
- A trailing `()` after the member name is ignored for resolution.

#### `@dataProvider`
- **D5** The reference's **member** (the last component; for a bare word, the
  word itself; for `C::m`, `m` in `C`) must resolve to a method. If there is no
  resolvable reference at all, or it resolves to nothing / to something other
  than a method → report (error). A `C::m` reference whose class `C` exists
  but has no `m` is reported (no fallback to the class).
- **D6** (only when `SUGGEST_TO_USE_NAMED_DATASETS` is on) The provider
  resolved in D5 is a non-abstract method whose body's **last statement** is a
  `return` of an array literal (`[...]` or `array(...)`). Look at the array's
  **first element**:
  - no elements (`[]`) → fine;
  - a `key => value` element whose key is a string literal (any quoting) →
    fine;
  - anything else (a value-only element, or a non-string-literal key such as
    an integer, constant or concatenation) → report on the **test** method
    (info). Only the first element is checked.
  Providers whose last statement is not such a `return` (yield-based,
  returning a variable or a call, empty body) are not reported.

#### `@depends`
- **D7** Resolve the member like D5. If it does not resolve to a method →
  report (error).
- **D8** If it resolves to a method whose name does **not** start with `test`
  (case-sensitive prefix) **and** whose docblock has no `@test` tag anywhere
  (any position; missing docblock counts as none) → report (error).
  A method named `test…` or carrying `@test` is a valid dependency.

#### `@covers`
- **D9** Let `T` be the reference text as written (e.g. `Book::settle`,
  `\App\Cart`, `::strlen`, `Cart::<public>`, `Cart::pay()`).
  A *callable target* is required when `T` contains `::` and does **not**
  contain `::<`; otherwise a *class target* is required.
- **D10** Resolution outcome:
  - `C::m` where `m` resolves to a method of `C` (incl. inherited) → class
    and callable both satisfied;
  - `C::m` where `C` resolves but `m` does not → only class satisfied;
    callable satisfied only if `T` ends with `::`;
  - `::f` (or a bare name) resolving to a **global function** → callable
    satisfied, class **not** satisfied;
  - a bare name / `C::<selector>` / `C::` where `C` resolves to a class-like →
    class satisfied (and callable too if `T` ends with `::`);
  - nothing resolves → nothing satisfied.
  Report (error) when a callable target is required and the callable is not
  satisfied, or a class target is required and the class is not satisfied.
  So: `@covers Foo` needs class `Foo`; `@covers Foo::bar` needs method `bar`
  in `Foo` (or its ancestors); `@covers ::fn` needs global function `fn`;
  `@covers Foo::<private>` / `@covers Foo::<!public>` need only `Foo`.
- **D10a** `::m` (empty class part, not `::<…>`) when the class docblock
  carries an annotation-position `@coversDefaultClass C`: first resolve
  `C::m` per D10; if that satisfies the callable target, no report.
  Otherwise fall back to D10 (global function `m`).

#### `@test`
- **D11** Tag `@test` in annotation position (D3) on a method whose name
  starts with `test` (case-sensitive) → report the redundant tag (info,
  deprecated-style highlight). A `@test` on a method not named `test…` is fine.

- **D12** Report placement: D5–D10 highlight the **name identifier of the
  method carrying the docblock** (the test method), not the tag. D11
  highlights only the tag-name token `@test`.

### Part B — assertion and mock calls

- **D13** Visit every method call: instance (`$x->m()`, `$x?->m()`) and static
  (`self::m()`, `static::m()`, `parent::m()`, `Foo::m()`). The receiver is
  never checked. Let `M` be the method name, compared case-insensitively as
  PHP compares method names (`AssertTrue`, `$this->Exactly(1)`, `->Will(...)`
  match; every PHPUnit method name in D13–D30 is compared this way, and the
  enclosing-function names of D28a/D28b and E8 too). Fix texts use the
  declared spelling of the suggested method.
  - `M` starts with `assert` and `M` ≠ `assert` → run the assertion
    strategies (D14) in order, **stopping at the first one that reports**;
  - `M` = `expects` → D29 (if `PROMOTE_MOCKING_ONCE`);
  - `M` = `will` → D30 (if `PROMOTE_MOCKING_WILL_RETURN`).
- **D14** Strategy order (the order matters because several can match the
  same call):
  1. inverted boolean (D15) — always;
  2. boolean of comparison (D16) — always;
  3. strict equality (D17) — only if `SUGGEST_TO_USE_ASSERTSAME`;
  then, only if `PROMOTE_PHPUNIT_API`:
  4. empty (D18); 5. constant (D19); 6. internal type (D20);
  7. instanceof (D21, D22); 8. resource exists (D23); 9. count (D24);
  10. contains (D25); 11. regex (D26, D27); 12. file equals (D28a);
  13. string equals file (D28b).

Conventions for D15–D28: `args` are the assertion call's arguments (`n` of
them); "strip parens" means removing any number of wrapping parentheses; a
"function call `f`" means a plain function call (not a method call, not
`new`) resolving to the global function `f`: names compare
case-insensitively like PHP function names (`Count(...)`, `\IS_INT(...)`
match), and a call resolving to a same-named namespaced function (declared
in the current namespace, imported with `use function`, or written with a
non-global qualifier) does not match. "Number" means an
integer/float literal, optionally with a unary minus.

- **D15 Inverted boolean.** `M` ∈ {`assertTrue`, `assertFalse`}, `n ≥ 1`,
  `args[0]` with parens stripped is a logical-not `!X`. `X'` = `X` with parens
  stripped. Suggest `assertNotTrue` (for `assertTrue`) / `assertNotFalse`
  (for `assertFalse`). Other unary operators do not match.
- **D16 Boolean of comparison.** `M` ∈ {`assertTrue`, `assertNotTrue`,
  `assertFalse`, `assertNotFalse`}, `n ≥ 1`, `args[0]` with parens stripped
  is a binary `L op R` with `op` ∈ {`==`, `!=`, `<>`, `===`, `!==`}.
  - method negates = `M` ∈ {`assertFalse`, `assertNotTrue`};
  - operator negates = `op` ∈ {`!=`, `<>`, `!==`};
  - strict = `op` ∈ {`===`, `!==`}.
  Suggest `assert` + (`Not` if exactly one of the two negates) + (`Same` if
  strict else `Equals`). `L` and `R` keep their original text (not stripped).
- **D17 Strict equality** (option). `M` ∈ {`assertEquals` → `assertSame`,
  `assertNotEquals` → `assertNotSame`}, `n ≥ 2`, and both `args[0]` and
  `args[1]` have a **fully known** inferred type (no unknown/mixed part) in
  which no member is a class/interface type and no member is `array`
  (i.e. only `int`, `float`, `string`, `bool`/`true`/`false`, `null`). Literals
  qualify (`3`, `'a'`, `1.5`, `true`, `null`); typed scalar parameters
  qualify; untyped variables, arrays (`[]`, `array $p`) and objects don't.
- **D18 Empty.** `M` ∈ {`assertTrue`, `assertNotFalse` → `assertEmpty`;
  `assertFalse`, `assertNotTrue` → `assertNotEmpty`}, `n ≥ 1`, `args[0]` is
  directly (no parens) an `empty(X)` construct with exactly one argument.
- **D19 Constant.** `M` ∈ {`assertSame` → `assert<C>`, `assertNotSame` →
  `assertNot<C>`}, `n ≥ 2`. Scan **all** arguments in order (including a 3rd
  message argument) for the first that is a bare constant `null`, `true` or
  `false` (case-insensitive, e.g. `NULL`, `True`); `<C>` is `Null`, `True` or
  `False`. The *other* value is the first argument (in order) that is not that
  constant node. No match → strategy does not apply.
- **D20 Internal type.** `M` ∈ {`assertTrue`, `assertNotFalse`} (positive) or
  {`assertFalse`, `assertNotTrue`} (negative), `n ≥ 1`, `args[0]` is a
  function call with **at least one** argument to one of:
  `is_array`→`array`, `is_bool`→`bool`, `is_float`→`float`, `is_int`→`int`,
  `is_null`→`null`, `is_numeric`→`numeric`, `is_object`→`object`,
  `is_resource`→`resource`, `is_string`→`string`, `is_scalar`→`scalar`,
  `is_callable`→`callable`, `is_iterable`→`iterable`.
  (Aliases such as `is_integer`, `is_long`, `is_double` are not mapped.)
  - PHPUnit version **< 8.0**: suggest `assertInternalType` (positive) /
    `assertNotInternalType` (negative) with the type name.
  - PHPUnit version **≥ 8.0**: suggest `assertIs<Type>` (positive) /
    `assertIsNot<Type>` (negative), `<Type>` = type name with first letter
    upper-cased: `assertIsArray`, `assertIsNotScalar`, `assertIsNull` (sic,
    see Divergences), …
- **D21 instanceof.** `M` ∈ {`assertTrue`, `assertNotFalse` →
  `assertInstanceOf`; `assertFalse`, `assertNotTrue` → `assertNotInstanceOf`},
  `n ≥ 1`, `args[0]` is directly (no parens) a binary `S instanceof K`.
  Class expression `CD`:
  - `K` is a class name reference (incl. `self`/`static`/`parent`):
    - language level ≥ 5.5 → `K` as written + `::class` (`\Foo\Bar::class`,
      `Bar::class`);
    - level < 5.5 → a single-quoted string of `K`'s **fully qualified** name
      (leading `\`, resolved through namespace/imports) with every `\`
      doubled: `'\\Foo\\Bar'`;
  - otherwise (`$cls`, expression) → `K`'s text as is.
- **D22 get_class comparison.** `M` ∈ {`assertSame`, `assertEquals` →
  `assertInstanceOf`; `assertNotSame`, `assertNotEquals` →
  `assertNotInstanceOf`}, `n ≥ 2`. The literal is `args[0]` if it is a string
  literal, else `args[1]`; it must be a string literal without interpolation
  whose raw contents (between the quotes, escapes not processed) are longer
  than 3 characters. The other of the first two arguments must be a function
  call `get_class` with exactly one argument `O`. `FQ` = `\` + contents with
  each `\\` pair collapsed to one `\`. `CD` = `FQ::class` at level ≥ 5.5, else
  `'` + `FQ` with each `\` doubled + `'`.
- **D23 Resource exists.** `M` ∈ {`assertTrue`, `assertNotFalse`} (positive)
  or {`assertFalse`, `assertNotTrue`} (negative), `n ≥ 1`, `args[0]` is a
  function call with exactly one argument to `file_exists` (→ `File`) or
  `is_dir` (→ `Directory`). Suggest `assert<K>Exists` (positive) /
  a negative name that depends on the PHPUnit version:
  - PHPUnit version **≥ 9.1**: `assert<K>DoesNotExist`
    (`assertFileDoesNotExist`, `assertDirectoryDoesNotExist`);
  - PHPUnit version **< 9.1**: `assert<K>NotExists`
    (`assertFileNotExists`, `assertDirectoryNotExists`), the only names
    those versions provide.
  The positive names `assertFileExists` / `assertDirectoryExists` are the same
  at every version.
- **D24 Count.** `M` ∈ {`assertSame`, `assertEquals` → `assertCount`;
  `assertNotSame`, `assertNotEquals` → `assertNotCount`}, `n ≥ 2`, `args[1]`
  (the actual value — position matters, `args[0]` is not checked) is a
  function call `count` with exactly one argument `A`.
- **D25 Contains.** Only when PHPUnit version **< 9.0**. `M` ∈
  {`assertTrue`, `assertNotFalse` → `assertContains`; `assertFalse`,
  `assertNotTrue` → `assertNotContains`}, `n ≥ 1`, `args[0]` is a function
  call `in_array` with **at least two** arguments `N`, `H` (a third `strict`
  argument is dropped by the fix).
- **D26 Regex, numeric form.** `M` ∈ {`assertSame`, `assertEquals` (expected
  `1`), `assertNotSame`, `assertNotEquals` (expected `0`)}, `n ≥ 2`, `args[0]`
  is a number, `args[1]` is a function call `preg_match` with exactly two
  arguments `P`, `S`. If `args[0]`'s text equals the expected text (`1` resp.
  `0`, exact text) → the positive regex name, otherwise → the negative one.
  So `assertSame(1, …)`/`assertNotSame(0, …)` → positive;
  `assertSame(0, …)`/`assertNotSame(1, …)` → negative.
  Regex names depend on the PHPUnit version: from **9.1** positive
  `assertMatchesRegularExpression`, negative
  `assertDoesNotMatchRegularExpression`; below 9.1 positive `assertRegExp`,
  negative `assertNotRegExp`.
- **D27 Regex, comparison form.** `M` ∈ {`assertTrue` → positive regex
  name, `assertFalse` → negative regex name (as in D26)} (not the `Not…`
  variants), `n ≥ 1`,
  `args[0]` is directly (no parens) a binary `L > R` where `R` is a number
  whose text is exactly `0` and `L` is a function call `preg_match` with
  exactly two arguments `P`, `S`. (`>=`, `!=`, `0 < …` do not match.)
- **D28a File equals.** `M` ∈ {`assertSame`, `assertEquals`,
  `assertStringEqualsFile`}, `n ≥ 2`, and the nearest enclosing
  function/method/closure is **not** named `assertFileEquals` (any case). Let
  `E0`/`E1` be the single argument of `args[0]`/`args[1]` when that argument
  is a function call `file_get_contents` with exactly one argument (else
  none).
  - `M` = `assertStringEqualsFile` and `E1` exists → suggest
    `assertFileEquals` with expected `args[0]` (text as is) and actual `E1`;
  - otherwise, `E0` and `E1` both exist → suggest `assertFileEquals` with
    `E0`, `E1`;
  - otherwise the strategy does not apply.
- **D28b String equals file.** `M` ∈ {`assertSame`, `assertEquals`}, `n ≥ 2`,
  `args[0]` is a function call `file_get_contents` with exactly one argument
  `F`, and the nearest enclosing function/method/closure is named neither
  `assertFileEquals` nor `assertStringEqualsFile`. Suggest
  `assertStringEqualsFile`. (`file_get_contents` in `args[1]` alone does not
  match.)
- **D29 expects(exactly(1)).** Outer call `M` = `expects` with exactly one
  argument which is a method call (instance or static, any receiver) named
  `exactly` (any case) with exactly one argument that is a number whose
  text is exactly `1`. Suggest `once`.
- **D30 will(returnX(...)).** Outer call `M` = `will` with exactly one
  argument which is a method call (any receiver) named one of
  `returnValue` → `willReturn`, `returnValueMap` → `willReturnMap`,
  `returnCallback` → `willReturnCallback`, `returnArgument` →
  `willReturnArgument`, with exactly one argument `V`.

## Exceptions (no report)
- **E1** Methods without docblock, functions, closures (Part A).
- **E2** Tags not in annotation position (D3); tags with no reference value.
- **E3** `@dataProvider`/`@depends` resolving to a valid target (D5/D8);
  `@covers` satisfied per D10.
- **E4** Named-dataset check: option off; abstract provider; last statement
  not `return <array literal>`; empty array; first key a string literal.
- **E5** `@test` on a method whose name does not start with `test`.
- **E6** Method calls named exactly `assert`, or not starting with `assert`
  (other than `expects`/`will`); name comparison is case-sensitive
  (`AssertTrue` is ignored).
- **E7** Shapes not listed in D15–D30, e.g. `assertSame(count($l), 2)`
  (count in first position), `assertSame('x', file_get_contents($f))`,
  `assertTrue(in_array($n))` (one argument), `count($l, COUNT_RECURSIVE)`,
  `preg_match` with a 3rd argument, `expects($this->exactly(2))`,
  `will($this->returnSelf())`, `will($this->onConsecutiveCalls(1, 2))`.
- **E8** D28a inside a function named `assertFileEquals`; D28b inside a
  function named `assertFileEquals` or `assertStringEqualsFile` (a custom
  helper implementing the assertion itself).
- **E9** D25 when the configured PHPUnit version is ≥ 9.0.
- **E10** D17 when either operand's type is unknown, or contains a class or
  `array`.
- **E11** Disabled option groups (see Options): with `PROMOTE_PHPUNIT_API`
  off only D15, D16 (and D17 if enabled) run.

## Report
- Severity: rule default `info` for everything **except** the unresolved /
  inappropriate targets D5, D7, D8, D10, which are `error`. D6 is `info`.
  D11 is `info` (upstream marks it with a strike-through "deprecated" style;
  custos may tag it as deprecated).
- Range:
  - D5–D10: the name identifier of the method carrying the tag.
  - D11: the `@test` tag-name token only (e.g. the 5 characters `@test`).
  - D15–D28, D30: the **whole method call expression**, from the start of its
    receiver to the closing `)` — e.g. `$this->assertTrue($v > 1)`,
    `self::assertSame(1, count($l))`,
    `$stub->method('fetch')->will($this->returnValue(3))` (the entire chain
    that forms the receiver of `will` is included). The trailing `;` is not.
  - D29: the inner call only, e.g. `$this->exactly(1)`.
- Messages (our wording):
  - D5: `The @dataProvider target cannot be resolved to a method.`
  - D6: `Give the provider's datasets string keys.`
  - D7/D8: `The @depends target is missing or is not a test.`
  - D9/D10: `The @covers target '{T}' cannot be resolved.`
  - D11: `Remove '@test': the method name already marks it as a test.`
  - D15–D16, D18–D28: `Use '{assertion}()' instead.` (for D20 below 8.0:
    `Use '{assertion}('{type}', …)' instead.`)
  - D17: `Loose comparison; use '{assertion}()' instead.`
  - D29: `Use '->once()' instead.`
  - D30: `Use '->{method}()' instead.`

## Fix
All Part B fixes keep the receiver and the `->`/`::` operator, **rename the
called method** to the suggested name and **replace the whole argument list**
`( … )` with a freshly built one: the listed argument texts copied verbatim
(original source text of each node), joined by `, ` (comma + one space), no
other whitespace or comments kept from the old list. Text outside the call
(e.g. a space before `;`) is untouched.

"msg" below is the original `args[1]` or `args[2]` text as indicated, added
only when the call has that argument. Several strategies build an argument
list whose length is derived from `n` (given as `len`); with superfluous
arguments after the message, the slots left after the listed ones are
filled, in order, with the original arguments that follow the message
(`assertTrue(empty($x), 'm', 3)` → `assertEmpty($x, 'm', 3)`).

- **F1 (D15)** `assertNotTrue|assertNotFalse(X')` when `n ≠ 2`;
  `assertNotTrue|assertNotFalse(X', args[1])` when `n = 2`. (With `n ≥ 3`
  everything after `X'` is dropped.)
  `$this->assertTrue(!($a && $b), 'm')` → `$this->assertNotTrue($a && $b, 'm')`.
- **F2 (D16)** `<suggested>(L, R)` when `n ≠ 2`; `<suggested>(L, R, args[1])`
  when `n = 2`. Operand order kept.
  `$this->assertFalse($q !== 0)` → `$this->assertSame($q, 0)`.
- **F3 (D17)** Only the name changes; all `n` argument texts are re-emitted
  joined by `, `: `assertEquals(4, 2+2, 'x')` → `assertSame(4, 2+2, 'x')`.
- **F4 (D18)** `(X[, args[1]])`, `len = n`.
  `assertFalse(empty($rows), 'm')` → `assertNotEmpty($rows, 'm')`.
- **F5 (D19)** `(other[, args[2]])`, `len = n − 1`.
  `assertNotSame(FALSE, $ok)` → `assertNotFalse($ok)`;
  `assertSame($id, null, 'm')` → `assertNull($id, 'm')`.
- **F6 (D20)** below 8.0: `('<type>', A0[, args[1]])` where `A0` is the
  `is_*` call's first argument, `len = n + 1`
  (`assertTrue(is_int($n))` → `assertInternalType('int', $n)`);
  8.0+: `(A0[, args[1]])`, `len = n`
  (`assertNotTrue(is_string($s), 'm')` → `assertIsNotString($s, 'm')`).
- **F7 (D21)** `(CD, S[, args[1]])`, `len = n + 1`.
  `assertNotTrue($e instanceof \Shop\Cart)` →
  `assertNotInstanceOf(\Shop\Cart::class, $e)`.
- **F8 (D22)** `(CD, O[, args[2]])`, `len = n`.
  `assertEquals('Shop\\Cart', get_class($e))` →
  `assertInstanceOf(\Shop\Cart::class, $e)`.
- **F9 (D23)** `(A[, args[1]])`, `len = n`.
  `assertNotTrue(is_dir($cache))` → `assertDirectoryDoesNotExist($cache)` at
  PHPUnit ≥ 9.1, `assertDirectoryNotExists($cache)` below.
- **F10 (D24)** `(args[0], A[, args[2]])`, `len = n`.
  `assertNotEquals(3, count($bag), 'm')` → `assertNotCount(3, $bag, 'm')`.
- **F11 (D25)** `(N, H[, args[1]])`, `len = n + 1`.
  `assertNotFalse(in_array($k, $keys, true))` → `assertContains($k, $keys)`.
- **F12 (D26)** `(P, S[, args[2]])`, `len = n`. **(D27)** `(P, S[, args[1]])`,
  `len = n + 1`. `assertFalse(preg_match('/^a/', $w) > 0, 'm')` →
  `assertNotRegExp('/^a/', $w, 'm')` below PHPUnit 9.1,
  `assertDoesNotMatchRegularExpression('/^a/', $w, 'm')` from 9.1.
- **F13 (D28a)** `(args[0], E1[, args[2]])` for `assertStringEqualsFile`, or
  `(E0, E1[, args[2]])`; `len = n`.
  `assertEquals(file_get_contents($a), file_get_contents($b))` →
  `assertFileEquals($a, $b)`.
- **F14 (D28b)** `(F, args[1][, args[2]])`, `len = n`.
  `assertSame(file_get_contents($p), $body)` →
  `assertStringEqualsFile($p, $body)`.
- **F15 (D29)** The inner call is renamed to `once` and its argument list
  becomes `()`: `$this->exactly(1)` → `$this->once()`.
- **F16 (D30)** The outer call is renamed and its argument list becomes
  `(V)`: `->will($this->returnCallback($fn))` → `->willReturnCallback($fn)`.
- **F17 (D11)** Delete the `@test` tag (the tag element, including any value
  text on the same line belonging to it). Before that, if the tag is
  immediately preceded by whitespace, delete that whitespace; then, if what
  now immediately precedes is a leading `*` of the docblock line, delete it
  too. The docblock opener `/**` is never deleted.
  `/** @test */` → `/** */`; a line ` * @test` inside a multi-line docblock
  becomes an (otherwise blank) line.
- D5–D10 and D6 have no fix.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `PHP_UNIT_VERSION` | enum | `PHPUNIT80` | Target PHPUnit version, one of `PHPUNIT70`, `PHPUNIT71`, `PHPUNIT72`, `PHPUNIT73`, `PHPUNIT74`, `PHPUNIT75`, `PHPUNIT80`, `PHPUNIT81`, `PHPUNIT82`, `PHPUNIT83`, `PHPUNIT84`, `PHPUNIT85`, `PHPUNIT90`, `PHPUNIT91`, `PHPUNIT92`, `PHPUNIT93`, `PHPUNIT94`, `PHPUNIT95` (ordered; comparisons are by this order). Below 8.0 D20 suggests `assertInternalType`/`assertNotInternalType`, from 8.0 `assertIs*`/`assertIsNot*`. D25 only runs below 9.0. D23 uses the `NotExists` negative names below 9.1 and `DoesNotExist` from 9.1. D26/D27 use `assertRegExp`/`assertNotRegExp` below 9.1 and `assertMatchesRegularExpression`/`assertDoesNotMatchRegularExpression` from 9.1. No other effect on detection or fixes. Conformance passes it as `PhpUnitVersion.PHPUNIT75` style values; strip the prefix. Unset → inferred from the indexed PHPUnit (see Divergences), else `PHPUNIT80`. |
| `SUGGEST_TO_USE_ASSERTSAME` | bool | `false` | Enables D17. |
| `SUGGEST_TO_USE_NAMED_DATASETS` | bool | `false` | Enables D6. |
| `PROMOTE_PHPUNIT_API` | bool | `true` | Enables D18–D28. |
| `PROMOTE_MOCKING_ONCE` | bool | `true` | Enables D29. |
| `PROMOTE_MOCKING_WILL_RETURN` | bool | `true` | Enables D30. |

## PHP versions
- No rule-level gating; PHPUnit version handled by `PHP_UNIT_VERSION`.
- D21/D22 fix output depends on the language level: ≥ 5.5 uses `::class`,
  below uses a quoted FQN with doubled backslashes. Upstream fixtures run at
  the IDE test default (between 5.6 and 7.0) → `::class` form.

## Examples

Docblock checks (`SUGGEST_TO_USE_NAMED_DATASETS` on):

```php
<?php
namespace Shop\Billing {
    class Ledger extends \SplObjectStorage {
        public function settle() {}
    }
}

namespace {
    use Shop\Billing\Ledger as Book;

    class LedgerTest {
        /** @covers Book::settle */
        public function testSettles() {}

        /** @covers Book::attach() */
        public function testAttachesInherited() {}

        /**
         * Rolls back.
         * @covers Book::rollback
         */
        public function <error descr="The @covers target 'Book::rollback' cannot be resolved.">testRollback</error>() {}

        /** @covers \Shop\Billing\Journal */
        public function <error descr="The @covers target '\Shop\Billing\Journal' cannot be resolved.">testJournal</error>() {}

        /** @covers ::strlen */
        public function testMeasures() {}

        /** @covers Book::<protected> */
        public function testInternals() {}

        /** Unlike @covers Nowhere, this is prose. */
        public function audit_notes() {}

        /** @test */
        public function seedsLedger() {}

        /** <weak_warning descr="Remove '@test': the method name already marks it as a test.">@test</weak_warning> */
        public function testBalances() {}

        /** @depends seedsLedger */
        public function testAfterSeed() {}

        /** @depends testBalances */
        public function testAfterBalance() {}

        /** @depends helper */
        public function <error descr="The @depends target is missing or is not a test.">testNeedsHelper</error>() {}

        /** @depends \LedgerTest::vanished */
        public function <error descr="The @depends target is missing or is not a test.">testNeedsVanished</error>() {}

        /** @dataProvider amounts */
        public function testAmounts($v) {}

        /** @dataProvider plainAmounts */
        public function <weak_warning descr="Give the provider's datasets string keys.">testPlainAmounts</weak_warning>($v) {}

        /** @dataProvider noAmounts */
        public function testNoAmounts($v) {}

        /** @dataProvider phantom */
        public function <error descr="The @dataProvider target cannot be resolved to a method.">testPhantom</error>($v) {}

        public function helper() {}
        public static function amounts() { return ['ten' => [10], 'zero' => [0]]; }
        public static function plainAmounts() { return [[10], [0]]; }
        public static function noAmounts() { return array(); }
    }
}
```

Assertions (defaults: PHPUnit 8.0, `PROMOTE_PHPUNIT_API` on; language level ≥ 5.5):

```php
<?php
class CartTest
{
    public function testThings($rows, $bag, $cart, $path, $body, $slug)
    {
        <weak_warning descr="Use 'assertNotTrue()' instead.">$this->assertTrue(!($cart->open && $rows))</weak_warning>;
        <weak_warning descr="Use 'assertSame()' instead.">self::assertFalse($cart->size !== 4, 'size')</weak_warning>;
        <weak_warning descr="Use 'assertNotEquals()' instead.">$this->assertTrue($slug <> 'home')</weak_warning>;
        <weak_warning descr="Use 'assertNotEmpty()' instead.">$this->assertFalse(empty($rows))</weak_warning>;
        <weak_warning descr="Use 'assertNull()' instead.">$this->assertSame($cart->coupon, NULL, 'coupon')</weak_warning>;
        <weak_warning descr="Use 'assertIsNotString()' instead.">$this->assertNotTrue(is_string($slug))</weak_warning>;
        <weak_warning descr="Use 'assertInstanceOf()' instead.">$this->assertNotFalse($cart instanceof \Shop\Cart, 'type')</weak_warning>;
        <weak_warning descr="Use 'assertNotInstanceOf()' instead.">$this->assertNotEquals('Shop\\Order', get_class($cart))</weak_warning>;
        <weak_warning descr="Use 'assertDirectoryNotExists()' instead.">$this->assertNotTrue(is_dir($path))</weak_warning>;
        <weak_warning descr="Use 'assertNotCount()' instead.">$this->assertNotEquals(3, count($bag), 'bag')</weak_warning>;
        <weak_warning descr="Use 'assertContains()' instead.">$this->assertNotFalse(in_array($slug, $rows, true))</weak_warning>;
        <weak_warning descr="Use 'assertNotRegExp()' instead.">$this->assertFalse(preg_match('/^[a-z]+$/', $slug) > 0)</weak_warning>;
        <weak_warning descr="Use 'assertRegExp()' instead.">$this->assertNotEquals(0, preg_match('/\d/', $slug), 'digits')</weak_warning>;
        <weak_warning descr="Use 'assertFileEquals()' instead.">$this->assertEquals(file_get_contents($path), file_get_contents($body))</weak_warning>;
        <weak_warning descr="Use 'assertStringEqualsFile()' instead.">$this->assertSame(file_get_contents($path), $body)</weak_warning>;

        $this->assertSame(count($bag), 3);
        $this->assertEquals($body, file_get_contents($path));
        $this->assertTrue(in_array($slug));
        $this->AssertTrue(!$rows);
        $this->assert(!$rows);
    }

    public function assertStringEqualsFile($file, $text)
    {
        $this->assertSame(file_get_contents($file), $text);
    }

    public function testMocks()
    {
        $stub = $this->createMock(\ArrayAccess::class);
        $stub->expects(<weak_warning descr="Use '->once()' instead.">$this->exactly(1)</weak_warning>)->method('offsetGet');
        <weak_warning descr="Use '->willReturnCallback()' instead.">$stub->method('offsetGet')->will($this->returnCallback('strtoupper'))</weak_warning>;
        $stub->expects($this->exactly(2))->method('offsetSet');
        $stub->method('offsetExists')->will($this->returnSelf());
    }
}
```

```php
<?php
class CartTest
{
    public function testThings($rows, $bag, $cart, $path, $body, $slug)
    {
        $this->assertNotTrue($cart->open && $rows);
        self::assertSame($cart->size, 4, 'size');
        $this->assertNotEquals($slug, 'home');
        $this->assertNotEmpty($rows);
        $this->assertNull($cart->coupon, 'coupon');
        $this->assertIsNotString($slug);
        $this->assertInstanceOf(\Shop\Cart::class, $cart, 'type');
        $this->assertNotInstanceOf(\Shop\Order::class, $cart);
        $this->assertDirectoryNotExists($path);
        $this->assertNotCount(3, $bag, 'bag');
        $this->assertContains($slug, $rows);
        $this->assertNotRegExp('/^[a-z]+$/', $slug);
        $this->assertRegExp('/\d/', $slug, 'digits');
        $this->assertFileEquals($path, $body);
        $this->assertStringEqualsFile($path, $body);

        $this->assertSame(count($bag), 3);
        $this->assertEquals($body, file_get_contents($path));
        $this->assertTrue(in_array($slug));
        $this->AssertTrue(!$rows);
        $this->assert(!$rows);
    }

    public function assertStringEqualsFile($file, $text)
    {
        $this->assertSame(file_get_contents($file), $text);
    }

    public function testMocks()
    {
        $stub = $this->createMock(\ArrayAccess::class);
        $stub->expects($this->once())->method('offsetGet');
        $stub->method('offsetGet')->willReturnCallback('strtoupper');
        $stub->expects($this->exactly(2))->method('offsetSet');
        $stub->method('offsetExists')->will($this->returnSelf());
    }
}
```

`PHP_UNIT_VERSION = PHPUNIT75`, `SUGGEST_TO_USE_ASSERTSAME` on:

```php
<?php
class LegacyTest
{
    public function testTypes(int $qty, $raw, array $list)
    {
        <weak_warning descr="Use 'assertInternalType('int', …)' instead.">$this->assertTrue(is_int($qty))</weak_warning>;
        <weak_warning descr="Use 'assertNotInternalType('iterable', …)' instead.">$this->assertFalse(is_iterable($raw), 'iter')</weak_warning>;
        <weak_warning descr="Loose comparison; use 'assertSame()' instead.">$this->assertEquals(12, $qty)</weak_warning>;
        <weak_warning descr="Loose comparison; use 'assertNotSame()' instead.">$this->assertNotEquals('7', 3.5, 'odd')</weak_warning>;
        <weak_warning descr="Use 'assertNotContains()' instead.">$this->assertNotTrue(in_array($qty, $list))</weak_warning>;
        $this->assertEquals($list, []);
        $this->assertEquals($raw, 12);
    }
}
```

```php
<?php
class LegacyTest
{
    public function testTypes(int $qty, $raw, array $list)
    {
        $this->assertInternalType('int', $qty);
        $this->assertNotInternalType('iterable', $raw, 'iter');
        $this->assertSame(12, $qty);
        $this->assertNotSame('7', 3.5, 'odd');
        $this->assertNotContains($qty, $list);
        $this->assertEquals($list, []);
        $this->assertEquals($raw, 12);
    }
}
```

## Divergences
- **Name matching in Part B (custos diverges from upstream).** Upstream
  compares the assertion and mock method names case-sensitively (so
  `$this->AssertTrue(!$x)` is skipped) and matches the inner function calls
  by their last name segment without resolution (so a namespace's own
  `count()` is treated as the builtin). custos compares method names
  case-insensitively and requires inner function calls to resolve to the
  global function.
- **`assertIsNull` / `assertIsNotNull`** (D20, PHPUnit ≥ 8.0 with `is_null`):
  upstream builds these names, which do not exist in PHPUnit.
  Recommendation: suggest `assertNull` / `assertNotNull` instead. No upstream
  fixture covers `is_null`.
- **Regex names from 9.1 — custos diverges from upstream** (D26/D27).
  Upstream suggests `assertRegExp`/`assertNotRegExp` at every PHPUnit
  version, although they are deprecated in 9.1 (and removed in 10), so the
  "fix" introduces deprecation warnings. custos suggests
  `assertMatchesRegularExpression`/`assertDoesNotMatchRegularExpression` from
  9.1 and the old names below.
- **Resource exists below 9.1** (D23; custos diverges from upstream).
  Upstream always suggests `assertFileDoesNotExist` /
  `assertDirectoryDoesNotExist`, but PHPUnit only added those names in 9.1;
  below that the fix would call a method that does not exist and break the
  test. custos suggests `assertFileNotExists` / `assertDirectoryNotExists`
  below 9.1 and the `DoesNotExist` names from 9.1. The upstream fixture runs
  at the 8.0 default and expects the 9.1 names, so it is a listed
  conformance divergence. (Upstream's message at ≥ 9.1 also misspells the
  name; custos' message uses the suggested name.)
- **Superfluous arguments (custos diverges from upstream).** Upstream fills
  the argument slots that follow the message with the literal `null`, so
  `assertTrue(empty($x), 'm', 3)` becomes `assertEmpty($x, 'm', null)` and
  the original extra arguments are lost. custos copies the remaining
  original arguments into those slots (Fix section).
- **D19 scans the message argument**: `assertSame($a, $b, true)` becomes
  `assertTrue($a, true)`. Recommendation: only consider `args[0]`/`args[1]`
  as the constant.
- **D22 leading backslash**: a literal that already starts with `\`
  (`'\\Shop\\Cart'`) produces `\\Shop\Cart::class` (double leading backslash,
  invalid). Recommendation: strip a leading `\` from the contents before
  prefixing.
- **D17 "primitive" types**: upstream accepts any fully-known type set
  without class types or `array`, which also lets through pseudo-types like
  `mixed`, `callable`, `iterable`, `resource`, `int[]`. Recommendation: accept
  only `int`, `float`, `string`, `bool`, `true`, `false`, `null`.
- **Function-name matching (custos diverges from upstream).** Upstream
  matches `count`, `is_int`, `in_array`, `get_class`, … case-sensitively, so
  `Count($x)` or `IS_INT($n)` are not recognised although PHP function names
  are case-insensitive. custos compares them case-insensitively.
- **`@covers ::method` with `@coversDefaultClass` (custos diverges)**:
  PHPUnit resolves `::method` against the class-level `@coversDefaultClass`;
  upstream treats `::name` as a global function and reports when none exists
  (false positive on `@covers ::__construct` under a default class). custos
  resolves `::m` as `C::m` first (D10a).
- **PHPUnit version when `PHP_UNIT_VERSION` is unset (custos diverges)**:
  upstream assumes `PHPUNIT80`, so a PHPUnit 10+ project is told to use
  `assertRegExp()`, which no longer exists. When the option is not
  configured, custos infers the version from the indexed PHPUnit (vendor
  included): `PHPUnit\Framework\Assert` declares `assertMatchesRegularExpression` →
  `PHPUNIT91`; else it declares `assertIsInt` but not `assertInternalType` →
  `PHPUNIT90`; else `assertIsInt` → `PHPUNIT80`; else `PHPUNIT70`. When
  PHPUnit is not indexed either, `PHPUNIT80`.
- **`@depends clone x` / `@depends shallowClone x`**: the modifier word is
  taken as the reference and reported. Recommendation: skip `clone` /
  `shallowClone` and resolve the following word.
- **No test-case check**: any class method docblock and any `assert*` call
  on any receiver is examined. Kept as upstream: the upstream fixtures wrap
  their code in plain classes, and restricting to test classes would drop all
  conformance coverage for a rare false positive.
- **PHPUnit 10+ attributes** (`#[DataProvider]`, `#[Depends]`, `#[CoversClass]`,
  `#[Test]`) are not examined. Recommendation: out of scope, as upstream.
