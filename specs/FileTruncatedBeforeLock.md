---
id: FileTruncatedBeforeLock
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FileTruncatedBeforeLock

## Summary

Detect fopen in w/w+ variants followed by exclusive flock on the same identity.

## Detection

- D1. Report fopen in w/w+ variants followed by exclusive flock on the same identity.
  The lock is exclusive only when the `flock` operation (second argument, or named `operation`) is proven to be `LOCK_EX`, optionally combined with `LOCK_NB` through bitwise or. The operation may be written with the constants, with their integer values, or through a local holding such a value.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Nontruncating modes, already-held established locks, unrelated handles and unknown modes are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A later `flock` that only requests a shared lock (`LOCK_SH`) or releases one (`LOCK_UN`), with or without `LOCK_NB`, is not an exclusive lock and produces no report. The same holds when the operation is missing, dynamic or only partly known.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Acquire the lock before truncating the file.`

## Fix

No automatic fix: changing mode also requires correctly positioned truncation and failure checks.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $h = <warning descr="Acquire the lock before truncating the file.">fopen($path, 'wb')</warning>; if ($h !== false) { flock($h, LOCK_EX); fclose($h); } }
function unlockOnly($path) { $h = fopen($path, 'w'); fwrite($h, 'x'); flock($h, LOCK_UN); fclose($h); }
function sharedOnly($path) { $h = fopen($path, 'w'); flock($h, LOCK_SH); fclose($h); }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
