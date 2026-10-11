---
id: FileLockFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FileLockFailureUnchecked

## Summary

Detect a write following flock requiring an exclusive lock where acquisition success is not established.

## Detection

- D1. Report a write following flock requiring an exclusive lock where acquisition success is not established.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. Acquisition is established when the write can only run after the last `flock` call on the handle returned true. The result may be tested directly or through a plain local that still holds it unchanged, and the test may be negated, error-suppressed (`@`), or compared strictly or loosely with `true` or `false`. Accepted tests are:
  - an `if`, `elseif` or `while` condition whose guarded body contains the write, or the matching side of a conditional expression;
  - the right operand of `&&`/`and` (result true) or `||`/`or` (result false) containing the write;
  - an `if` whose failure branch always leaves (return, throw, exit, break, continue), with the write later in the statement list that owns the `if`;
  - a statement of the form `flock(...) or exit` / `|| throw`, with the write later in the same statement list;
  - a `while` or `do`-`while` retry loop repeating until the result is true and containing no `break`, with the write after the loop.
- D5. A test proves nothing, and the write is still reported, when the local was reassigned or merged from another value, when the test sits on only one path to the write (a nested conditional, another closure or function, an earlier `case` label), when the failure branch can continue, or when the result is combined with an unrelated condition that can let the failing side through. `assert(flock(...))` is not a test: assertions may be compiled out.

## Exceptions (no report)

- E1. Guarded acquisition and retry until success in any of the D4 forms, explicit unlocked best-effort contracts and unrelated handles are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Acquire the lock successfully before writing.`

## Fix

No automatic fix: lock failure and retry policies are unknown.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $h = fopen($path, 'cb'); if ($h === false) { return; } flock($h, LOCK_EX); <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'next')</warning>; fclose($h); }
```

```php
<?php
function stored($path) {
    $h = fopen($path, 'cb');
    if ($h === false) { return false; }
    $locked = flock($h, LOCK_EX);
    if (!$locked) { fclose($h); return false; }
    fwrite($h, 'next'); // not reported: the stored result was tested
    return fclose($h);
}

function ignored($h) {
    $locked = flock($h, LOCK_EX);
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'next')</warning>;
    return $locked;
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
