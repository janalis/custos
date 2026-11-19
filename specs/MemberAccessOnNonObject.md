---
id: MemberAccessOnNonObject
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# MemberAccessOnNonObject

## Summary

Report a method call or property access on a value that the code itself
shows is never an object. Calling a method or writing a property on a scalar
or an array throws an `Error`; reading a property raises a warning and yields
null.

## Detection

- D1. A method call or a property read or write through `->` or `?->`.
- D2. The type of the receiver comes from native declarations only (typed
  parameters, properties and returns, literals, operations), and without null
  every member of it is an int, a float, a string, a bool or an array.

## Exceptions (no report)

- E1. A receiver of unknown type, or one that may be `mixed`, `object`,
  `iterable`, `callable` or a class.
- E2. A receiver that can only be null, which PossiblyNullAccess covers.
- E3. A type that comes only from PHPDoc.
- E4. A property read inside `isset()`, `empty()` or on the left of `??`,
  which PHP evaluates without complaint.

## Report

- Range: the receiver.
- Severity: error.
- Message: `Access this member on an object, not <type>.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function label(int $id): string
{
    return <error descr="Access this member on an object, not int.">$id</error>->name;
}
```
