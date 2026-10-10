---
id: ExhaustedGeneratorReused
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# ExhaustedGeneratorReused

## Summary

Detect a resolved Generator identity traversed again after a proven complete iterator_to_array call or complete foreach.

## Detection

- D1. Report a resolved Generator identity traversed again after a proven complete iterator_to_array call or complete foreach.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Early breaks, potentially exceptional completion, IteratorAggregate values, and creation of a fresh generator invalidate exhaustion proof.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete consuming foreach statement or array/iterator-consuming call.
- Severity: error.
- Message: `Create a fresh generator before traversing again.`

## Fix

No automatic fix: recreation and materialization have different side effects and memory costs.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 5.5. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function values() { yield 1; } $g = values(); iterator_to_array($g); <error descr="Create a fresh generator before traversing again.">foreach ($g as $v) {}</error>
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
