---
id: PdoPlaceholderBindingMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoPlaceholderBindingMismatch

## Summary

Detect execute on a resolved PDOStatement whose literal prepared SQL requires markers absent from the complete effective binding set, or whose binding set has extra markers.

## Detection

- D1. Report execute on a resolved PDOStatement whose literal prepared SQL requires markers absent from the complete effective binding set, or whose binding set has extra markers.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Unknown SQL, unknown parameter spreads, unknown earlier bindings and driver-specific unresolved statements are excluded; parse quoted literals and comments before counting markers.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. When `execute()` receives no parameter array and the statement variable is captured by a closure (`use ($s)`, by value or by reference) or referenced inside an arrow function written before the call, that callable may bind markers out of sight, so the earlier bindings are unknown and nothing is reported. A callable written after the call does not hide the mismatch, and an explicit parameter array is still checked because it alone supplies the values.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Match SQL placeholders to the effective parameter bindings.`

## Fix

No automatic fix: missing parameter values and extra-binding intent cannot be inferred.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); <error descr="Match SQL placeholders to the effective parameter bindings.">$s->execute(['other' => 2])</error>; }
```

```php
<?php
function fine(PDO $pdo) {
    $s = $pdo->prepare('SELECT :a, :b');
    $rest = function () use ($s) { $s->bindValue(':b', 2); };
    $s->bindValue(':a', 1);
    $rest();
    $s->execute(); // not reported: the closure may have bound :b
}
```

Executable MySQL/MariaDB comments and double-dash sequences without following
whitespace depend on the database dialect. Without a resolved dialect, their
placeholder count is unknown and does not establish a mismatch.

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
