---
id: SqliteFetchBothLeaksDuplicateColumns
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SqliteFetchBothLeaksDuplicateColumns

## Summary

Fetch named SQLite columns before JSON serialization. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. SQLite3Result fetchArray explicitly uses SQLITE3_BOTH or omits its mode, and its unchanged result flows to json_encode. Enabling establishes named-column JSON policy. Numeric indexing or mixed-key consumers before serialization invalidate proof.
- D1a. When the fetched row is stored in a plain local variable, json_encode must be its only consumer. Every other mention of that variable in the scope invalidates the proof, including element writes, unset, destructuring, interpolation and reads after such a write, because a modified row is still the fetched row. A mention is unrelated only when it is a complete plain overwrite, when it sits textually before the fetch in code that cannot repeat, or when local flow proves it holds another complete value. Inside a loop, every other mention counts, wherever it appears.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

Resolve inserted SQLITE3_ASSOC to the global builtin constant. Parentheses around the fetch do not bypass consumer checks. Captures into nested scopes and exhausted consumer budgets withhold the finding.

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E1a. Exclude scopes that can reach the row without naming it: variable-variables, compact, extract, get_defined_vars, include or require, eval and goto in the same scope, and `$GLOBALS` when the row is a file-level variable. Exclude a row variable that is a by-reference parameter, a by-reference closure capture, or declared global or static, and a fetch stored through a compound or reference assignment.
- E1b. Exclude a fetchArray written as a first-class callable, which yields a Closure rather than a row, and a fetchArray whose arguments contain a spread, whose mode is unknown.
- E2. Both @custos-ignore SqliteFetchBothLeaksDuplicateColumns and @noinspection SqliteFetchBothLeaksDuplicateColumns suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Fetch named SQLite columns before JSON serialization.`

## Fix

Replace the explicit resolved SQLITE3_BOTH constant with SQLITE3_ASSOC, or append SQLITE3_ASSOC to an argument-free fetchArray. Require the named-column policy and no other consumers as defined by D1a, E1a and E1b; a report is never emitted without that proof, so no fix can remove a numeric key that later code reads.

Use exact byte edits and preserve comments, evaluation count, argument order and supported PHP syntax. Withhold the fix whenever its prerequisites cannot be proven.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f(SQLite3Result $r) { $row=<warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray(SQLITE3_BOTH)</warning>; echo json_encode($row); }
```

Fixed result:

```php
<?php
function f(SQLite3Result $r) { $row=$r->fetchArray(SQLITE3_ASSOC); echo json_encode($row); }
```

Valid case:

```php
<?php
function f(SQLite3Result $r) { $row=$r->fetchArray(SQLITE3_ASSOC); echo json_encode($row); }
```

No report, the numeric key is read after an element write:

```php
<?php
function g(SQLite3Result $r) { $row=$r->fetchArray(); echo json_encode($row); $row['seen']=true; return $row[0]; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
