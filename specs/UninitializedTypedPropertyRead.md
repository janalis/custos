---
id: UninitializedTypedPropertyRead
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.4", max: "" }
---

# UninitializedTypedPropertyRead

## Summary

Detect reading a resolved typed instance property when every incoming path proves it remains uninitialized.

## Detection

- D1. Report reading a resolved typed instance property when every incoming path proves it remains uninitialized.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Nullable properties still require initialization; exclude isset/empty checks, explicit defaults, guarded initialization, unknown constructors and magic access.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Initialize this typed property before reading it.`

## Fix

No automatic fix: an appropriate initial value cannot be inferred.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 7.4. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
class Meter {public int $value;} echo <error descr="Initialize this typed property before reading it.">(new Meter())->value</error>;
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: null-coalescing on the left operand is a guarded read and is excluded. Every ancestor and trait must resolve before absence of constructor or magic initialization can establish an uninitialized fresh instance.
