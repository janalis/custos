---
id: TempnamDirectoryFallbackUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# TempnamDirectoryFallbackUnchecked

## Summary

Detect tempnam result used for sensitive content under an explicitly enabled required-directory contract without verifying its parent directory.

## Detection

- D1. Report tempnam result used for sensitive content under an explicitly enabled required-directory contract without verifying its parent directory.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Verified canonical parent, intentional system-temp fallback, unknown sensitivity and tempnam failure already handled are excluded.
- E3. The canonical parent is verified when `realpath(dirname($path))` is compared with another `realpath(...)` call, in either operand order, after the `tempnam` call that produced `$path`, and the conditions guarding the write prove that comparison matched: an equality (`===`, `==`) known to hold or an inequality (`!==`, `!=`) known to fail where the write runs. That covers the body of an `if` (or the right operand of `&&`) on a match test, code after an `if` on a mismatch test whose body always leaves (return, throw, exit, break, continue), with or without other arms, and either test negated, compared with a boolean, combined through `&&`/`||` or kept in a local. The `else` of a match test is not covered. `$path` must hold the same value at the comparison and at the write, and a jump label between them voids the proof.

  Parentheses around the operands are ignored. The loose operators are accepted because a canonical path is an absolute string or `false`, so numeric-string juggling cannot occur. A comparison of uncanonical values (`dirname($path) !== $dir`) proves nothing: symlinks and relative segments make equal directories differ and different ones match.
- E4. A guard placed before the `tempnam` call that produced the written path verified a different file; the report is kept when the variable is assigned a new temporary file after the guard.
- E5. The fallback is intentional when the directory argument of `tempnam` is a builtin `sys_get_temp_dir()` call, directly or through a variable with that single known value: the system directory is the fallback target, so there is no other location to end up in. A path derived from it (`sys_get_temp_dir() . '/private'`) is an ordinary required directory.
- E6. Test code (a test file, or a method of a class whose name marks it as a test) is not reported: its temporary files are throwaway fixtures, not sensitive content.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Verify the temporary file remains in the required directory.`

## Fix

No automatic fix: fallback rejection and cleanup policy must be supplied explicitly.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($privateDir, $secret) { $path = tempnam($privateDir, 'job'); if ($path === false) { return; } <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>; }
```

```php
<?php
// Not reported: either operand order, loose or strict, exit or enclosing branch.
function swapped($privateDir, $secret) { $path = tempnam($privateDir, 'job'); if ($path === false || realpath($privateDir) != realpath(dirname($path))) { return; } file_put_contents($path, $secret); }
function matching($privateDir, $secret) { $path = tempnam($privateDir, 'job'); if ($path !== false && realpath(dirname($path)) === realpath($privateDir)) { file_put_contents($path, $secret); } }
// Not reported: the system temporary directory was requested.
function scratch($data) { $path = tempnam(sys_get_temp_dir(), 'scratch'); file_put_contents($path, $data); }
// Reported: the guard checked the previous file.
function recreated($privateDir, $secret) { $path = tempnam($privateDir, 'a'); if (realpath(dirname($path)) !== realpath($privateDir)) { return; } $path = tempnam($privateDir, 'b'); <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
