---
id: PdoRepeatedNamedMarkerWithoutEmulation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoRepeatedNamedMarkerWithoutEmulation

## Summary

Identify prepare on a proven PDO mysql native-preparation connection with EMULATE_PREPARES false and SQL containing repeated same named marker outside literals/comments. Exclude drivers permitting repetition, unresolved driver, emulation true and SQL parser uncertainty.

## Detection

- D1. Report prepare on a proven PDO mysql native-preparation connection with EMULATE_PREPARES false and SQL containing repeated same named marker outside literals/comments. Exclude drivers permitting repetition, unresolved driver, emulation true and SQL parser uncertainty.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore PdoRepeatedNamedMarkerWithoutEmulation` and `@noinspection PdoRepeatedNamedMarkerWithoutEmulation` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Use a distinct marker for each native parameter.`

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
$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :n + :n');
```

Valid case:

```php
<?php
$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :a + :b');
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
