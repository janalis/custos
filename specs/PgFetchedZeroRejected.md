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

- D1. A `pg_fetch_result()` call is the whole condition of an `if` or `while`, either directly or as the right-hand side of a plain `=` assignment to a simple variable (`$v = pg_fetch_result(...)`). The function returns a string, `false` when the fetch fails, or `null` for an SQL NULL column; a truthiness test therefore also rejects the fetched strings `"0"` and `""`. Explicit comparisons, negations and conditions combining other requirements are not reported.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E3. A first-class callable `pg_fetch_result(...)` creates a closure and fetches nothing.
- E4. Assignments that do not hand the fetched value to the condition unchanged: compound operators (`.=`, `??=`, ...), by-reference assignment, and targets other than a simple variable (property, array element, destructuring list, variable variable), where a typed target may coerce `false`.
- E2. Both @custos-ignore PgFetchedZeroRejected and @noinspection PgFetchedZeroRejected suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Distinguish PostgreSQL fetch failure from a zero value.`

## Fix

Wrap the reported condition in `is_string(...)`: the fetch still runs exactly once, `"0"` and `""` are now accepted, and both `false` (failed fetch) and `null` (SQL NULL) stay rejected as they were by the truthiness test. A bare `!== false` comparison is not offered because it would let SQL NULL into the branch. `is_string` is written `\is_string` when the unqualified name would resolve to a namespaced or imported function.

The fix only inserts text before and after the condition, so comments, arguments and their order are preserved byte for byte.

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
function f($r) { if(is_string($v=pg_fetch_result($r,0,0))){echo $v;} }
```

Valid cases:

```php
<?php
function f($r) { if(($v=pg_fetch_result($r,0,0))!==false){echo $v;} }
function g($r) { $v = ''; if ($v .= pg_fetch_result($r, 0, 0)) { echo $v; } }
function h() { if ($fetch = pg_fetch_result(...)) { return $fetch; } }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
