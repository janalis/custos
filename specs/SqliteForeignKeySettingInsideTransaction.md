---
id: SqliteForeignKeySettingInsideTransaction
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SqliteForeignKeySettingInsideTransaction

## Summary

Identify PRAGMA foreign_keys assignment inside a known active transaction on a resolved SQLite3 or PDO sqlite connection. Parse complete literal SQL statements; BEGIN opens, COMMIT/END/ROLLBACK close, and unknown transaction state excludes.

## Detection

- D1. Report an `exec()` call on a resolved SQLite3 or PDO sqlite connection whose literal SQL assigns PRAGMA foreign_keys while a transaction is proven open. Parse complete literal SQL statements; BEGIN opens, COMMIT/END/ROLLBACK close, and unknown transaction state excludes.
- D1a. Track the transaction through every earlier call on the same connection in the lexical scope:
  - `exec()` runs its whole literal script; `query()` and `querySingle()` (and `PDO::query()`) apply a single literal statement.
  - `PDO::beginTransaction()` opens; `commit()`, `rollBack()`, `SQLite3::close()` and `SQLite3::open()` leave no open transaction.
  - Other builtin methods of the connection do not change the state.
- D1b. Only a call that is certain to have run before the PRAGMA can open a transaction: it is a statement of a block enclosing the PRAGMA, or its success guards the PRAGMA. A call that may or may not have run (another `if`/`else` branch, a `try` or `catch` body, a loop body, the right side of `or`/`and`/`??`, a ternary arm) leaves the state unknown when it would change it, and is ignored when its literal SQL holds no transaction statement. A later certain BEGIN makes the state known again. A call is disregarded altogether when its path cannot reach the PRAGMA: its statement is followed, in a block that does not contain the PRAGMA, only by plain expression statements and then a `return`, `throw` or `exit`, with no enclosing `try` or nested function in between (`if ($failed) { $db->exec('ROLLBACK'); return; }`).
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1a. The state is unknown, so nothing is reported, after any of these on the connection: SQL that is not a known literal or cannot be tokenized; a multi-statement string passed to a method other than `exec()`; a user-defined or dynamically named method.
- E1b. No report at all when an earlier `prepare()` on the connection holds SQL that is unknown or contains a transaction statement: the prepared statement can run at any later point.
- E2. Both `@custos-ignore SqliteForeignKeySettingInsideTransaction` and `@noinspection SqliteForeignKeySettingInsideTransaction` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Change foreign-key enforcement before the transaction.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$db = new SQLite3(':memory:'); $db->exec('BEGIN; PRAGMA foreign_keys=ON;');
```

Valid case:

```php
<?php
$db = new SQLite3(':memory:'); $db->exec('PRAGMA foreign_keys=ON; BEGIN;');
```

Valid case, the transaction is closed on every path before the setting changes:

```php
<?php
function migrate() {
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    try {
        $db->exec('CREATE TABLE t (a)');
        $db->exec('COMMIT');
    } catch (Exception $e) {
        $db->exec('ROLLBACK');
    }
    $db->exec('PRAGMA foreign_keys=ON');
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
