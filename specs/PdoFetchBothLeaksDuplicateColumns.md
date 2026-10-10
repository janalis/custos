---
id: PdoFetchBothLeaksDuplicateColumns
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoFetchBothLeaksDuplicateColumns

## Summary

PDO's combined fetch mode exposes numeric and named copies of each field. Select the intended row representation before JSON serialization. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved PDOStatement::fetchAll explicitly uses PDO::FETCH_BOTH and its result is supplied to resolved json_encode without shape conversion. Flag opt-in because duplicate numeric/name representation can be intended.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Use associative fetch mode before serializing rows.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
function bad(PDOStatement $s){<warning descr="Use associative fetch mode before serializing rows.">json_encode($s->fetchAll(PDO::FETCH_BOTH))</warning>;}
```

Valid case:

```php
<?php
$p=new PDO($dsn);$s=$p->query("SELECT id FROM users");$a=$s->fetchAll(PDO::FETCH_ASSOC);echo json_encode($a);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
