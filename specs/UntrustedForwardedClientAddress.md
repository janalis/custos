---
id: UntrustedForwardedClientAddress
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedForwardedClientAddress

## Summary

Detect HTTP_X_FORWARDED_FOR-derived address used in an access-control comparison without established trusted-proxy validation.

## Detection

- D1. Report HTTP_X_FORWARDED_FOR-derived address used in an access-control comparison without established trusted-proxy validation.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D4. Require a direct equality condition guarding a call to a project function explicitly annotated `@custos-protected`. Unknown intent and additional unmodelled proxy guards suppress the finding.
- D5. A condition may validate the proxy boundary when it reads `$_SERVER['REMOTE_ADDR']` or contains a method call, a static call, or a function call that does not resolve to a builtin. Such a condition counts as an unmodelled proxy guard when it belongs to an `if`/`elseif` chain enclosing the comparison's `if` within the same function scope, or to an earlier `if` chain in an enclosing statement list of that scope with at least one arm that never falls through (return, throw, exit, break, continue).
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Configured trusted-proxy resolver summaries, logging-only use, non-security uses and missing source-to-sink proof are excluded.
- E3. The comparison is governed by an unmodelled proxy guard as defined in D5: nested inside a guarded `if` chain, or placed after a guard clause that leaves the statement list. Guards outside the enclosing closure or function, earlier checks whose every arm falls through, and conditions made only of plain variables, builtin calls or other `$_SERVER` keys do not suppress the report.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Validate the trusted proxy boundary before trusting forwarded addresses.`

## Fix

No automatic fix: trusted proxy addresses and forwarding-chain policy must be supplied by deployment configuration.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
/** @custos-protected */ function grantAccess() {}
function bad($allowedIp) { if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $allowedIp</warning>) { grantAccess(); } }
```

No report when a proxy check already decides whether the comparison runs:

```php
<?php
/** @custos-protected */ function openGate() {}
function nested($expected) {
    if ($_SERVER['REMOTE_ADDR'] === '192.0.2.10') {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
function guardClause($expected) {
    if (!Proxies::knows($_SERVER['REMOTE_ADDR'])) { return; }
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
