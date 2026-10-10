---
id: NonVoidFunctionFallsThrough
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# NonVoidFunctionFallsThrough

## Summary

Identify the return type of an explicitly nonvoid nongenerator function when structural control flow proves a reachable normal exit without return or throw. Exclude mixed, void, never and untyped returns; a bare return is covered separately. Restrict proof to finite straight-line blocks and if/else branches; loops, switch, try, goto, parser recovery and all calls on a candidate exit path exclude because they may terminate, throw or exit.

## Detection

- D1. Report the return type of an explicitly nonvoid nongenerator function when structural control flow proves a reachable normal exit without return or throw. Exclude mixed, void, never and untyped returns; a bare return is covered separately. Restrict proof to finite straight-line blocks and if/else branches; loops, switch, try, goto, parser recovery and all calls on a candidate exit path exclude because they may terminate, throw or exit.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore NonVoidFunctionFallsThrough` and `@noinspection NonVoidFunctionFallsThrough` suppress this native rule.

## Report

- Range: the declared return type. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Return a value on every reachable path.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 7.0 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
function countItems(bool $ready): int { if ($ready) return 7; }
```

Valid case:

```php
<?php
function countItems(bool $ready): int { if ($ready) return 7; return 0; }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
