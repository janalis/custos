---
id: CallbackRequiredArgumentMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CallbackRequiredArgumentMismatch

## Summary

Detect array_map with N supplied input arrays and a resolved callback requiring more than N positional parameters.

## Detection

- D1. Report array_map with N supplied input arrays and a resolved callback requiring more than N positional parameters.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. A string callback naming a project function is resolved only when that fully qualified function name has exactly one declaration in the index, and its body is a direct statement of the analysed file or of a namespace. Such a declaration is the signature every call reaches.

## Exceptions (no report)

- E1. Optional parameters, variadics, unresolved callbacks, and array_map with null callback are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A string callback whose function is declared inside a condition, loop, function body or any other nested statement, or its fully qualified name has multiple indexed declarations (for example one signature per `if`/`else` branch, or a `function_exists` guarded fallback), is unresolved: the signature bound at run time is not known, so nothing is reported.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Match the callback parameters to the supplied arrays.`

## Fix

No automatic fix: an additional input array cannot be invented.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Match the callback parameters to the supplied arrays.">array_map(fn($value, $key) => $key . $value, $items)</error>;

function join_pair($left, $right) { return $left . $right; }
<error descr="Match the callback parameters to the supplied arrays.">array_map('join_pair', $items)</error>;

// Not reported: which declaration runs depends on the branch taken.
if (PHP_VERSION_ID < 80000) {
    function label($item, $prefix) { return $prefix . $item; }
} else {
    function label($item) { return (string) $item; }
}
array_map('label', $items);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
