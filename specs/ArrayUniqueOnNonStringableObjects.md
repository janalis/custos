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
- D1d. Evaluate the effective array after PHP key normalization and duplicate-key overwrites, recursively where nested values supply evidence. Only surviving entries establish row collisions, nested arrays, non-stringable objects, formatted numeric strings or nested list shapes. Earlier entries overwritten by a later occurrence of the same normalized key are not evidence; unknown keys or unsupported effective-array evaluation discard proof.
- D1e. Require at least two surviving effective array entries. A singleton is returned unchanged without requiring comparison/string conversion and is excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
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

Valid case:

```php
<?php
$u=array_unique(["a","a"]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
