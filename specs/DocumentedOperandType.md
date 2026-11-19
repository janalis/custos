---
id: DocumentedOperandType
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "8.0", max: "" }
---

# DocumentedOperandType

## Summary

Report an arithmetic or bitwise operation on an array or an object when
PHPDoc says the operand cannot take part in it. The code may hold a number at
run time, since PHP does not check PHPDoc, but the operation and the
documentation disagree. Operands the code alone proves wrong are
InvalidOperandType's.

## Detection

- D1. An operand of `+`, `-`, `*`, `/`, `%`, `**`, `&`, `|`, `^`, `<<` or `>>`,
  or of the compound assignment of one of them.
- D2. The types of both operands are the ones PHPDoc-aware inference gives:
  `@param`, `@var`, `@return` and `@property` tags count.
- D3. Every member of the operand's type is an array, or every member of it
  is a class whose whole hierarchy is known.

## Exceptions (no report)

- E1. Operands InvalidOperandType reports.
- E2. `+` between arrays, which merges them: an array operand of `+` is
  reported only when the other operand is known and cannot be an array.
- E3. An operand that may also be null, a scalar or an unresolved class.
- E4. An object of a class that takes part in arithmetic: `GMP`,
  `BcMath\Number`, `FFI\CData`, `SimpleXMLElement` and their subclasses.

## Report

- Range: the operand.
- Severity: warning.
- Message: `Use a number here, not <type>.`

## Fix

None: either the operation or the documentation is wrong.

## Options

None.

## PHP versions

Active from PHP 8.0, where these operations throw a `TypeError`.

## Examples

```php
<?php
/** @param int[] $counts */
function bump($counts): int
{
    return <warning descr="Use a number here, not int[].">$counts</warning> * 2;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
