---
id: SameSiteNoneWithoutSecure
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.3", max: "" }
---

# SameSiteNoneWithoutSecure

## Summary

Detect setcookie with literal options selecting SameSite None and secure absent or false.

## Detection

- D1. Report setcookie with literal options selecting SameSite None and secure absent or false.
- D1a. Option names are compared without regard to letter case, as PHP does: `SameSite`, `SAMESITE`, `Secure` and `SECURE` name the same options as their lower-case spellings.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Secure true, other SameSite policies, unknown options and unrelated user functions are excluded.
- E1a. An options array that spells one option name twice with different letter case is not reported: the effective value depends on entry order.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Use a Secure cookie with SameSite None.`

## Fix

No automatic fix: enabling Secure can make a cookie unusable on non-HTTPS deployments; document the HTTPS prerequisite.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 7.3. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<warning descr="Use a Secure cookie with SameSite None.">setcookie('sid', 'abc', ['samesite' => 'None', 'secure' => false])</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
