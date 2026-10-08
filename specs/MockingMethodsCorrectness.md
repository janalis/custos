---
id: MockingMethodsCorrectness
group: PHPUnit
kind: semantic
needs: [names, index, hierarchy]
php: { min: "", max: "" }
---

# MockingMethodsCorrectness

## Summary

Two classes of PHPUnit mock-configuration mistakes inside tests:
passing a stub *object* (`$this->returnValue(...)`, `$this->returnCallback(...)`)
to `willReturn()` — which then returns the stub object itself instead of
running it — and configuring (`->method('name')`) a method that the mocked
class does not have, or that is `final` and therefore cannot be overridden by
the generated double.

## Detection

All checks run on method calls (`->`, `?->` or `::`) and only in a **test
context**:

- **D0** Test context: the file path ends with `Test.php`, `Spec.php` or
  `.phpt`, or contains `/Fixtures/`; or the innermost enclosing class FQN ends
  with `Test` or contains `\Tests\` or `\Test\`. Outside a test context
  nothing is reported.

### Part A — `willReturn()` given a stub object

Method names named in this spec (`willReturn`, `returnCallback`,
`returnValue`, `method`, `expects`, `getMock`, `getMockBuilder`,
`setMethods`) compare case-insensitively, as PHP resolves method names;
"named exactly" below means "no other name", not "same case".

- **D1** A method call named `willReturn`, with exactly one argument.
- **D2** That argument is itself a method call (`->`, `?->` or `::`, any
  receiver: `$this`, `self`, `static`, a variable…) whose name is exactly
  `returnCallback` or `returnValue`. Its own arguments are irrelevant.
  Plain function calls (`returnValue(1)`), other stub factories
  (`returnArgument`, `returnSelf`, `returnValueMap`, `onConsecutiveCalls`,
  `throwException`), or the stub wrapped in parentheses are not matched.
- Report the `willReturn` name token (finding **WILL**).

### Part B — configured method does not exist / is final

- **D3** A method call `C` whose name as written is exactly `method`, with
  exactly one argument `L` which is a string literal (single- or
  double-quoted; heredoc/nowdoc count as string literals too).
- **D4** Let `R` be the receiver expression of `C` (the expression left of
  `->`/`::`). If `R` is itself a method call named exactly `expects` (any
  arguments, including none), replace `R` with *that* call's receiver
  (one level only). A receiver wrapped in parentheses is not unwrapped at
  this step.
- **D5** Compute the possible values of `R` (rules V1–V7 below). There must be
  **exactly one** value `M`, and `M` must be a method call named exactly
  `getMock`.
- **D6** Search the descendants of `M` (pre-order: a node before its
  children, children left to right — so along a call chain the call closest
  to `getMock` comes first, receivers before arguments) for the first method
  call named exactly `getMockBuilder` → `B`. None → stop.
- **D7** Search the descendants of `M` the same way for the first method call
  named exactly `setMethods` → `S`. If `S` exists, its first argument is an
  array literal (`[...]` or `array(...)`), and **any** string literal anywhere
  inside that array (values, keys, nested arrays) has source text identical to
  `L`'s source text (quotes included, so `'run'` ≠ `"run"`), stop — the method
  was explicitly added to the double. custos also searches the first
  `addMethods` call the same way, comparing the literal contents
  case-insensitively (see Divergences); `onlyMethods()` is not recognised
  (it only lists existing methods).
- **D8** `B` must have exactly one argument, of the form `X::class` where `X`
  is a class name (including `self`/`static`/`parent` when they resolve);
  `X` must resolve to a class-like `K` (class, interface, trait, enum).
  Anything else (string class name, variable, `$obj::class`, unresolvable
  name) → stop.
- **D9** Look up a method named `L`'s raw content (between the quotes, no
  escape processing; for a heredoc/nowdoc, its body with the closing
  marker's indentation removed from each line, as PHP 7.3+ does; method-name matching case-insensitive, like PHP) in `K`
  including inherited members (parent classes, implemented interfaces, used
  traits).
  - not found → report `L` (finding **MISSING**). A magic `__call` on `K`
    does not prevent the report.
  - found and declared `final` → report `L` (finding **FINAL**).
  - found and not final → nothing (visibility, static-ness and abstractness
    are not checked).

Possible-values rules for D5 (applied recursively; each node visited at
most once; parentheses around any expression are stripped first):

- **V1** Ternary `c ? a : b` / `c ?: b`: union of the values of both result
  branches.
- **V2** `a ?? b`: union of the values of `a` and `b`.
- **V3** Variable `$v`: take the innermost enclosing function, method,
  closure or arrow function (none, i.e. top-level code → no values). Add the
  values of the default of a same-named parameter if it has one; then for
  every plain `=` assignment (by value or `=&`; not compound assignments, not
  list destructuring) anywhere in that function's body (nested closures
  included) whose left side is equivalent to `$v`, add the values of the
  assigned expression (for chains `$v = $w = e` use the innermost `e`).
  Assignment position is irrelevant: an assignment *after* the use counts.
  **Unstable variable** (custos refinement, see Divergences): if `$v` is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown* → no report.
- **V4** Property fetch `$this->p` / `$o->p`: resolve to the declared
  property; add the values of its default (unless the default's source text
  ends with the property name); then apply the V3 assignment scan for that
  property fetch in the enclosing function and in the declaring class's
  constructor.
- **V5** Class constant fetch `A::K`: add the values of the constant's
  initializer.
- **V6** Global constant (other than `true`/`false`/`null`): add the value of
  its `define()` / `const` definition as-is; unresolvable → no values.
- **V7** Anything else (a method call, `new`, literal, …): the expression
  itself.

So `$this->getMockBuilder(K::class)->getMock()->method('x')` is checked
directly (V7), and `$double` is checked when it has exactly one assignment in
the function, e.g. `$double = $this->getMockBuilder(K::class)->getMock();`.

## Exceptions (no report)

- **E1** Code outside a test context (D0).
- **E2** `willReturn($stub)` where the stub is not a direct
  `returnValue`/`returnCallback` method call, or `willReturn` with zero or
  two-plus arguments; `will($this->returnValue(...))` (the correct form).
- **E3** `method()` with a non-literal argument (`method($name)`,
  `method(self::NAME)`), or with zero / several arguments.
- **E4** Receivers whose possible values are not exactly one `getMock` call:
  `createMock(...)`, `getMockForAbstractClass()`, `prophesize(...)`,
  variables assigned twice (even with identical values), parameters without
  default, top-level variables.
- **E5** Builders without a `X::class` argument, unresolvable classes.
- **E6** Method listed in `setMethods([...])` with identical literal text.
- **E7** Existing non-final methods (own or inherited).

## Report

- Range:
  - WILL: the method-name identifier `willReturn` only.
  - MISSING / FINAL: the string literal argument `L` including its quotes.
- Severity:
  - WILL: warning (rule default).
  - MISSING and FINAL: **error** (forced, regardless of the rule default).
- Messages (our wording):
  - WILL: `The stub object is returned as-is here; use '->will(...)'.`
  - MISSING: `The mocked class has no such method.`
  - FINAL: `Final methods cannot be mocked.`

## Fix

- **F1** (WILL only) Rename the method call `willReturn` to `will`: only the
  name identifier changes; receiver, arrow, argument list, whitespace and
  comments are untouched.
  `$double->method('a')->willReturn($this->returnValue(3));` →
  `$double->method('a')->will($this->returnValue(3));`
- MISSING / FINAL: no fix.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php

class Ledger
{
    public function balance() {}
    final public function seal() {}
}

class LedgerTest
{
    public function testDoubles()
    {
        $stub = $this->getMockBuilder(Ledger::class)
            ->setMethods(['virtualTotal'])
            ->getMock();

        $stub->method('balance')-><warning descr="The stub object is returned as-is here; use '->will(...)'.">willReturn</warning>($this->returnValue(42));
        $stub->method('balance')-><warning descr="The stub object is returned as-is here; use '->will(...)'.">willReturn</warning>(self::returnCallback('strtoupper'));
        $stub->method('balance')->will($this->returnValue(42));
        $stub->method('balance')->willReturn(42);

        $stub->method(<error descr="Final methods cannot be mocked.">'seal'</error>)->willReturn(true);
        $stub->method(<error descr="The mocked class has no such method.">'refund'</error>)->willReturn(0);
        $stub->expects($this->once())->method(<error descr="The mocked class has no such method.">"refund"</error>);
        $stub->method('virtualTotal')->willReturn(7);
        $stub->method('BALANCE')->willReturn(1);
        $stub->method($dynamic)->willReturn(1);

        $this->getMockBuilder(Ledger::class)->getMock()->method(<error descr="Final methods cannot be mocked.">'seal'</error>);
        $this->createMock(Ledger::class)->method('refund');
    }
}
```

```php
<?php

class Ledger
{
    public function balance() {}
    final public function seal() {}
}

class LedgerTest
{
    public function testDoubles()
    {
        $stub = $this->getMockBuilder(Ledger::class)
            ->setMethods(['virtualTotal'])
            ->getMock();

        $stub->method('balance')->will($this->returnValue(42));
        $stub->method('balance')->will(self::returnCallback('strtoupper'));
        $stub->method('balance')->will($this->returnValue(42));
        $stub->method('balance')->willReturn(42);

        $stub->method('seal')->willReturn(true);
        $stub->method('refund')->willReturn(0);
        $stub->expects($this->once())->method("refund");
        $stub->method('virtualTotal')->willReturn(7);
        $stub->method('BALANCE')->willReturn(1);
        $stub->method($dynamic)->willReturn(1);

        $this->getMockBuilder(Ledger::class)->getMock()->method('seal');
        $this->createMock(Ledger::class)->method('refund');
    }
}
```

## Divergences

- **Case of method names (custos diverges from upstream).** Upstream
  compares the PHPUnit method names case-sensitively, so
  `->WillReturn($this->ReturnValue(1))` or `->METHOD('missing')` are not
  checked although PHP calls the same methods. custos compares them
  case-insensitively.
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/internals/decisions.md` ("Spec-level false positives").
- D7 compares literals by raw source text, so `setMethods(["run"])` does not
  cover `method('run')`. Recommendation: compare unquoted contents; no fixture
  depends on the quote style.
- **`addMethods([...])` (custos diverges).** Upstream ignores
  `addMethods` (PHPUnit 8.3+/9), so a method added to the double that way
  (`getMockBuilder(\stdClass::class)->addMethods(['getRepository'])`) is
  reported as missing. custos treats it like `setMethods` (D7), matching
  method names case-insensitively.
- A class with `__call` is still reported as MISSING; kept (matches
  upstream, PHPUnit cannot configure magic methods without `addMethods`).
- The `expects(...)` unwrap is skipped when the receiver is parenthesised
  (`($d->expects($x))->method('a')`), which then yields no report; harmless.
- **Unresolvable ancestors (custos diverges).** Upstream reports a method
  as missing when the mocked class extends, implements or uses a type the
  project does not contain (e.g. a vendor base class outside the analysed
  tree): the method may well be declared there. custos reports MISSING only
  when every parent, interface and trait of the class resolves; FINAL is
  unaffected.
