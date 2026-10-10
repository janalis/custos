---
id: AttributeTargetMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# AttributeTargetMismatch

## Summary

Use the attribute on a permitted target.

## Detection

- D1. Resolve the attribute class and its builtin Attribute annotation; constant-fold the target bitmask (default TARGET_ALL). Report an application on a class, function, method, property, class constant or parameter whose corresponding bit is absent. This predicts failure of ReflectionAttribute::newInstance, not immediate parse failure. PHP 8.5 property hooks are method targets.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

A constructor-promoted parameter is also a property declaration. Accept attributes permitting either parameter or property targets at that combined declaration.

- E1. Skip unresolved attribute declarations and unknown bitmasks. Attribute::IS_REPEATABLE is not a target bit.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore AttributeTargetMismatch` and `@noinspection AttributeTargetMismatch` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Use the attribute on a permitted target.`

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
#[Attribute(Attribute::TARGET_METHOD)] class Marker {} <error descr="Use the attribute on a permitted target.">#[Marker]</error> class Product {}
```

Valid case:

```php
<?php
#[Attribute(Attribute::TARGET_CLASS)] class Marker {} #[Marker] class Product {}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
