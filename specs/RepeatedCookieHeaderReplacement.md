---
id: RepeatedCookieHeaderReplacement
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RepeatedCookieHeaderReplacement

## Summary

Detect a second header call with a Set-Cookie literal and default/true replace flag on a path with an earlier Set-Cookie header.

## Detection

- D1. Report a second header call with a Set-Cookie literal and default/true replace flag on a path with an earlier Set-Cookie header.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. setcookie calls, replace false, header_remove clearing the earlier value and different response lifetimes are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Append distinct cookie headers instead of replacing them.`

## Fix

Set the second call replace argument to false only when both cookie names are different and no existing replacement policy is expressed.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
header('Set-Cookie: a=1'); <warning descr="Append distinct cookie headers instead of replacing them.">header('Set-Cookie: b=2')</warning>;
```

```php
<?php
header('Set-Cookie: a=1'); header('Set-Cookie: b=2', false);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
