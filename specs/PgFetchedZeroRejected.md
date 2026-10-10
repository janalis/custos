---
id: PgFetchedZeroRejected
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PgFetchedZeroRejected

## Summary

Distinguish PostgreSQL fetch failure from a zero value.

## Detection

- D1. pg_fetch_result is used as a positive truthiness condition directly or through an assignment in if/while. Its result is a string-or-false contract, so string zero is rejected. Exclude explicit intentional comparisons and guards combining other requirements.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore PgFetchedZeroRejected and @noinspection PgFetchedZeroRejected suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Distinguish PostgreSQL fetch failure from a zero value.`

## Fix

Wrap the existing assignment/call in parentheses and compare !== false for positive if/while conditions only. Withhold for negated, compound, ternary or boolean-arithmetic uses; retain the original evaluation once.

Use exact byte edits and preserve comments, evaluation count, argument order and supported PHP syntax. Withhold the fix whenever its prerequisites cannot be proven.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f($r) { if(<warning descr="Distinguish PostgreSQL fetch failure from a zero value.">$v=pg_fetch_result($r,0,0)</warning>){echo $v;} }
```

Fixed result:

```php
<?php
function f($r) { if(($v=pg_fetch_result($r,0,0))!==false){echo $v;} }
```

Valid case:

```php
<?php
function f($r) { if(($v=pg_fetch_result($r,0,0))!==false){echo $v;} }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
