---
id: ForeachOverNonIterable
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# ForeachOverNonIterable

## Summary

Report a `foreach` over a value that the code itself shows is neither an
array nor an object. PHP raises a warning and skips the loop.

## Detection

- D1. The subject of a `foreach`.
- D2. The type of the subject comes from native declarations only, and every
  member of it is an int, a float, a string, a bool or null.

## Exceptions (no report)

- E1. A subject of unknown type, or one that may be an array, an object,
  `iterable`, `callable` or `mixed`: a `foreach` walks the properties of any
  object.
- E2. A type that comes only from PHPDoc.

## Report

- Range: the subject of the loop.
- Severity: warning.
- Message: `Iterate over an array or an object, not <type>.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function total(int $count): void
{
    foreach (<warning descr="Iterate over an array or an object, not int.">$count</warning> as $item) {
    }
}
```
