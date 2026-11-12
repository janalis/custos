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
object can belong to at once, two locals holding different literals. Such
a comparison always gives the same answer, so the branch it guards is dead
or always taken. Types that come only from PHPDoc never count.

## Detection

- D1. A `===` or `!==` comparison. Neither operand is the `null` literal
  (RedundantTypeCheck judges those), and not both are literals.
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
  that are or extend a class declared under a `vendor` directory: another
  release of that package, which the project may also support, may relate
  them differently.
- E6. Code in a test context (a file ending with `Test.php`, `Spec.php` or
  `.phpt`, a `Fixtures` directory, or a test class), and comparisons inside
  an `assert()` call.
- E7. Loose comparisons (`==`, `!=`), which convert their operands.
- E8. Comparisons other rules report, when they run: an operand of an
  `&&`/`||` chain one of whose operands tests it with `instanceof`
  (NotOptimalIfConditions); `array_search()` compared with `true`
  (ArraySearchUsedAsInArray); `xml_parse()` compared with `false`
  (XmlParseZeroComparedWithFalse); a `Collator` comparison result compared
  with `true` (CollatorComparisonTruthinessReversed). The last two follow
  the result through a local.

## Report

- Range: the comparison.
- Severity: error.
- Message: `This comparison is always false: int and string share no value.`
  (`always true` for `!==`), with both types, the exact values when they
  decided it.

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
}
```

Comparisons that may hold, PHPDoc-only types and loose comparisons are
silent:

```php
<?php
/** @param positive-int $qty */
function guard(int $qty, ?string $code, int|string $key) {
    if ($qty <= 0 || $qty === 0 || $code === 'x' || $key === 1 || $qty == '1') {
        throw new InvalidArgumentException('qty');
    }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
