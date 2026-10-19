---
id: PgAsyncDispatchAssumedQuerySuccess
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PgAsyncDispatchAssumedQuerySuccess

## Summary

Inspect the asynchronous PostgreSQL result status.

## Detection

This is an opt-in policy. Truthiness, previous-state checks, raw wait-status comparisons and dispatch acknowledgements can be intentional; enabling this inspection requests explicit handling of the contract described below.

- D1. A pg_send_query or pg_send_query_params boolean success condition directly returns true, a literal success string, or calls a success-output sink without retrieving pg_get_result and checking pg_result_status first. Initial sinks are return true and echo literal text. Ordinary queue bookkeeping is excluded. The sink is the first statement the condition guards, whatever the branch syntax: a braced body, an alternative-syntax body (`if: ... endif;`) or a single statement without braces are equivalent.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. An empty guarded body, or one whose first statement is anything other than an initial sink of D1, is not reported.
- E3. Both @custos-ignore PgAsyncDispatchAssumedQuerySuccess and @noinspection PgAsyncDispatchAssumedQuerySuccess suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Inspect the asynchronous PostgreSQL result status.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f($c) { if(<warning descr="Inspect the asynchronous PostgreSQL result status.">pg_send_query($c,'SELECT 7')</warning>){ return true; } return false; }
function g($c) { if (<warning descr="Inspect the asynchronous PostgreSQL result status.">pg_send_query($c, 'SELECT 8')</warning>) return true; return false; }
```

Valid case:

```php
<?php
function f($c) { if(!pg_send_query($c,'SELECT 7')){return false;} $r=pg_get_result($c); return $r!==false && pg_result_status($r)===PGSQL_TUPLES_OK; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
