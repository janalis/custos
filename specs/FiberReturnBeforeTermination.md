---
id: FiberReturnBeforeTermination
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# FiberReturnBeforeTermination

## Summary

Identify getReturn on a locally allocated Fiber after start of a callback proven to suspend before termination, without a subsequent completion-proving resume.

## Detection

- D1. Report getReturn on a locally allocated Fiber after start of a callback proven to suspend before termination, without a subsequent completion-proving resume.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore FiberReturnBeforeTermination` and `@noinspection FiberReturnBeforeTermination` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Read the fiber return after termination.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 8.1 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();
```

Valid case:

```php
<?php
$f = new Fiber(fn() => 7); $f->start(); $f->getReturn();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
