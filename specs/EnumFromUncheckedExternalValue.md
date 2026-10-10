---
id: EnumFromUncheckedExternalValue
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# EnumFromUncheckedExternalValue

## Summary

Detect a resolved backed enum from call with a request-derived backing value not proven within its declared case set and without handling ValueError.

## Detection

- D1. Report a resolved backed enum from call with a request-derived backing value not proven within its declared case set and without handling ValueError.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Known valid backing constants, tryFrom, validation establishing membership, and a catching error handler are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Handle unknown enum backing values explicitly.`

## Fix

No automatic fix: choosing null fallback versus throwing is a caller contract.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.1. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
enum Phase:string {case Ready='ready';} $p=<warning descr="Handle unknown enum backing values explicitly.">Phase::from($_GET['phase'])</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: an immediate first use in the positive body of a resolved strict in_array check is accepted when its literal allowed values match declared backing values. Unknown sets, non-strict comparison, unrelated inputs and intervening evaluation do not prove membership.
