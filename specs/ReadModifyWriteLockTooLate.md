---
id: ReadModifyWriteLockTooLate
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ReadModifyWriteLockTooLate

## Summary

Detect file_get_contents of a path feeding a transformed value into file_put_contents on the same proven path with LOCK_EX but no enclosing lock spanning the read.

## Detection

- D1. Report file_get_contents of a path feeding a transformed value into file_put_contents on the same proven path with LOCK_EX but no enclosing lock spanning the read.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Separate destinations, read-only paths, established spanning locks, atomic database alternatives and unknown path identity are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Hold the lock across the complete read-modify-write operation.`

## Fix

No automatic fix: the complete locked read-modify-write protocol must handle partial reads/writes and failures.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $n = (int) file_get_contents($path); <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
