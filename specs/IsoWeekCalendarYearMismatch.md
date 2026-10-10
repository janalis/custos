---
id: IsoWeekCalendarYearMismatch
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# IsoWeekCalendarYearMismatch

## Summary

Detect a date format literal containing unescaped Y and W in the same year-week label.

## Detection

- D1. Report a date format literal containing unescaped Y and W in the same year-week label.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Escaped tokens, ISO week-year o, and formats also expressing a complete calendar date are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Use the ISO week year with the ISO week number.`

## Fix

Replace the unescaped Y token with o in the literal when it solely labels the ISO week; preserve all other format bytes.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$day=new DateTimeImmutable(); $label=<warning descr="Use the ISO week year with the ISO week number.">$day->format('Y-W')</warning>;
```

```php
<?php
$day=new DateTimeImmutable(); $label=$day->format('o-W');
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
