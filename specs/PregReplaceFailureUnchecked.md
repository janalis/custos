---
id: PregReplaceFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PregReplaceFailureUnchecked

## Summary

Detect a nullable result of preg_replace or preg_replace_callback flowing into a non-null string parameter or dereference without a successful null guard.

## Detection

A sentinel or length guard applies only to the same reaching value. Reassignment or reevaluating an identical-looking call invalidates the earlier guard.

- D1. Report a nullable result of preg_replace or preg_replace_callback flowing into a non-null string parameter or dereference without a successful null guard.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Explicit null fallback, guard, termination, and known non-PCRE callees are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Handle regex replacement failure before using its result.`

## Fix

No automatic fix: the correct handling of regex failure is unknown.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$text = <warning descr="Handle regex replacement failure before using its result.">strtoupper(preg_replace('/z/u', '', $input))</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
