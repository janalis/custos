---
id: ArrayColumnDuplicateIndexLoss
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayColumnDuplicateIndexLoss

## Summary

Using duplicate index values in array_column overwrites earlier rows. Use unique indices or group rows when every value must remain available.

## Detection

- D1. Resolved array_column receives a literal array of literal rows, a literal column key, and a literal index key whose normalized array-key values repeat in two rows. Both rows contain the requested column.
- D1d. Evaluate the effective array after PHP key normalization and duplicate-key overwrites, recursively where nested values supply evidence. Only surviving entries establish row collisions, nested arrays, non-stringable objects, formatted numeric strings or nested list shapes. Earlier entries overwritten by a later occurrence of the same normalized key are not evidence; unknown keys or unsupported effective-array evaluation discard proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use unique index values or group the duplicate rows.

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
$a=<warning descr="Use unique index values or group the duplicate rows.">array_column([["id"=>7,"v"=>"a"],["id"=>7,"v"=>"b"]],"v","id")</warning>;
```

Valid case:

```php
<?php
$a=array_column([["id"=>7,"v"=>"a"],["id"=>8,"v"=>"b"]],"v","id");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
