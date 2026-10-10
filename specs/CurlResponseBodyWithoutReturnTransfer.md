---
id: CurlResponseBodyWithoutReturnTransfer
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlResponseBodyWithoutReturnTransfer

## Summary

Detect curl_exec result used as a string or JSON document when a locally created handle is proven never configured with CURLOPT_RETURNTRANSFER true.

## Detection

- D1. Report curl_exec result used as a string or JSON document when a locally created handle is proven never configured with CURLOPT_RETURNTRANSFER true.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Unknown handle origins/options, enabled return transfer, custom write callbacks with separately consumed body, and success checks are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Enable response-body return before decoding curl output.`

## Fix

No automatic fix: enabling return transfer changes output behavior and failure handling remains necessary.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$ch = curl_init('https://example.invalid/'); $body = curl_exec($ch); <warning descr="Enable response-body return before decoding curl output.">json_decode($body, true)</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
