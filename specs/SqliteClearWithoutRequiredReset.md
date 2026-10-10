---
id: SqliteClearWithoutRequiredReset
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "7.1" }
---

# SqliteClearWithoutRequiredReset

## Summary

Reset the SQLite statement before rebinding cleared parameters.

## Detection

- D1. On PHP 5.3–7.1, a SQLite3Stmt with a successful guarded row fetch from its prior execute result is cleared, then rebound with bindValue/bindParam and executed again without reset between the prior execute and rebinding. Report the second execute. Require same local receiver identity and an unambiguous ordered path.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore SqliteClearWithoutRequiredReset and @noinspection SqliteClearWithoutRequiredReset suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Reset the SQLite statement before rebinding cleared parameters.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

PHP resets statements before rebinding during execute starting with 7.2.14 and 7.3.0. Custos selects minor-version targets, so it conservatively excludes all PHP 7.2 and later. Earlier targets require a successful guarded fetch to establish an active cursor; execute alone resets its cursor and is insufficient.

## Examples

```php
<?php
function f(SQLite3Stmt $s) { $r=$s->execute(); if($r->fetchArray()){ $s->clear(); $s->bindValue(1,'later'); <warning descr="Reset the SQLite statement before rebinding cleared parameters.">$s->execute()</warning>; } }
```

Valid case:

```php
<?php
function f(SQLite3Stmt $s) { $s->execute(); $s->reset(); $s->clear(); $s->bindValue(1,'later'); $s->execute(); }
```

## References

- [PHP rebinding version history](https://www.php.net/manual/en/sqlite3stmt.bindparam.php).

- [PHP API contract](https://www.php.net/manual/en/sqlite3stmt.clear.php).

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
