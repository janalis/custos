---
id: SortResultUsedAsArray
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SortResultUsedAsArray

## Summary

Detect the boolean result of sort, rsort, asort, arsort, ksort, krsort, usort, uasort, or uksort when consumed directly or through a proven alias by foreach or an array-only builtin.

## Detection

- D1. Report the boolean result of sort, rsort, asort, arsort, ksort, krsort, usort, uasort, or uksort when consumed directly or through a proven alias by foreach or an array-only builtin.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Success checks and unknown overwritten aliases are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete consuming foreach statement or array/iterator-consuming call.
- Severity: warning.
- Message: `Use the sorted input array instead of the success flag.`

## Fix

No automatic fix: restructuring statements and ownership of the sorted input requires intent.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$sorted = sort($items); <warning descr="Use the sorted input array instead of the success flag.">foreach ($sorted as $item) {}</warning>
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
