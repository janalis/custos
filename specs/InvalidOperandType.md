---
id: InvalidOperandType
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "8.0", max: "" }
---

# InvalidOperandType

## Summary

Report an arithmetic or bitwise operation on an array or an object that the
code itself shows cannot take part in it. PHP throws a `TypeError`.

## Detection

- D1. An operand of `+`, `-`, `*`, `/`, `%`, `**`, `&`, `|`, `^`, `<<` or `>>`,
  or of the compound assignment of one of them.
- D2. The type of the operand comes from native declarations only, and every
  member of it is an array, or every member of it is a class whose whole
  hierarchy is known.

## Exceptions (no report)

- E1. `+` between arrays, which merges them: an array operand of `+` is
  reported only when the other operand is known and cannot be an array.
- E2. An operand that may also be null, a scalar or an unresolved class.
- E3. An object of a class that takes part in arithmetic: `GMP`,
  `BcMath\Number`, `FFI\CData`, `SimpleXMLElement` and their subclasses.
- E4. A type that comes only from PHPDoc.

## Report

- Range: the operand.
- Severity: error.
- Message: `Use a number here, not <type>.`

## Fix

None.

## Options

None.

## PHP versions

Active from PHP 8.0, where these operations throw a `TypeError`.

## Examples

```php
<?php
function bump(array $counts): int
{
    return <error descr="Use a number here, not array.">$counts</error> * 2;
}
```
