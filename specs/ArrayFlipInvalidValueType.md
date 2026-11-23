---
id: ArrayFlipInvalidValueType
group: Probable bugs
kind: semantic
needs: [names]
php: { min: "", max: "" }
---

# ArrayFlipInvalidValueType

## Summary

Detect array_flip of a literal array containing a value proven neither string nor integer.

## Detection

- D1. Report array_flip of a literal array containing a value proven neither string nor integer.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases or recursion cannot establish definite facts. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Unknown values are not evidence; resolved user functions named array_flip are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Supply only string or integer values to array_flip.`

## Fix

No automatic fix: conversion changes key meaning and may collide.

## Options

None.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within the supported PHP versions). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<warning descr="Supply only string or integer values to array_flip.">array_flip([true])</warning>;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
