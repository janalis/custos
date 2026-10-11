---
id: StreamOpenFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# StreamOpenFailureUnchecked

## Summary

Detect fopen result passed to a stream-only operation without excluding false.

## Detection

- D1. Report fopen result passed to a stream-only operation without excluding false.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Strict/truthy success guards, terminating error branches, proven wrappers and overwritten identities are excluded.
- E1a. Because fopen yields a stream resource or false, any condition on the same reaching handle that separates the two outcomes is a guard, not only a strict comparison. Recognised conditions: the bare variable (truthiness), `empty()`, `is_resource()`, `is_bool()`, strict or loose comparison of any of these with a `true`/`false` literal on either side, their negation with `!`, a conjunct of `&&`/`and` that holds, a disjunct of `||`/`or` that fails, and a plain assignment of the fopen call itself used as the condition (`if ($h = fopen(...))`, `if (($h = fopen(...)) !== false)`).
- E1b. A guard applies when it dominates the use: the use is in the branch of an `if`/`elseif`/`else` or ternary the condition selects, in the body of a `while` loop whose condition is the guard, in the right operand of a short-circuit operator whose left operand is the condition, or after an earlier statement of the same statement list that is either an `if` whose body always leaves (return, throw, exit, break, continue, goto) or an expression statement that only completes on success: `$h = fopen(...) or die();`, `$h === false and throw ...;`, `$h ?: throw ...;`. The guarded variable must still hold the same fopen result at the use; a reassignment between guard and use voids the guard.
- E1c. Not treated as guards: `assert()` (it can be compiled out), a comparison stored in another variable, `switch (true)`/`match (true)` arms, an error branch that only calls a function or method without leaving syntactically, and a preceding stream call on the same handle.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Check that the stream opened successfully.`

## Fix

No automatic fix: choosing fallback versus failure belongs to the caller.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $h = fopen($path, 'rb'); <warning descr="Check that the stream opened successfully.">fread($h, 12)</warning>; }
```

```php
<?php
function header_bytes($path) {
    $h = fopen($path, 'rb');
    if (!$h) {
        throw new RuntimeException('cannot open ' . $path);
    }
    return fread($h, 12); // not reported: the falsy branch always leaves
}
function first_line($path) {
    if ($h = fopen($path, 'rb')) {
        return fgets($h); // not reported: the assignment is the condition
    }
    return null;
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
