---
id: NestedPdoTransactionWithoutSavepoint
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# NestedPdoTransactionWithoutSavepoint

## Summary

PDO does not start nested transactions by calling beginTransaction again. Finish the current transaction or use an appropriate savepoint abstraction.

## Detection

- D1. Same proven PDO receiver calls beginTransaction twice in straight-line scope with no commit/rollBack between. Highlight second call. Unknown intervening calls discard proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Finish the transaction before beginning another one.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$pdo=new PDO("sqlite::memory:");$pdo->beginTransaction();<warning descr="Finish the transaction before beginning another one.">$pdo->beginTransaction()</warning>;
```

Valid case:

```php
<?php
$p=new PDO($dsn);$p->beginTransaction();$p->commit();$p->beginTransaction();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
