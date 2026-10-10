---
id: ForeachReferenceSurvivesLoop
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ForeachReferenceSurvivesLoop

## Summary

A by-reference foreach leaves its value variable attached to the last array element. Unset that reference before reusing the variable.

## Detection

- D1. A foreach binds a simple local variable by reference; after the loop, a reachable assignment to that same variable occurs before unset. Require the same lexical scope and no intervening control-flow ambiguity. Highlight the subsequent assignment.
- D1d. Exclude a statically proven empty iterable: no iteration then binds the reference. Also exclude a binding that is cleared with unset or rebound in the loop body; only a surviving element reference can justify the later-write finding. Unknown binding mutations discard survival proof.
- D1e. Known emptiness includes a directly written empty array and an immediately preceding ordinary non-reference assignment of an empty array to the same simple iterable local. Any unset of the binding or by-reference rebinding inside the loop body excludes the finding, even if the operation is conditional.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report a reference never reused, a by-value foreach, or a proven unset/rebinding before reuse.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Unset the foreach reference before reusing the variable.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$xs=[1,2]; foreach($xs as &$x) {} <warning descr="Unset the foreach reference before reusing the variable.">$x=9</warning>;
```

Valid case:

```php
<?php
$xs=[1,2]; foreach($xs as &$x) {} unset($x); $x=9;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
