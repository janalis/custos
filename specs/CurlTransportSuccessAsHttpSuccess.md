---
id: CurlTransportSuccessAsHttpSuccess
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlTransportSuccessAsHttpSuccess

## Summary

Detect curl_exec success used as the sole condition for a literal successful return when no HTTP response-code validation occurs for that handle.

## Detection

- D1. Report curl_exec success used as the sole condition for a literal successful return when no HTTP response-code validation occurs for that handle.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Non-HTTP protocols, configured fail-on-error, explicit permitted status checks, and escaped handles invalidate the report.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Check the HTTP response status before reporting success.`

## Fix

No automatic fix: accepted HTTP statuses are application-specific.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($url) { $ch = curl_init('https://example.invalid/api'); if (<warning descr="Check the HTTP response status before reporting success.">curl_exec($ch)</warning> !== false) { return true; } return false; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
