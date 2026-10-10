---
id: CatchCannotHandleKnownThrowable
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CatchCannotHandleKnownThrowable

## Summary

Identify a catch clause when its directly enclosing try consists of one explicit throw of a resolved Error subtype and every catch type excludes that subtype; exclude finally and calls that can throw another type.

## Detection

- D1. Report a catch clause when its directly enclosing try consists of one explicit throw of a resolved Error subtype and every catch type excludes that subtype; exclude finally and calls that can throw another type.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore CatchCannotHandleKnownThrowable` and `@noinspection CatchCannotHandleKnownThrowable` suppress this native rule.

## Report

- Range: the catch type. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: `Catch a type that handles the thrown value.`

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
try { throw new Error('failed'); } catch (Exception $e) {}
```

Valid case:

```php
<?php
try { throw new Error('failed'); } catch (Throwable $e) {}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
