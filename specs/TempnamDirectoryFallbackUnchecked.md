---
id: TempnamDirectoryFallbackUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# TempnamDirectoryFallbackUnchecked

## Summary

Detect tempnam result used for sensitive content under an explicitly enabled required-directory contract without verifying its parent directory.

## Detection

- D1. Report tempnam result used for sensitive content under an explicitly enabled required-directory contract without verifying its parent directory.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Verified canonical parent, intentional system-temp fallback, unknown sensitivity and tempnam failure already handled are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Verify the temporary file remains in the required directory.`

## Fix

No automatic fix: fallback rejection and cleanup policy must be supplied explicitly.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($privateDir, $secret) { $path = tempnam($privateDir, 'job'); if ($path === false) { return; } <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
