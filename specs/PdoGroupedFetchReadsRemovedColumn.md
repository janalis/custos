---
id: PdoGroupedFetchReadsRemovedColumn
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoGroupedFetchReadsRemovedColumn

## Summary

Identify read of first selected named column from a row in unchanged FETCH_GROUP combined with FETCH_ASSOC results. Require complete projection and mode knowledge; wildcard, duplicate names or ambiguous SQL exclude.

## Detection

- D1. Report read of first selected named column from a row in unchanged FETCH_GROUP combined with FETCH_ASSOC results. Require complete projection and mode knowledge; wildcard, duplicate names or ambiguous SQL exclude.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- Array and list destructuring destinations create slots rather than reading them. Destructuring keys and ordinary array values still require reads.

- Guarded probes with isset, empty or null coalescing, unsets and plain nested slot creation do not read a failed or missing value. Index expressions and the fallback operand of null coalescing still require normal reads.

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore PdoGroupedFetchReadsRemovedColumn` and `@noinspection PdoGroupedFetchReadsRemovedColumn` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Read the group column from the group key.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$pdo = new PDO($dsn); $rows = $pdo->query("SELECT 'x' AS category, 7 AS value")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['category'];
```

Valid case:

```php
<?php
$pdo = new PDO($dsn); $rows = $pdo->query("SELECT 'x' AS category, 7 AS value")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['value'];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
