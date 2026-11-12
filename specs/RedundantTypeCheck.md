---
id: RedundantTypeCheck
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, flow, stubs]
php: { min: "", max: "" }
---

# RedundantTypeCheck

## Summary

Report a type test whose answer the code already decides: `instanceof`,
`is_int()` and the other type predicates, or a strict comparison with
`null`, on a value whose native type makes it always true or always
false. An always-true check is dead defensive code; an always-false one
guards a branch that never runs, often a sign that the value is not what
the author thinks. Types that come only from PHPDoc never count, so a
guard that enforces a documented type at run time stays silent.

## Detection

- D1. One of these tests on a value:
  - `x instanceof C`, with `C` a written name (not `self`, `static` or
    `parent`) of a class or interface whose whole hierarchy is known;
  - a call to the global function `is_null`, `is_int`, `is_integer`,
    `is_long`, `is_float`, `is_double`, `is_string`, `is_bool`,
    `is_array`, `is_object`, `is_scalar` or `is_iterable`, with the tested
    value as its only argument (by position, or named `value`);
  - `x === null` or `x !== null`, in either operand order.
- D2. The value's type comes from native declarations only, narrowed by
  the conditions guarding the test, as for PossiblyNullAccess. The test
  must also be decided the same way with PHPDoc taken into account, and
  with the newest supported PHP version as target.
- D3. Every value of that type passes the test (always true), or none
  does (always false). A class passes `instanceof C` when it is `C` or a
  subtype of it, and never does when it cannot be: an unrelated class, or
  a final class (or enum) outside `C`'s subtypes. `callable`, `iterable`
  and `object` may pass the tests their values may pass.

## Exceptions (no report)

- E1. A value whose type is unknown, includes `mixed`, or is an
  intersection; an unreachable test.
- E2. A constant (global or class constant, or an element of one), whose
  value may depend on configuration; `$this` inside a closure or an arrow
  function, which may run bound to another object.
- E3. A value written again later in a loop around the test (assigned,
  incremented, iterated into, destructured into, unset), itself or a value
  it reads through: its type may change from one iteration to the next.
- E4. A value that is, or a local assigned, a call to a PHP function or
  method whose declared return type changes after the target PHP version
  (`sort()` returns `bool` before PHP 8.2 and `true` after).
- E5. Classes whose hierarchy is not wholly known, traits (`$this` in a
  trait is the class using it), and classes that are or extend a class
  declared under a `vendor` directory: another release of that package,
  which the project may also support, may implement more interfaces.
- E6. A test that is always true and is, through parentheses and `!`, the
  whole condition of the last `elseif` (or `else if`) of a chain without
  `else`: it names the case left rather than testing anything.
- E7. Code in a test context (a file ending with `Test.php`, `Spec.php` or
  `.phpt`, a `Fixtures` directory, or a test class), and tests inside an
  `assert()` call.
- E8. Tests other rules report, when they run: an always-false `is_int()`,
  `is_float()`, `is_string()`, `is_bool()`, `is_array()` or `is_object()`
  on a parameter of the enclosing function or method
  (CallableParameterUseCaseInTypeContext); a test in an `&&`/`||` chain
  another operand of which tests the same value with `instanceof`, or, for
  an `instanceof`, compares it with a literal (NotOptimalIfConditions); a
  `null` comparison of `DOMElement::getAttribute()`, directly or through a
  local (DomMissingAttributeCheckedAsNull). An `instanceof` of a trait is
  left to SuspiciousBinaryOperation.

## Report

- Range: the test, without a `!` before it.
- Severity: warning.
- Message: `This check is always true: the value is int.` (or `always
  false`), with the value's type, without value details.

## Fix

None: whether to drop the test, or to fix what the value holds, is the
author's choice.

## Options

None.

## PHP versions

All versions.

## Examples

```php
<?php
interface Priced {}
final class Order implements Priced {}
function make(): Order { return new Order(); }
function checks(int $qty, Order $order, ?Order $maybe): void {
    $flags = [
        <warning descr="This check is always true: the value is int.">is_int($qty)</warning>,
        <warning descr="This check is always true: the value is Order.">$order instanceof Priced</warning>,
        <warning descr="This check is always false: the value is Order.">make() === null</warning>,
    ];
    if ($maybe === null) {
        return;
    }
    if (<warning descr="This check is always true: the value is Order.">$maybe !== null</warning>) {
        echo 'set';
    }
}
```

Guards of documented types, open tests and the last branch of a chain are
silent:

```php
<?php
/** @param string $code */
function parse($code, array|string $value): string {
    if (!is_string($code)) {
        throw new InvalidArgumentException('code');
    }
    if (is_string($value)) {
        return $value;
    } elseif (is_array($value)) {
        return implode(',', $value);
    }
    return '';
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
