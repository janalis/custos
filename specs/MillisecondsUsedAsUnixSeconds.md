---
id: MillisecondsUsedAsUnixSeconds
group: Probable bugs
kind: semantic
needs: [names]
php: { min: "", max: "" }
---

# MillisecondsUsedAsUnixSeconds

## Summary

Detect date, gmdate or DateTime setTimestamp consuming a positive literal timestamp between 100000000000 and 9999999999999 under an explicit millisecond-input contract.

## Detection

- D1. Report date, gmdate or DateTime setTimestamp consuming a positive literal timestamp between 100000000000 and 9999999999999 under an explicit millisecond-input contract.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or code too complex to analyze cannot establish definite facts.

## Exceptions (no report)

- E1. No magnitude-only reports when the rule is disabled; unknown units and arithmetic already converting units are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Convert millisecond timestamps to seconds explicitly.`

## Fix

No automatic fix: rounding negative/fractional timestamps and intended dates require caller policy.

## Options

None.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within the supported PHP versions). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
echo <warning descr="Convert millisecond timestamps to seconds explicitly.">date('c',1730000000000)</warning>;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
