---
id: ReadModifyWriteLockTooLate
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ReadModifyWriteLockTooLate

## Summary

Detect file_get_contents of a path feeding a transformed value into file_put_contents on the same proven path with LOCK_EX but no enclosing lock spanning the read.

## Detection

- D1. Report file_get_contents of a path feeding a transformed value into file_put_contents on the same proven path with LOCK_EX but no enclosing lock spanning the read.
- D1a. A lock spans the operation when, in the same function or top-level scope, a builtin `flock` call that ends before the earliest feeding read takes an exclusive lock (`LOCK_EX`, optionally with `LOCK_NB`) on any handle and that handle is not provably released before the write. An operation that cannot be folded to a known value counts as exclusive. `LOCK_SH` does not serialize writers and never counts.
- D1b. A release is `flock` with a known non-exclusive operation (`LOCK_UN`, or a downgrade to `LOCK_SH`) or `fclose` on the same handle expression. It ends the lock only when it is a statement completed before the write in a statement list enclosing the write, or when it sits in the statement list that also contains the acquisition. A release inside a branch that leaves (the failed-lock cleanup, a `finally` block) does not end it.
- D1c. An application (non-builtin) function, method or static call completed before the earliest read may take the lock itself: no report when its callee text contains `lock` without containing `unlock`, or when one of its arguments is the same proven path. Other unknown calls do not silence the report.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Separate destinations, read-only paths, established spanning locks (D1a–D1c), atomic database alternatives and unknown path identity are excluded.
- E1a. A lock taken after the read, a shared lock, and a lock released or closed before the read or between the read and the write are not spanning locks: the report stays.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Hold the lock across the complete read-modify-write operation.`

## Fix

No automatic fix: the complete locked read-modify-write protocol must handle partial reads/writes and failures.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $n = (int) file_get_contents($path); <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>; }
```

```php
<?php
function good($path, $lock) {
    if (!flock($lock, LOCK_EX)) { return; }
    $n = (int) file_get_contents($path);
    file_put_contents($path, $n + 1, LOCK_EX);
    flock($lock, LOCK_UN);
}
function shared($path, $lock) {
    flock($lock, LOCK_SH);
    $n = (int) file_get_contents($path);
    <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>;
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
