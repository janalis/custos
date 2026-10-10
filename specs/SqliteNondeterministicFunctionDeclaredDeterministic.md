---
id: SqliteNondeterministicFunctionDeclaredDeterministic
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.1", max: "" }
---

# SqliteNondeterministicFunctionDeclaredDeterministic

## Summary

Identify SQLite3::createFunction with SQLITE3_DETERMINISTIC when inline callback directly returns random_int, random_bytes, time or microtime, without composition. Unknown callback behavior cannot prove nondeterminism.

## Detection

- D1. Report SQLite3::createFunction with SQLITE3_DETERMINISTIC when inline callback directly returns random_int, random_bytes, time or microtime, without composition. Unknown callback behavior cannot prove nondeterminism.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- Equal proven bounds make random_int return that bound. This does not establish a nondeterministic callback.

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore SqliteNondeterministicFunctionDeclaredDeterministic` and `@noinspection SqliteNondeterministicFunctionDeclaredDeterministic` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Remove the deterministic flag from this callback.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 7.1 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$db = new SQLite3(':memory:'); $db->createFunction('dice', fn() => random_int(1, 6), 0, SQLITE3_DETERMINISTIC);
```

Valid case:

```php
<?php
$db = new SQLite3(':memory:'); $db->createFunction('twice', fn($n) => $n * 2, 1, SQLITE3_DETERMINISTIC);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
