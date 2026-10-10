---
id: NonRepeatableAttributeRepeated
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# NonRepeatableAttributeRepeated

## Summary

Remove the repeated attribute.

## Detection

- D1. Within one attribute-bearing declaration, resolve attribute class identities; report the second and subsequent use of a class whose Attribute flags prove IS_REPEATABLE absent. This predicts failure on ReflectionAttribute::newInstance.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Aliases of the same class count as identical. Different attribute classes, separate declarations and unknown repeatability are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore NonRepeatableAttributeRepeated` and `@noinspection NonRepeatableAttributeRepeated` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Remove the repeated attribute.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 8.0. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
#[Attribute] class Label {} #[Label, <error descr="Remove the repeated attribute.">Label</error>] class Parcel {}
```

Valid case:

```php
<?php
#[Attribute(Attribute::TARGET_CLASS | Attribute::IS_REPEATABLE)] class Label {} #[Label, Label] class Parcel {}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
