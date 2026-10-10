---
id: IteratorMaterializationKeyCollision
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# IteratorMaterializationKeyCollision

## Summary

Detect iterator_to_array with key preservation enabled/default on a resolved generator yielding two certainly equal keys on the same execution path.

## Detection

- D1. Report iterator_to_array with key preservation enabled/default on a resolved generator yielding two certainly equal keys on the same execution path.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Unknown keys, mutually exclusive yields, or preserve_keys false are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Preserve all yielded values when materializing the iterator.`

## Fix

No automatic fix: preserving an associative mapping versus every value is an application decision.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 5.5. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function rows() { yield 'k' => 2; yield 'k' => 5; } $a = <warning descr="Preserve all yielded values when materializing the iterator.">iterator_to_array(rows())</warning>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
