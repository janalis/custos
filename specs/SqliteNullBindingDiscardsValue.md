---
id: SqliteNullBindingDiscardsValue
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SqliteNullBindingDiscardsValue

## Summary

Pass null when using the SQLite null binding type.

## Detection

- D1. SQLite3Stmt bindValue has explicit SQLITE3_NULL type and a proven nonnull scalar value. Unknown values and actual null are excluded; bindParam is excluded because later mutations change its value.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore SqliteNullBindingDiscardsValue and @noinspection SqliteNullBindingDiscardsValue suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Pass null when using the SQLite null binding type.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f(SQLite3Stmt $s) { <warning descr="Pass null when using the SQLite null binding type.">$s->bindValue(1, 'ready', SQLITE3_NULL)</warning>; }
```

Valid case:

```php
<?php
function f(SQLite3Stmt $s) { $s->bindValue(1, null, SQLITE3_NULL); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
