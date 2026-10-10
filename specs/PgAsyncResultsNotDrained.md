---
id: PgAsyncResultsNotDrained
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PgAsyncResultsNotDrained

## Summary

Drain all pending PostgreSQL results before another asynchronous query.

## Detection

- D1. On the same proven connection, a successful pg_send_query of a constant SQL string with at least two statements is followed by fewer fixed pg_get_result calls than statements and another pg_send_query. Also report a second send with no retrieval after any first successful send. Recognize loops until pg_get_result is false as draining; unknown SQL or calls invalidate count proof.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore PgAsyncResultsNotDrained and @noinspection PgAsyncResultsNotDrained suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Drain all pending PostgreSQL results before another asynchronous query.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f($c) { if(pg_send_query($c,'SELECT 7; SELECT 8')){ pg_get_result($c); <warning descr="Drain all pending PostgreSQL results before another asynchronous query.">pg_send_query($c,'SELECT 9')</warning>; } }
```

Valid case:

```php
<?php
function f($c) { if(pg_send_query($c,'SELECT 7; SELECT 8')){ while(pg_get_result($c)!==false){} pg_send_query($c,'SELECT 9'); } }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
