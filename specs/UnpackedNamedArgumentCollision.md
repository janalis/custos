---
id: UnpackedNamedArgumentCollision
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "8.0", max: "" }
---

# UnpackedNamedArgumentCollision

## Summary

Detect an explicit or unpacked argument binding a parameter already certainly bound by a literal unpacked array on the same call.

## Detection

- D1. Report an explicit or unpacked argument binding a parameter already certainly bound by a literal unpacked array on the same call.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Unknown spreads, conditional keys, valid disjoint bindings and variadic distinct-name collection are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Bind each argument only once.`

## Fix

No automatic fix: choosing which value to retain is application intent.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.0. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function send($id) {} send(...['id'=>7],<error descr="Bind each argument only once.">id:8</error>);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: duplicate array keys are normalized before argument binding. Unknown key types, negative automatic indices with version-dependent behavior, and maximum integer keys cannot establish a collision and are excluded.
