---
id: TransactionExceptionWithoutRollback
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# TransactionExceptionWithoutRollback

## Summary

Detect a caught exceptional path after an owned PDO transaction begins when the catch continues or returns with the transaction active.

## Detection

- D1. Report a caught exceptional path after an owned PDO transaction begins when the catch continues or returns with the transaction active.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Finally rollback, catch rollback, proven transaction auto-termination, rethrow to established owner and unknown state are excluded.
- E2. A catch-path return whose own expression always ends the transaction is a catch rollback: `commit()` or `rollBack()` on the transaction receiver as the returned expression, inside parentheses or a unary operator, as the condition of a ternary or the subject of a `match`, or in both arms of a full ternary or every arm of a `match`. A call in only one arm, in a short ternary fallback or behind a short-circuit operator is conditional and does not count.
- E3. A finally rollback counts when the enclosing `finally` block calls `commit()` or `rollBack()` on the receiver as a top-level statement, or as a direct statement of a top-level `if` whose whole condition is `inTransaction()` on that same receiver. Other guards and deeper nesting do not count.
- E4. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Roll back the active transaction on the caught failure path.`

## Fix

No automatic fix: safe rollback must account for begin failure and possible earlier termination.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { <warning descr="Roll back the active transaction on the caught failure path.">return false;</warning> } }
```

No report, the return itself or a guarded `finally` ends the transaction:

```php
<?php
function returnsRollback(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { return $pdo->rollBack(); } }
function guardedFinally(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { return false; } finally { if ($pdo->inTransaction()) { $pdo->rollBack(); } } }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
