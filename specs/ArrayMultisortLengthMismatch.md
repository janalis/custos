---
id: ArrayMultisortLengthMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayMultisortLengthMismatch

## Summary

Parallel arrays passed to array_multisort must have equal lengths. Align the records before sorting them together.

## Detection

- D1. Resolved array_multisort receives at least two statically known arrays of distinct lengths. Resolve simple preceding literal assignments; sorting flag scalar arguments do not count as arrays.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Provide parallel arrays of equal length.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$a=[1,2]; $b=[3]; <warning descr="Provide parallel arrays of equal length.">array_multisort($a,$b)</warning>;
```

Valid case:

```php
<?php
$a=[2,1]; $b=["x","y"]; array_multisort($a,$b);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
