---
id: PgIdentifierEscapedAsLiteral
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.4", max: "" }
---

# PgIdentifierEscapedAsLiteral

## Summary

Escape PostgreSQL identifiers with the identifier API.

## Detection

- D1. pg_escape_literal result occupies a proven single table identifier after FROM, JOIN, UPDATE, INTO or TABLE in a constant SQL skeleton passed to pg_query. Exclude values, expression positions, qualified identifier assembly and ambiguous SQL.
  - The keyword is read inside the innermost parenthesis group still open at the operand. An open group counts only when it is a subquery (it starts with SELECT or WITH); any other open group is a function call or an expression list, where FROM introduces a value (`substring(a FROM $x)`, `extract(year FROM $x)`, `trim(both FROM $x)`).
  - FROM directly preceded by DISTINCT is the comparison `IS [NOT] DISTINCT FROM` and takes a value.
  - A statement starting with COPY is never a table position for this rule: `COPY t FROM $x` names a file.
  - Unbalanced closing parentheses before the operand make the SQL ambiguous.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

Parentheses around the producer do not bypass consumer checks. Captures into nested scopes and exhausted consumer budgets withhold the fix, while retaining the finding at the proven identifier use.

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore PgIdentifierEscapedAsLiteral and @noinspection PgIdentifierEscapedAsLiteral suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Escape PostgreSQL identifiers with the identifier API.`

## Fix

Replace only the builtin function name pg_escape_literal with pg_escape_identifier when every use of the local result is a proven identifier position. Preserve leading qualification and do not change arguments.

Uses are not only those written after the call. When the call sits in a loop, a read of the assigned variable placed earlier in that loop receives the result on the next iteration, so any such read withholds the fix; reads before the loop do not. A `goto` in the same scope withholds the fix as well, and so does a call to `compact` or `get_defined_vars`, which read locals without naming them. A later read that several definitions can reach (after a branch or a loop that reassigns the variable) may still carry the result, so it withholds the fix too; a plain overwrite of the variable is not a use.

Use exact byte edits and preserve comments, evaluation count, argument order and supported PHP syntax. Withhold the fix whenever its prerequisites cannot be proven.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist. PostgreSQL literal/identifier escaping requires PHP 5.4.4 or later. Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f($c) { $t=<warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($c,'inventory')</warning>; pg_query($c,"SELECT * FROM $t"); }
```

Fixed result:

```php
<?php
function f($c) { $t=pg_escape_identifier($c,'inventory'); pg_query($c,"SELECT * FROM $t"); }
```

Valid case:

```php
<?php
function f($c) { $t=pg_escape_identifier($c,'inventory'); pg_query($c,"SELECT * FROM $t"); }
```

## References

- [PHP API contract](https://www.php.net/manual/en/function.pg-escape-identifier.php).

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
