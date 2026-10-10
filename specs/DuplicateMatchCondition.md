---
id: DuplicateMatchCondition
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# DuplicateMatchCondition

## Summary

Detect a later match arm condition statically strictly identical to an earlier condition, including resolved references to the same enum case.

## Detection

- D1. Report a later match arm condition statically strictly identical to an earlier condition, including resolved references to the same enum case.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Unknown values, different strict types and effectful condition expressions are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Remove the repeated match condition.`

## Fix

Remove a duplicate condition only when its whole arm becomes unreachable and its result has no comments requiring preservation; otherwise offer no fix.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.0. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$v=match($input) {4=>'first',<warning descr="Remove the repeated match condition.">4</warning>=>'second',default=>'other'};
```

```php
<?php
$v=match($input) {4=>'first',default=>'other'};
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: arm removal is withheld when its byte range contains a PHP hash comment, as well as slash comments. The diagnostic remains available without a fix.
