---
id: ArrayMapMissingCallbackReturn
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# ArrayMapMissingCallbackReturn

## Summary

Detect array_map with a locally defined callback whose reachable normal exits all return no value and whose body evaluates a discarded expression with a non-void result.

## Detection

- D1. Report array_map with a locally defined callback whose reachable normal exits all return no value and whose body evaluates a discarded expression with a non-void result.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- Ignore writes or transformations after unconditional termination or inside a literal false if branch; these statements cannot establish a missing update or mapped value.

- An explicit void return declaration establishes that the callback intentionally supplies no mapped value.

- E1. Explicit return null, any value-returning path, unknown callbacks, and purely side-effect callbacks are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Return the mapped value from this callback.`

## Fix

An automatic return insertion requires no declared return type; the replacement must not violate a callback return contract. In particular, [PHP prohibits returning an expression from void functions](https://www.php.net/manual/en/language.types.void.php).

Add return before the sole discarded expression only for a one-expression callback with no other statements; retain expression text and comments.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<warning descr="Return the mapped value from this callback.">array_map(function($v){trim($v);}, $names)</warning>;
```

```php
<?php
array_map(function($v){return trim($v);}, $names);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
