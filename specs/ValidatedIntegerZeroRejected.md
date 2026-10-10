---
id: ValidatedIntegerZeroRejected
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ValidatedIntegerZeroRejected

## Summary

Detect FILTER_VALIDATE_INT result tested for falsiness when zero is permitted by configured min/max bounds.

## Detection

- D1. Report FILTER_VALIDATE_INT result tested for falsiness when zero is permitted by configured min/max bounds.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. FILTER_NULL_ON_FAILURE, bounds excluding zero, strict false checks, and explicit intended nonzero constraints are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Compare integer validation failure strictly with false.`

## Fix

Replace only the negation token and append the strict false comparison. Retain existing parentheses and comments between the negation and validation call.

Replace direct negation with a strict false comparison, preserving one evaluation; suppress fix when flags or options make failure null.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
if (!<warning descr="Compare integer validation failure strictly with false.">filter_var($input, FILTER_VALIDATE_INT)</warning>) { return false; }
```

```php
<?php
if ((filter_var($input, FILTER_VALIDATE_INT) === false)) { return false; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
