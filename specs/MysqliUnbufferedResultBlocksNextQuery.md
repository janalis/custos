---
id: MysqliUnbufferedResultBlocksNextQuery
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# MysqliUnbufferedResultBlocksNextQuery

## Summary

An unbuffered mysqli result can block further queries on the same connection. Consume or free it before issuing another query.

## Detection

- D1. Resolved mysqli::query uses MYSQLI_USE_RESULT; same connection issues next query before returned result is freed or fully consumed by a proven exhaustion loop. Require direct result ownership and no intervening unknown escape.
- D1d. The unbuffered query must be a known literal result-producing SELECT, SHOW, DESCRIBE or EXPLAIN statement. Mutating statements that return only a success boolean, unknown SQL, and unsupported statement kinds do not establish an outstanding result set.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Consume or free the unbuffered result before another query.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$db=new mysqli();$r=$db->query("SELECT id FROM users",MYSQLI_USE_RESULT);<warning descr="Consume or free the unbuffered result before another query.">$db->query("SELECT 1")</warning>;
```

Valid case:

```php
<?php
$db=new mysqli();$r=$db->query("SELECT id FROM users",MYSQLI_USE_RESULT);$r->free();$db->query("SELECT 1");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
