---
id: NonExhaustiveEnumMatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# NonExhaustiveEnumMatch

## Summary

Detect match over a resolved enum when reachable enum cases are missing and there is no default arm.

## Detection

- D1. Report match over a resolved enum when reachable enum cases are missing and there is no default arm.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Exhaustive matches, flow-narrowed case sets, default arms and unknown subject types are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Handle every reachable enum case.`

## Fix

No automatic fix: missing result expressions cannot be invented.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.1. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
enum Choice {case Left;case Right;} function label(Choice $c) {return <error descr="Handle every reachable enum case.">match($c) {Choice::Left=>'L'}</error>;}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: a single enum case established by a dominating local assignment narrows the required case set. Other enum cases are unreachable for that subject and are not reported.
