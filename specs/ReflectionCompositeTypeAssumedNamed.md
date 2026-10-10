---
id: ReflectionCompositeTypeAssumedNamed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# ReflectionCompositeTypeAssumedNamed

## Summary

Inspect the composite reflection type members.

## Detection

- D1. Resolve getType on a ReflectionParameter, ReflectionProperty or getReturnType on a ReflectionFunctionAbstract for a known union type (PHP 8.0+) or intersection type (PHP 8.1+). Report getName invoked on that result, because only ReflectionNamedType provides it. Follow local assignments and indexed parameter selection.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Guarded instanceof ReflectionNamedType branches, calls on members returned by getTypes and unknown reflected declarations are excluded. Nullable single named types remain ReflectionNamedType.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ReflectionCompositeTypeAssumedNamed` and `@noinspection ReflectionCompositeTypeAssumedNamed` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Inspect the composite reflection type members.`

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
function consume(int|string $x) {} $r = new ReflectionFunction("consume"); $t = $r->getParameters()[0]->getType(); <error descr="Inspect the composite reflection type members.">$t->getName()</error>;
```

Valid case:

```php
<?php
function consume(int|string $x) {} $r = new ReflectionFunction("consume"); $t = $r->getParameters()[0]->getType(); foreach ($t->getTypes() as $part) { $part->getName(); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
