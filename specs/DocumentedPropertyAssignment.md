---
id: DocumentedPropertyAssignment
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types]
php: { min: "", max: "" }
---

# DocumentedPropertyAssignment

## Summary

Report a value stored in a property that breaks the property's contract
when PHPDoc is taken into account: the `@var` type of the property rejects
it, or the declared type rejects a value whose type only PHPDoc describes.
PHP may run such an assignment, but the code and its documentation
disagree. Mismatches the code alone proves are InvalidPropertyAssignment's.

## Detection

- D1. Assignments, stored values and property resolution as
  InvalidPropertyAssignment defines them: the written class still comes
  from native declarations only.
- D2. The type of the value is the one PHPDoc-aware inference gives:
  `@param`, `@var` and `@return` tags count.
- D3. Every declaration the write may reach rejects every value of the
  stored type, through the declared type of the property or its `@var`
  type.
- D4. The weak-mode conversions and the exceptions of
  InvalidPropertyAssignment apply alike, except that an untyped property
  with a `@var` type is checked.

## Exceptions (no report)

- E1. Assignments InvalidPropertyAssignment reports: the declared types
  reject what the code alone shows.
- E2. A `@var` type naming templates, and a `@var` type the declared type
  rejects (a stale tag).
- E3. The value details of a `@var` type: integer ranges, string lengths
  and literal string values. `@var positive-int` accepts any integer.
- E4. Every exception of InvalidPropertyAssignment other than untyped
  properties.
- E5. `null` stored in a property whose default value is `null`: the
  default makes its `@var` type nullable, as it does for a parameter.

## Report

- Range: the assigned value.
- Severity: warning.
- Message: `Assign Order|null to Cart::$order, not string.`, with the
  documented or declared type that rejects the value.

## Fix

None: either the assignment or the documentation is wrong.

## Options

None.

## PHP versions

All versions.

## Examples

```php
<?php
final class Session
{
    /** @var string */
    public $token;

    public function close()
    {
        $this->token = <warning descr="Assign string to Session::$token, not null.">null</warning>;
    }
}
```

A template type is not compared:

```php
<?php
/** @template T */
final class Box
{
    /** @var T */
    public $value;

    public function clear() { $this->value = null; }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
