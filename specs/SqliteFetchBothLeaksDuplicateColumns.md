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
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

Resolve inserted SQLITE3_ASSOC to the global builtin constant. Parentheses around the fetch do not bypass consumer checks. Captures into nested scopes and exhausted consumer budgets withhold the finding.

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore SqliteFetchBothLeaksDuplicateColumns and @noinspection SqliteFetchBothLeaksDuplicateColumns suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Fetch named SQLite columns before JSON serialization.`

## Fix

Replace the explicit resolved SQLITE3_BOTH constant with SQLITE3_ASSOC, or append SQLITE3_ASSOC to an argument-free fetchArray. Require the named-column policy and no other consumers.

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

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
