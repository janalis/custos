---
id: DocumentedForeachOverNonIterable
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# DocumentedForeachOverNonIterable

## Summary

Report a `foreach` over a value that PHPDoc says is neither an array nor an
object. The loop may still run, since PHP does not check PHPDoc, but the loop
and the documentation disagree. Subjects the code alone proves are not
iterable are ForeachOverNonIterable's.

## Detection

- D1. The subject of a `foreach`.
- D2. The type of the subject is the one PHPDoc-aware inference gives:
  `@param`, `@var`, `@return` and `@property` tags count.
- D3. Every member of that type is an int, a float, a string, a bool or null.

## Exceptions (no report)

- E1. Subjects ForeachOverNonIterable reports.
- E2. A subject of unknown type, or one that may be an array, an object,
  `iterable`, `callable` or `mixed`.

## Report

- Range: the subject of the loop.
- Severity: warning.
- Message: `Iterate over an array or an object, not <type>.`

## Fix

None: either the loop or the documentation is wrong.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
/** @param string $names */
function greet($names): void
{
    foreach (<warning descr="Iterate over an array or an object, not string.">$names</warning> as $name) {
    }
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
