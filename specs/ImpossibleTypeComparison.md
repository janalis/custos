---
id: ImpossibleTypeComparison
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, flow, stubs]
php: { min: "", max: "" }
---

# ImpossibleTypeComparison

## Summary

Report a strict comparison (`===`, `!==`) whose operands can never be
identical, because the code itself shows they have no value in common: an
`int` against a `string`, an object against a scalar, two classes no
object can belong to at once, two locals holding different literals. Also
report a loose comparison (`==`, `!=`, `<>`) that PHP 8's conversions can
never make true, such as objects of two unrelated classes or an array
against a number. Such a comparison always gives the same answer, so the
branch it guards is dead or always taken. Types that come only from PHPDoc
never count.

## Detection

- D1. A `===`, `!==`, `==`, `!=` or `<>` comparison. Neither operand is
  the `null` literal (RedundantTypeCheck judges the strict ones), and not
  both are literals.
- D2. Each operand's type comes from native declarations only (typed
  parameters, properties and returns, literals, operations, the types
  PHP's own functions declare), narrowed by the conditions guarding the
  comparison, as for PossiblyNullAccess. The type must also be decided the
  same way with PHPDoc taken into account, and with the newest supported
  PHP version as target.
- D3. No value of one type is identical to a value of the other. Scalar
  types, `null`, arrays and objects never meet each other; `int` never
  meets `float`; `true` never meets `false`. Two classes meet unless one
  can never be the other: unrelated classes (single inheritance), or a
  final class (or enum) outside the other's subtypes. `callable`,
  `iterable` and `object` may meet the values they can hold.
- D4. Exact values tell two operands of the same type apart only when both
  are literals, or locals whose every reaching value is a literal assigned
  in the same function (`$mode = 'cart'; if ($bulk) { $mode = 'bulk'; }`
  never equals `'sale'`). Integer ranges and string lengths never do.
- D5. A loose comparison is reported only when every pair of values of the
  two types is one PHP 8 always finds unequal, whatever the values (exact
  values never decide it):

  | Left | Right | Why never equal |
  | --- | --- | --- |
  | plain object | plain object of a class no object of the first can be | objects of different classes are unequal |
  | plain object | array, `null`, `false` | an object converts to no array, and to `true` |
  | array | `int`, `float`, `string` | an array only equals an array, or `null` and a `bool` by truthiness |
  | `true` | `false`, `null` | both sides compare as booleans |

  A plain object is an instance of a class or an enum (not an interface,
  which an object of any class may implement) whose hierarchy is wholly
  known, no dependency declares (E5), and that extends no class of PHP
  itself: `DateTime` equals a `DateTimeImmutable` at the same instant, an
  empty `SimpleXMLElement` equals `false`, a `GMP` number equals an `int`.

  Every other pair may be equal and keeps the comparison silent: an object
  against an `int` or a `float` (it converts to 1, with a notice), `true`,
  or a `string` (through `__toString()`); `null` against `false`, `0`,
  `0.0`, `''` or `[]`; an array against a `bool`; a number against a
  numeric string (`1 == '01'`), an `int` against a `float`; `object`,
  `iterable` and `callable` values against anything.

## Exceptions (no report)

- E1. An operand whose type is unknown, includes `mixed`, or is an
  intersection; an unreachable comparison.
- E2. An operand that reads a constant (global or class constant, or an
  element of one), whose value may depend on configuration; `$this` inside
  a closure or an arrow function, which may run bound to another object.
- E3. An operand written again later in a loop around the comparison
  (assigned, incremented, iterated into, destructured into, unset), itself
  or a value it reads through: its type may change from one iteration to
  the next.
- E4. An operand that is, or a local assigned, a call to a PHP function or
  method whose declared return type changes after the target PHP version
  (`sort()` returns `bool` before PHP 8.2 and `true` after).
- E5. Classes whose hierarchy is not wholly known, traits, and classes
  that are or extend a class declared in the project's vendor directory
  (Composer's `config.vendor-dir`, `vendor` by default): another release of
  that package, which the project may also support, may relate them
  differently. A directory named `vendor` elsewhere holds project code.
- E6. Code in a test context (a file ending with `Test.php`, `Spec.php` or
  `.phpt`, a `Fixtures` directory, or a test class), and comparisons inside
  an `assert()` call.
- E7. Loose comparisons outside the pairs of D5, which convert their
  operands.
- E8. Comparisons other rules report, when they run and check that code:
  in the condition of an `if` or `elseif` that is an `&&` chain (one
  symbolic operator), a comparison of the value the first `instanceof` of
  the chain tests with a literal, also under prefix operators (`-1`,
  `!false`), unless loose against a string when that class has
  `__toString()` (NotOptimalIfConditions, when it reports `instanceof`
  flaws); a call to `array_search()` compared with `true`, unless the call
  is in parentheses (ArraySearchUsedAsInArray); `xml_parse()` compared with
  `false`, also through a local (XmlParseZeroComparedWithFalse); a call to
  the `compare()` method of `Collator` (or of a subclass) tested equal to
  `true` (CollatorComparisonTruthinessReversed). The function form
  `collator_compare()`, a collation result held in a local and a test that
  it is not identical to `true` stay this rule's.

## Report

- Range: the comparison.
- Severity: error.
- Message: `This comparison is always false: int and string share no value.`
  (`always true` for `!==`), with both types, the exact values when they
  decided it. For a loose comparison:
  `This comparison is always false: Order and Refund never compare equal.`
  (`always true` for `!=` and `<>`).

## Fix

None: the comparison reveals a wrong assumption, which only the author
can correct.

## Options

None.

## PHP versions

All versions.

## Examples

```php
<?php
final class Order {}
final class Refund {}
function compare(int $qty, string $code, Order $order, Refund $refund, bool $bulk): void {
    if (<error descr="This comparison is always false: int and string share no value.">$qty === $code</error>) {
        return;
    }
    $same = <error descr="This comparison is always false: Order and Refund share no value.">$order === $refund</error>;
    $mode = 'cart';
    if ($bulk) {
        $mode = 'bulk';
    }
    $sale = <error descr="This comparison is always false: 'bulk'|'cart' and 'sale' share no value.">$mode === 'sale'</error>;
    $loose = <error descr="This comparison is always false: Order and Refund never compare equal.">$order == $refund</error>;
}
```

Comparisons that may hold, PHPDoc-only types and loose comparisons a
conversion may make true are silent:

```php
<?php
final class Ticket {}
/** @param positive-int $qty */
function guard(int $qty, ?string $code, int|string $key, Ticket $ticket, ?array $tags) {
    if ($qty <= 0 || $qty === 0 || $code === 'x' || $key === 1 || $qty == '1' ||
        $ticket == 1 || $tags == $code) {
        throw new InvalidArgumentException('qty');
    }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
