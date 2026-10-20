---
id: JsonInvalidDepth
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# JsonInvalidDepth

## Summary

Identify json_decode or json_encode with known integer depth at most zero. Additionally report json_decode depth above 2147483647; do not assume the decoder upper bound applies to json_encode.

## Detection

- D1. Report json_decode or json_encode with known integer depth at most zero. Additionally report json_decode depth above 2147483647; do not assume the decoder upper bound applies to json_encode.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. A variable depth counts as known only when no write that the value resolution cannot follow may run between its assignment and the call: an increment or decrement of that variable, a by-reference closure capture of it, a `foreach` or `catch` binding to it, or an `extract()` call in the same scope. Inside a loop such a write anywhere in the scope disqualifies the depth, because it precedes the call on the next iteration.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E3. A depth variable initialized to a non-positive value and then stepped (`$depth = 0; $depth++;`, or a counter advanced inside the loop that makes the call) is not reported: the value reaching the call is no longer the initial one.
- E2. Both `@custos-ignore JsonInvalidDepth` and `@noinspection JsonInvalidDepth` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Supply a supported JSON depth.`

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
json_decode('{}', true, 0);
```

Valid case:

```php
<?php
json_decode('{}', true, 512);

$depth = 0;
$depth++;
json_decode('[]', true, $depth);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
