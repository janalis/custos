---
id: ArrayFilterDropsZero
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# ArrayFilterDropsZero

## Summary

Detect builtin array_filter with its callback omitted or explicitly null, when the effective entries of a literal array contain integer zero or string zero and another nonempty scalar value. Later entries replacing the same key determine the effective value.

## Detection

- D1. Report builtin array_filter with its callback omitted or explicitly null, when the effective entries of a literal array contain integer zero or string zero and another nonempty scalar value. Later entries replacing the same key determine the effective value.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Do not infer intent from an unknown input array; an explicit non-null predicate suppresses the report. An explicit null callback retains PHP’s default filtering behavior.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Preserve zero values with an explicit filter predicate.`

## Fix

Use an explicit predicate preserving valid zero values. No automatic fix: the rejected-value policy is unknown.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<warning descr="Preserve zero values with an explicit filter predicate.">array_filter([0, 2])</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
