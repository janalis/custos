---
id: SqliteResultUsedAfterFinalize
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SqliteResultUsedAfterFinalize

## Summary

Read the SQLite result before finalizing it.

## Detection

- D1. After successful SQLite3Result finalize on a proven local receiver, fetchArray, reset, columnName, columnType or numColumns is called on that same result without reassignment. Report the invalid use. If finalize success is unknown, require an explicit true guard or unconditional lifecycle state guaranteed by the selected runtime contract.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore SqliteResultUsedAfterFinalize and @noinspection SqliteResultUsedAfterFinalize suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Read the SQLite result before finalizing it.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f(SQLite3Result $r) { if($r->finalize()){ <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; } }
```

Valid case:

```php
<?php
function f(SQLite3Result $r) { $row=$r->fetchArray(); $r->finalize(); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
