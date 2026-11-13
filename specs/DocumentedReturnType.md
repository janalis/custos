---
id: DocumentedReturnType
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# DocumentedReturnType

## Summary

Report a returned value that breaks the function's return contract when
PHPDoc is taken into account: the `@return` type rejects it, or the
declared return type rejects a value whose type only PHPDoc describes. PHP
may run such a `return`, but the code and its documentation disagree.
Mismatches the code alone proves are InvalidReturnType's.

## Detection

- D1. Returned values and the functions that own them as
  InvalidReturnType defines them.
- D2. The type of the value is the one PHPDoc-aware inference gives:
  `@param`, `@var` and `@return` tags count.
- D3. The declared return type or the function's own `@return` type
  (`@phpstan-return` and `@psalm-return` first) accepts no value of the
  value's type. `$this` in a `@return` type accepts nothing that the
  declaring class rejects.
- D4. Integer ranges, string lengths and literal values of the `@return`
  type count only against a value whose possible values are all known
  (an integer literal such as `-1`, the count of a literal array) or, for
  a string, whose values or byte lengths are (a string literal such as
  `''` or `'gone'`); otherwise `@return positive-int` accepts any integer
  and `@return 'open'|'closed'` any string.
- D5. The weak-mode conversions and the exceptions of InvalidReturnType
  apply alike.

## Exceptions (no report)

- E1. Values InvalidReturnType reports: the declared type rejects what the
  code alone shows.
- E2. A `@return` type naming a template, a conditional `@return` type,
  and a `@return` type the declared type rejects (a stale tag).
- E3. A `@return` type the method does not write itself: inherited from a
  parent declaration, or through `{@inheritDoc}`.
- E4. Every exception of InvalidReturnType.

## Report

- Range: the returned value.
- Severity: warning.
- Message: `Return Order from latest(), not string.`, with the documented
  or declared type that rejects the value.
  The given type names the value itself (`not 0.`, `not ''.`) when only
  the range or length of the documented type rejects it.

## Fix

None: either the code or the documentation is wrong.

## Options

None.

## PHP versions

All versions.

## Examples

```php
<?php
declare(strict_types=1);

class Order {}
/** @return Order */
function latest()
{
    return <warning descr="Return Order from latest(), not string.">'order-1'</warning>;
}

/** @param string $amount */
function cents($amount): int
{
    return <warning descr="Return int from cents(), not string.">$amount</warning>;
}
```

A template, or a range compared with a value whose integers are not all
known, is not reported:

```php
<?php
/**
 * @template T
 * @param T $value
 * @return T
 */
function keep($value)
{
    return 'x';
}

/** @return positive-int */
function quantity(int $n)
{
    return $n;
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
