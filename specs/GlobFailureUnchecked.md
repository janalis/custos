---
id: GlobFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GlobFailureUnchecked

## Summary

Detect glob result directly or through a proven alias consumed by foreach or array-only operations without excluding false.

## Detection

- D1. Report glob result directly or through a proven alias consumed by foreach or array-only operations without excluding false.
- D1a. Array-only operations are calls to these builtins with the glob result as their array argument, positionally or by name: `count`, `sizeof`, `array_map`, `array_filter`, `array_values`, `array_keys`, `array_reverse`, `array_slice`, `array_unique`, `array_flip`, `array_sum`, `array_product`, `array_column`, `array_chunk`, `array_key_first`, `array_key_last`, `array_is_list`, `array_rand`, `array_key_exists`, `in_array`, `array_search`, `implode`, `join`, `array_combine`, `array_fill_keys`, and the first array of `array_merge`, `array_diff` and `array_intersect`. They throw a `TypeError` for false since PHP 8.0 and warn with a meaningless result before. Each unguarded use is reported on its own.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Strict failure guards, explicit fallback and unknown wrapper contracts are excluded.
- E1a. Because glob yields an array or false, any condition on the same reaching value that separates the two outcomes is a guard, not only a strict comparison. Recognised conditions: the bare variable (truthiness), `empty()`, `isset()` of one of its offsets, `is_array()`, `is_iterable()`, `is_countable()`, `is_bool()`, strict or loose comparison of any of these with a `true`/`false` literal, their negation with `!`, a conjunct of `&&`/`and` that holds, and a disjunct of `||`/`or` that fails. `isset()` of the variable itself only excludes null and is not a guard.
- E1b. A guard applies when it dominates the use: the use is in the branch of an `if`/`elseif`/`else` or ternary the condition selects, in the right operand of a short-circuit operator whose left operand is the condition, or after an earlier statement of the same statement list that is either an `if` whose body always leaves (return, throw, exit, break, continue) or a `condition or exit`/`condition and throw` style expression statement. The guarded variable must still hold the same glob result at the use; a reassignment between guard and use voids the guard.
- E1c. Not treated as guards: `assert()`, a comparison stored in another variable, loop conditions and a preceding array-only call on the same value. `implode`/`join` called with the array as their only argument are not inspected.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Handle glob failure before iterating its result.` for foreach; `Handle glob failure before using its result as an array.` for an array-only call (range: the argument).

## Fix

No automatic fix: failure and an empty match set must remain distinguishable unless caller policy establishes otherwise.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($pattern) { foreach (<warning descr="Handle glob failure before iterating its result.">glob($pattern)</warning> as $path) { echo $path; } }
function names($pattern) {
    $paths = glob($pattern);
    return array_map('basename', <warning descr="Handle glob failure before using its result as an array.">$paths</warning>);
}
function guarded($pattern) {
    $paths = glob($pattern);
    if (!$paths) {
        return 0;
    }
    return count($paths);
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
