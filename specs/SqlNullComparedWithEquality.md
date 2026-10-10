---
id: SqlNullComparedWithEquality
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SqlNullComparedWithEquality

## Summary

SQL equality comparisons do not test whether a value is null. Use IS NULL or IS NOT NULL instead.

## Detection

- D1. Resolved PDO query/exec/prepare or mysqli query/prepare receives literal SQL containing unquoted = NULL, <> NULL, != NULL or NULL = expression inside a WHERE, HAVING or JOIN ON predicate, outside comments. Exclude UPDATE SET assignments, INSERT values, column defaults and expressions outside a proven predicate context. Highlight entire SQL argument. SQL lexer distinguishes strings/quoted identifiers/comments.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use IS NULL or IS NOT NULL for SQL null checks.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
function bad(PDO $p){$p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT * FROM t WHERE x = NULL"</warning>);$p->exec(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"UPDATE t SET x=1 WHERE NULL = y"</warning>);}
```

Valid case:

```php
<?php
$p=new PDO($dsn);$p->query("SELECT * FROM users WHERE deleted_at IS NULL");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
