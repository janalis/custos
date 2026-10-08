---
id: UnnecessaryAssertion
group: PHPUnit
kind: semantic
needs: [names, types, hierarchy]
php: { min: "", max: "" }
---

# UnnecessaryAssertion

## Summary

Some PHPUnit assertions verify something the language already guarantees:
asserting `null`/emptiness on the result of a `void` function, or the class
or type of a value returned by a function with a declared return type. Also,
`->expects($this->any())` on a mock asserts nothing and can be dropped.

## Detection

Visit every method call (`->`, `?->` or `::`, any receiver). The method name,
compared case-insensitively as PHP does (`AssertNull`, `EXPECTS`, `$this->ANY()`
match; this applies to every method name in this spec), selects the check:
names starting with `assert` go to Part A, every other name to Part B. The assertion call itself is never
resolved and there is no test-context check.

### Part A — assertion on a typed return value (PHP ≥ 7.0)

- **D1** The method name is exactly one of the following; `p` is the 0-based
  position of the checked argument and `T` the expected type:

  | name                 | `p` | `T`                              |
  |----------------------|-----|----------------------------------|
  | `assertNull`         | 0   | `void`                           |
  | `assertEmpty`        | 0   | `void`                           |
  | `assertInstanceOf`   | 1   | class of argument 0 (see D6)     |
  | `assertInternalType` | 1   | type named by argument 0 (D6)    |

  The call must have at least `p + 1` arguments.
- **D2** Compute the possible values of argument `p` (rules V1–V7 below).
  There must be **exactly one** value `F`.
- **D3** `F` is a function call or method call (`f()`, `$o->m()`,
  `Cls::m()`, `$this->m()`; not `new`, not a closure invocation of a
  variable) that resolves to a function or method declaration.
- **D4** That declaration has an explicit return type declaration (any
  type). Docblock-only `@return` does not qualify.
- **D5** The inferred type of the call expression `F` (declared return type
  combined with any `@return` docblock types) consists of **exactly one**
  type and contains nothing unknown. So `?Foo` (= `Foo|null`), unions, or a
  declared `object` plus a docblock `@return Foo` (two types) disqualify.
- **D6** Match against `T`:
  - `assertNull` / `assertEmpty`: report when that single type is `void`
    (compared case-insensitively, a leading `\` ignored).
  - `assertInternalType`: argument 0 must be a single- or double-quoted
    string literal (decoded) naming a PHPUnit internal type, and the single
    declared type must always satisfy it (type names compared
    case-insensitively; `X[]` counts as `array`, `\X` as a class):

    | type name                      | satisfied by                         |
    |--------------------------------|--------------------------------------|
    | `array`                        | `array`, `X[]`                       |
    | `bool`, `boolean`              | `bool`, `true`, `false`              |
    | `float`, `double`, `real`      | `float`                              |
    | `int`, `integer`               | `int`                                |
    | `numeric`                      | `int`, `float`                       |
    | `string`                       | `string`                             |
    | `scalar`                       | `int`, `float`, `string`, `bool`, `true`, `false` |
    | `null`                         | `null`, `void`                       |
    | `object`                       | `object`, any class                  |
    | `resource`                     | nothing (see Divergences)            |
    | `callable`                     | `callable`, `\Closure`               |
    | `iterable`                     | `iterable`, `array`, `X[]`           |

    Any other name, a non-literal argument 0, or a declared type that does
    not satisfy the name (`assertInternalType('string', f())` with
    `f(): object`) → no report.
  - `assertInstanceOf`: argument 0 must be `X::class` where `X` resolves to a
    class-like; report when the single type, as a fully-qualified class name,
    equals the FQN of `X` (`\Ns\Box` vs `\Ns\Box`; compare
    case-insensitively). A parent/child relationship is **not** enough —
    only the identical class. If argument 0 is not a resolvable `X::class`
    (string class name, variable, unknown class), no report (see
    Divergences).
- Report finding **TYPED** on the whole assertion call.

Possible-values rules for D2 (recursive; each node visited at most once;
parentheses stripped first):

- **V1** Ternary `c ? a : b` / `c ?: b`: union of the values of both result
  branches.
- **V2** `a ?? b`: union of the values of `a` and `b`.
- **V3** Variable `$v`: take the innermost enclosing function, method,
  closure or arrow function (none → no values). Add the values of the default
  of a same-named parameter if any; then for every plain `=` assignment (by
  value or by reference; not compound, not list destructuring) anywhere in
  that function's body (nested closures included) whose left side is `$v`,
  add the values of the assigned expression (for chains `$v = $w = e` the
  innermost `e`). Position relative to the assertion is irrelevant.
  **Unstable variable** (custos refinement, see Divergences): if `$v` is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown* → no report.
- **V4** Property fetch: resolve the property; add the values of its default
  (unless the default's text ends with the property name); then apply the V3
  assignment scan for that fetch in the enclosing function and in the
  declaring class's constructor.
- **V5** Class constant fetch: values of the constant's initializer.
- **V6** Global constant (not `true`/`false`/`null`): its `define()`/`const`
  value as-is; unresolvable → no values.
- **V7** Anything else (calls included): the expression itself.

### Part B — `expects($this->any())`

- **D7** The method name is exactly `expects`, with exactly one argument.
- **D8** That argument is a method call (`->`, `?->` or `::`, any receiver,
  any arguments) named exactly `any`. A plain function `any()` does not
  match. Report finding **ANY** on that argument.
- No PHP-level gating for Part B.

## Exceptions (no report)

- **E1** Part A when the PHP language level is below 7.0.
- **E2** Other assertions (`assertNotNull`, `assertIsString`,
  `assertSame`, `assertTrue`, …), and the listed ones with too few arguments.
- **E3** Arguments with zero or several possible values (variable assigned
  twice, ternary of two calls, parameter without default, top-level variable).
- **E4** Values that are not resolvable calls, calls to functions without a
  declared return type, calls whose inferred type is nullable, a union, or
  unknown.
- **E5** `assertNull`/`assertEmpty` on a non-`void` call;
  `assertInternalType` whose type name is not a literal, is not a PHPUnit
  internal type, or is not guaranteed by the declared type;
  `assertInstanceOf(A::class, …)` where the call's single type is a
  different class (even a subclass) or a non-class type.
- **E6** `expects($this->once())`, `expects($this->never())`,
  `expects($matcher)`, `expects(any())`.

## Report

- Range:
  - TYPED: the entire assertion method-call expression, from the start of
    its receiver (`$this`, `self`, `static`, a variable…) to its closing `)`;
    the statement's `;` is excluded.
  - ANY: the `expects` argument expression exactly (e.g. `$this->any()`).
- Severity: info (weak warning) for both.
- Messages (our wording):
  - TYPED: `The declared return type already guarantees this; the assertion can go.`
  - ANY: `expects(any()) verifies nothing; drop the expects() call.`

## Fix

- TYPED: no fix.
- **F1** ANY: replace the whole `expects(...)` call expression (receiver,
  arrow/`::`, name and argument list) with the source text of its receiver
  expression. Whatever follows the `expects(...)` call (e.g.
  `->method('x')->willReturn(1)`) stays as it is, including the whitespace
  before the next `->`. Whitespace/comments that were between the receiver
  and `->expects` disappear with the call.
  - `$double->expects($this->any())->method('save')->willReturn(true);`
    → `$double->method('save')->willReturn(true);`
  - multi-line
    `$double\n    ->expects(self::any())\n    ->method('save');`
    → `$double\n    ->method('save');`
  - a bare statement `$double->expects($this->any());` → `$double;`

## Options

None.

## PHP versions

- Part A requires PHP language level ≥ 7.0 (return type declarations).
  The EA fixture runs at 7.1 and uses `void` (7.1) and `object` (7.2)
  return types; custos parses them regardless of level, only the 7.0 gate
  applies.
- Part B: no gating.

## Examples

At PHP 7.4:

```php
<?php

namespace Shop;

class Cart {}

abstract class CartTest
{
    abstract protected function flush(): void;
    abstract protected function cart(): Cart;
    abstract protected function maybeCart(): ?Cart;
    abstract protected function anything();

    public function testTyped(bool $flag)
    {
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertNull($this->flush())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">self::assertEmpty($this->flush())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, $this->cart())</weak_warning>;
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInternalType('object', $this->cart())</weak_warning>;
        $this->assertInternalType('object', $this->cart());
        $this->assertInternalType('array', $this->cart());
        $this->assertInstanceOf(\ArrayObject::class, $this->cart());
        $this->assertInstanceOf(Cart::class, $this->maybeCart());
        $this->assertNull($this->cart());
        $this->assertNull($this->anything());

        $made = $this->cart();
        <weak_warning descr="The declared return type already guarantees this; the assertion can go.">$this->assertInstanceOf(Cart::class, $made)</weak_warning>;

        $either = $flag ? $this->cart() : $this->flush();
        $this->assertNull($either);
    }

    public function testMatcher()
    {
        $double = $this->createMock(Cart::class);
        $double->expects(<weak_warning descr="expects(any()) verifies nothing; drop the expects() call.">$this->any()</weak_warning>)->method('total')->willReturn(9);
        $double
            ->expects(<weak_warning descr="expects(any()) verifies nothing; drop the expects() call.">self::any()</weak_warning>)
            ->method('count');
        $double->expects($this->once())->method('total');
    }
}
```

```php
<?php

namespace Shop;

class Cart {}

abstract class CartTest
{
    abstract protected function flush(): void;
    abstract protected function cart(): Cart;
    abstract protected function maybeCart(): ?Cart;
    abstract protected function anything();

    public function testTyped(bool $flag)
    {
        $this->assertNull($this->flush());
        self::assertEmpty($this->flush());
        $this->assertInstanceOf(Cart::class, $this->cart());
        $this->assertInternalType('object', $this->cart());
        $this->assertInternalType('array', $this->cart());
        $this->assertInstanceOf(\ArrayObject::class, $this->cart());
        $this->assertInstanceOf(Cart::class, $this->maybeCart());
        $this->assertNull($this->cart());
        $this->assertNull($this->anything());

        $made = $this->cart();
        $this->assertInstanceOf(Cart::class, $made);

        $either = $flag ? $this->cart() : $this->flush();
        $this->assertNull($either);
    }

    public function testMatcher()
    {
        $double = $this->createMock(Cart::class);
        $double->method('total')->willReturn(9);
        $double
            ->method('count');
        $double->expects($this->once())->method('total');
    }
}
```

## Divergences

- **`resource` (custos diverges).** A native return type `resource` names a
  class called `resource` (PHP has no resource type declaration), so a
  function declared `: resource` returns an object and
  `assertInternalType('resource', f())` is not redundant; custos never
  reports the `resource` type name.
- **Case of method names (custos diverges from upstream).** Upstream
  compares the assertion, `expects` and `any` names case-sensitively, so
  `$this->AssertNull($this->clear())` is not reported although PHP calls the
  same method. custos compares them case-insensitively.
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/internals/decisions.md` ("Spec-level false positives").
- `assertInstanceOf` whose argument 0 is not a resolvable `X::class`
  (e.g. `'Shop\Cart'`, `$class`, an unknown class): upstream falls back to
  "no expected type" and reports whenever D2–D5 hold, regardless of the
  class asked for. Recommendation (adopted above): no report in that case.
  No upstream fixture covers it.
- **`assertInternalType` — custos diverges from upstream.** Upstream reports
  it whenever the checked value has a single declared return type, without
  reading the requested type name. When the name contradicts the declared
  type (`assertInternalType('array', f())` with `f(): Cart`), or is not a
  type PHPUnit knows (the assertion then errors out), the assertion is not
  redundant and "the declared return type already guarantees this" is
  wrong. custos reports only when the declared type always satisfies the
  named type (D6 table). The upstream fixture's assertInternalType cases are
  no longer reported; listed in `testdata/ea-divergences.json`.
- Upstream compares the class FQN case-sensitively; custos compares
  case-insensitively (PHP class names are case-insensitive). Identical
  results on all upstream fixtures.
