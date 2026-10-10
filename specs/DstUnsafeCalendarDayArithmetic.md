---
id: DstUnsafeCalendarDayArithmetic
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# DstUnsafeCalendarDayArithmetic

## Summary

Detect addition of literal 86400 to a timestamp derived from local midnight in a resolved DST-observing timezone under a daily-calendar scheduling contract.

## Detection

- D1. Report addition of literal 86400 to a timestamp derived from local midnight in a resolved DST-observing timezone under a daily-calendar scheduling contract.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. UTC/fixed-offset zones, elapsed-duration contracts, unknown timezone and calendar modify/add operations are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Use calendar arithmetic for the next local day.`

## Fix

No automatic fix: changing representation from elapsed seconds to calendar dates requires intent.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$day=new DateTimeImmutable('2025-03-30 00:00:00',new DateTimeZone('Europe/Paris')); $next=<warning descr="Use calendar arithmetic for the next local day.">$day->getTimestamp()+86400</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: mutable DateTime instances are accepted only as fresh receiver expressions. Stored mutable dates may have changed date or timezone since construction.
