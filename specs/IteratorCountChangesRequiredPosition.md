---
id: IteratorCountChangesRequiredPosition
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# IteratorCountChangesRequiredPosition

## Summary

Identify a position-dependent current or key read on a locally allocated finite ArrayIterator after iterator_count consumes it, without rewind or seek. Unknown iterator implementations are excluded.

## Detection

- D1. Report a position-dependent current or key read on a locally allocated finite ArrayIterator after iterator_count consumes it, without rewind or seek. Unknown iterator implementations are excluded.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore IteratorCountChangesRequiredPosition` and `@noinspection IteratorCountChangesRequiredPosition` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: `Restore the iterator position after counting.`

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
$it = new ArrayIterator([7, 11]); iterator_count($it); echo $it->current();
```

Valid case:

```php
<?php
$it = new ArrayIterator([7, 11]); iterator_count($it); $it->rewind(); echo $it->current();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
