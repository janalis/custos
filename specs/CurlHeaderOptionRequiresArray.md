---
id: CurlHeaderOptionRequiresArray
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlHeaderOptionRequiresArray

## Summary

Identify curl_setopt HTTPHEADER option with a value proven scalar rather than array; curl_setopt_array requires exact option-map extraction.

## Detection

- D1. Report curl_setopt HTTPHEADER option with a value proven scalar rather than array; curl_setopt_array requires exact option-map extraction.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore CurlHeaderOptionRequiresArray` and `@noinspection CurlHeaderOptionRequiresArray` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Supply headers as an array.`

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
curl_setopt($ch, CURLOPT_HTTPHEADER, 'Accept: text/plain');
```

Valid case:

```php
<?php
curl_setopt($ch, CURLOPT_HTTPHEADER, ['Accept: text/plain']);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
