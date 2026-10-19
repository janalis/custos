---
id: ArrayUniqueOnNonStringableObjects
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayUniqueOnNonStringableObjects

## Summary

Default array_unique comparison needs string representations. Use an explicit identity or key when objects cannot be converted to strings.

## Detection

- D1. Resolved array_unique with default/SORT_STRING flag receives a literal array containing new stdClass or a resolved object class proven to lack __toString/Stringable. Do not infer missing methods from incomplete index.
- D1a. A class is proven non-stringable only when all of the following hold: neither it nor any ancestor declares `__toString` or implements `Stringable`; every parent class and used trait in its hierarchy resolves and the hierarchy is fully enumerated; and every PHP-provided class in the hierarchy (the class itself or an inherited one) belongs to the closed list of builtins verified to raise an error on string conversion: `stdClass`, `ArrayObject`, `ArrayIterator`, `DateTime`, `DateTimeImmutable`, `DateTimeZone`, `DateInterval`, `SplObjectStorage`, `SplDoublyLinkedList`, `SplStack`, `SplQueue`, `SplFixedArray`, `SplPriorityQueue`, `SplMinHeap`, `SplMaxHeap`, `WeakMap`. Engine and extension classes may convert to string through an internal handler without exposing a `__toString` method (`GMP` does), so a missing method proves nothing for any other PHP-provided class.
- D1d. Evaluate the effective array after PHP key normalization and duplicate-key overwrites, recursively where nested values supply evidence. Only surviving entries establish row collisions, nested arrays, non-stringable objects, formatted numeric strings or nested list shapes. Earlier entries overwritten by a later occurrence of the same normalized key are not evidence; unknown keys or unsupported effective-array evaluation discard proof.
- D1e. Require at least two surviving effective array entries. A singleton is returned unchanged without requiring comparison/string conversion and is excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E1a. Do not report objects of a PHP-provided class outside the D1a list, or of a project class inheriting from one: `array_unique([new GMP(5), new GMP(5)])` deduplicates without error. Do not report a class whose parent or trait cannot be resolved.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Choose an explicit comparison for object deduplication.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

Effective array-key behavior follows the [PHP array contract](https://www.php.net/manual/en/language.types.array.php).

## Examples

```php
<?php
<warning descr="Choose an explicit comparison for object deduplication.">array_unique([new stdClass(),new stdClass()])</warning>;
```

Valid cases:

```php
<?php
$u=array_unique(["a","a"]);
$n=array_unique([new GMP(5),new GMP(5),new GMP(7)]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
