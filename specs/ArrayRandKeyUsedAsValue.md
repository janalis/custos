---
id: ArrayRandKeyUsedAsValue
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayRandKeyUsedAsValue

## Summary

array_rand returns a key rather than the selected element. Read the element at that key when the consumer requires an array value. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved array_rand with default or constant-one count operates on a known string-value list. Its result is passed directly to a consumer provably requiring one of those values (a local function with an explicit switch over the value literals). No name heuristic.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Read the array element at the returned random key.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
function color($v){switch($v){case "red":break;case "blue":break;}} $a=["red","blue"]; color(<warning descr="Read the array element at the returned random key.">array_rand($a)</warning>);
```

Valid case:

```php
<?php
$a=["red","blue"]; echo $a[array_rand($a)];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
