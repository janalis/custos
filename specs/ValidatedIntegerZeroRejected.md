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
- E2. An options array whose `options` entry carries a `default` key: a rejected value then yields that default instead of false, so a strict false comparison could never observe the failure. No report and no fix.
- E3. Flags containing FILTER_REQUIRE_ARRAY or FILTER_FORCE_ARRAY, given either as the integer third argument or as the `flags` entry of the options array: the call returns an array, negation is an emptiness test and the zero reasoning does not apply. No report and no fix.
- E4. Flags, options or bounds that are not statically known (variables, spreads, non-literal keys or values) leave the failure value unproven and produce no report.
- E5. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Compare integer validation failure strictly with false.`

## Fix

Replace only the negation token and append the strict false comparison. Retain existing parentheses and comments between the negation and validation call.

Replace direct negation with a strict false comparison, preserving one evaluation; the rule stays silent, and so offers no fix, whenever flags or options make the failure value something other than false (null on failure, a configured default) or make the result an array.

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

Not reported, because failure is not reported as false or the result is not an integer:

```php
<?php
if (!filter_var($page, FILTER_VALIDATE_INT, ['options' => ['default' => 1]])) { return; }
if (!filter_var($ids, FILTER_VALIDATE_INT, FILTER_REQUIRE_ARRAY)) { return; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
