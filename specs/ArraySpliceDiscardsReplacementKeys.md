---
id: ArraySpliceDiscardsReplacementKeys
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArraySpliceDiscardsReplacementKeys

## Summary

Array splicing discards the keys supplied by its replacement array. Insert map entries with an operation that preserves their keys. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved array_splice inserts a literal replacement with a string key absent from original literal array; a subsequent access on the modified array reads that string key. Require straight-line local mutation.
- D1a. The initial supported downstream proof is a single immediately following echo of the discarded string replacement key.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Preserve replacement keys before reading the inserted key.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$a=[]; <warning descr="Preserve replacement keys before reading the inserted key.">array_splice($a,0,0,["name"=>"Ada"])</warning>; echo $a["name"];
```

Valid case:

```php
<?php
$a=[]; $a["name"]="Ada"; echo $a["name"];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
