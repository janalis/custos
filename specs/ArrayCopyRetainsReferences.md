---
id: ArrayCopyRetainsReferences
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayCopyRetainsReferences

## Summary

Copying an array does not detach reference-valued elements. Changing a referenced element through the copy can also change the original array. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. An explicit reference is bound to a literal-index element of a local array. The array is copied to a distinct local, and that copy writes the same index while the reference remains live. The original element is subsequently read. Require straight-line provenance.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Detach referenced array elements before modifying the copy.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$a=[1]; $r=&$a[0]; $b=$a; <warning descr="Detach referenced array elements before modifying the copy.">$b[0]=9</warning>; echo $a[0];
```

Valid case:

```php
<?php
$a=[1]; $b=$a; $b[0]=9; echo $a[0];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
