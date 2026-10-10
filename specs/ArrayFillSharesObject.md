---
id: ArrayFillSharesObject
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayFillSharesObject

## Summary

Filling an array with one object places the same object in every entry. Construct separate objects when entries must hold independent state. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. A resolved array_fill call uses a newly constructed object as value and constant count greater than one; a property write through one literal array index is followed by a read of that property through a different in-range index.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Construct a separate object for each array entry.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$a=<warning descr="Construct a separate object for each array entry.">array_fill(0,3,new stdClass())</warning>; $a[0]->id=1; echo $a[1]->id;
```

Valid case:

```php
<?php
$rows=array_fill(0,3,0); $rows[0]=1; echo $rows[1];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
