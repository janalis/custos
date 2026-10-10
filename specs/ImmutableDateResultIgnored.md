---
id: ImmutableDateResultIgnored
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "5.5", max: "" }
---

# ImmutableDateResultIgnored

## Summary

Detect discarded return from modify, add, sub, setDate, setTime, setTimestamp or setTimezone on a receiver proven DateTimeImmutable.

## Detection

- D1. Report discarded return from modify, add, sub, setDate, setTime, setTimestamp or setTimezone on a receiver proven DateTimeImmutable.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Mutable DateTime, unknown receivers, subclasses overriding the method contract and deliberately returned/chained results are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Assign the returned immutable date.`

## Fix

Assign the result back to a simple local receiver only when subsequent reads establish that updated receiver is intended and no aliases escape; preserve the original call.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 5.5. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$day=new DateTimeImmutable('2025-03-10'); <warning descr="Assign the returned immutable date.">$day->modify('+1 day')</warning>; echo $day->format('c');
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
