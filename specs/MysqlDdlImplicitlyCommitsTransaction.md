---
id: MysqlDdlImplicitlyCommitsTransaction
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# MysqlDdlImplicitlyCommitsTransaction

## Summary

Some MySQL DDL statements implicitly commit an active transaction. Keep them outside operations that depend on rolling back the transaction.

## Detection

- D1. A PDO connection has a literal mysql: DSN, beginTransaction precedes exec of literal CREATE TABLE, ALTER TABLE or DROP TABLE (not TEMPORARY), then rollBack. No commit or unknown connection escape intervenes. Highlight DDL operation.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Keep implicitly committing DDL outside rollback-dependent transactions.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$p=new PDO("mysql:host=localhost");$p->beginTransaction();<warning descr="Keep implicitly committing DDL outside rollback-dependent transactions.">$p->exec("CREATE TABLE t (id INT)")</warning>;$p->rollBack();
```

Valid case:

```php
<?php
$p=new PDO("sqlite::memory:");$p->beginTransaction();$p->exec("CREATE TABLE audit (id INT)");$p->rollBack();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
