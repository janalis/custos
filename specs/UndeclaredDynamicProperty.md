---
id: UndeclaredDynamicProperty
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.2", max: "" }
---

# UndeclaredDynamicProperty

## Summary

Detect assigning a missing instance property on a resolved PHP 8.2+ class with a complete known hierarchy.

## Detection

- D1. Report assigning a missing instance property on a resolved PHP 8.2+ class with a complete known hierarchy.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. stdClass descendants, AllowDynamicProperties classes, classes implementing __set, unknown hierarchies and existing properties are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Declare this property before assigning it.`

## Fix

No automatic fix: declaring a property changes visibility, serialization and initialization contracts.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.2. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
class Person {} $p=new Person(); <warning descr="Declare this property before assigning it.">$p->label</warning>='Mia';
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: require a final runtime class or an exact fresh allocation. A nominal non-final base parameter may receive a subclass declaring the property. An immediately preceding resolved property_exists rejection with a single throw or return excludes a simple subsequent write; unknown calls or additional branch work cannot establish this proof.
