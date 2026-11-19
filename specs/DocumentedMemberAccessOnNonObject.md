---
id: DocumentedMemberAccessOnNonObject
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# DocumentedMemberAccessOnNonObject

## Summary

Report a method call or property access on a value that PHPDoc says is never
an object. The code may still run, since PHP does not check PHPDoc, but the
access and the documentation disagree. Receivers the code alone proves are
never objects are MemberAccessOnNonObject's.

## Detection

- D1. A method call or a property read or write through `->` or `?->`.
- D2. The type of the receiver is the one PHPDoc-aware inference gives:
  `@param`, `@var`, `@return` and `@property` tags count.
- D3. Without null, every member of that type is an int, a float, a string,
  a bool or an array.

## Exceptions (no report)

- E1. Receivers MemberAccessOnNonObject reports.
- E2. A receiver of unknown type, or one that may be `mixed`, `object`,
  `iterable`, `callable` or a class.
- E3. A receiver that can only be null, which PossiblyNullAccess covers.
- E4. A property read inside `isset()`, `empty()` or on the left of `??`,
  which PHP evaluates without complaint.
- E5. A receiver whose PHPDoc type the code contradicts: when the code
  alone gives the receiver a type that no value of the PHPDoc type fits
  (`/** @var int $x */` over `$x = new Order()`), the tag is stale.
- E6. `$this` inside a closure or an arrow function, which may be bound
  to an object of another class.

## Report

- Range: the receiver.
- Severity: warning.
- Message: `Access this member on an object, not <type>.`

## Fix

None: either the access or the documentation is wrong.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
/** @param int $id */
function label($id): string
{
    return <warning descr="Access this member on an object, not int.">$id</warning>->name;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
