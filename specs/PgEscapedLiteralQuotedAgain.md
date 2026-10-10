---
id: PgEscapedLiteralQuotedAgain
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.4", max: "" }
---

# PgEscapedLiteralQuotedAgain

## Summary

Use the already quoted PostgreSQL literal directly.

## Detection

- D1. pg_escape_literal result is interpolated or concatenated into a constant SQL skeleton with additional single quotes immediately surrounding the entire escaped expression. Parse SQL quote boundaries; report the final query-producing expression. Unknown SQL structure and mixed escaping are excluded.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore PgEscapedLiteralQuotedAgain and @noinspection PgEscapedLiteralQuotedAgain suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Use the already quoted PostgreSQL literal directly.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist. PostgreSQL literal/identifier escaping requires PHP 5.4.4 or later. Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f($c) { $v=pg_escape_literal($c,'Bex'); pg_query($c,<warning descr="Use the already quoted PostgreSQL literal directly.">"SELECT '$v'"</warning>); }
```

Valid case:

```php
<?php
function f($c) { $v=pg_escape_literal($c,'Bex'); pg_query($c,"SELECT $v"); }
```

## References

- [PHP API contract](https://www.php.net/manual/en/function.pg-escape-literal.php).

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
