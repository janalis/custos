---
id: ArraySliceDiscardsRequiredKeys
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArraySliceDiscardsRequiredKeys

## Summary

Array slicing normally renumbers numeric keys. Preserve keys when later code reads the original numeric positions. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved array_slice has absent/false preserve_keys, a literal array with an explicitly written nonzero integer key in the retained slice, and a subsequent result access uses that original key that is absent after reindexing.
- D1a. The initial supported downstream proof is a single immediately following echo of the absent original numeric key. Input array keys must all be explicit known integers.
- D1e. All literal numeric input keys must be known and unique after PHP key normalization. Duplicate or unknown keys are excluded: overwriting can change the effective length and selected slice positions, so original literal item positions cannot establish the retained-key proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Preserve numeric keys before reading the original key.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$p=<warning descr="Preserve numeric keys before reading the original key.">array_slice([100=>"a",200=>"b"],0,1)</warning>; echo $p[100];
```

Valid case:

```php
<?php
$a=array_slice([100=>"a"],0,1,true); echo $a[100];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
