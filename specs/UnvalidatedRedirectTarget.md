---
id: UnvalidatedRedirectTarget
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnvalidatedRedirectTarget

## Summary

Detect request-derived data reaching a Location header without a proven permitted-local-target or finite destination policy.

## Detection

- D1. Report request-derived data reaching a Location header without a proven permitted-local-target or finite destination policy.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Literal destinations, configured allowlisted origin/path validation and unknown taint flows are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Validate the redirect target against the permitted destination policy.`

## Fix

No automatic fix: the set of permitted destinations must be supplied by the application.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: ' . $_GET['next'])</error>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
