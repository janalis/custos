---
id: TransactionEarlyReturn
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# TransactionEarlyReturn

## Summary

Detect a reachable normal return from a scope owning a successfully begun PDO transaction where neither commit nor rollback occurs on the exit path.

## Detection

- D1. Report a reachable normal return from a scope owning a successfully begun PDO transaction where neither commit nor rollback occurs on the exit path.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Ownership transfer, exception unwinding into proven cleanup, surrounding finally cleanup, failure to begin and unknown transaction state are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A return whose own expression closes the transaction is not reported: a `commit()` or `rollBack()` call on the connection that began it must be evaluated every time the expression is, either unconditionally (directly, parenthesised, under a unary operator or cast, as an assigned value, or as an operand that is always evaluated) or in both branches of a full ternary or in every arm of a `match`. A call that is only the right operand of `&&`, `||`, `and`, `or`, `??` or `??=`, or one branch of a short ternary, may be skipped and proves nothing, so `return $saved && $pdo->commit();` is still reported.
- E4. Surrounding `finally` cleanup (E1) covers a direct `commit()`/`rollBack()` statement in the `finally` block and the guarded idiom `if ($pdo->inTransaction()) { $pdo->rollBack(); }` on the same connection, where the closing call is a direct statement of the guard body. A closing call under any other condition does not prove cleanup. A return written inside the `finally` block itself is not covered by the guarded idiom.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Close the owned transaction before returning.`

## Fix

No automatic fix: choose commit or rollback explicitly according to the transaction contract.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } if ($skip) { <warning descr="Close the owned transaction before returning.">return;</warning> } $pdo->commit(); }
```

```php
<?php
// Not reported: every exit path closes the transaction.
function commitInReturn(PDO $pdo) { if (!$pdo->beginTransaction()) { return false; } return $pdo->commit(); }
function guardedCleanup(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { return; } $pdo->commit(); } finally { if ($pdo->inTransaction()) { $pdo->rollBack(); } } }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
