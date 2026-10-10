---
id: OverrideAttributeWithoutTarget
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.3", max: "" }
---

# OverrideAttributeWithoutTarget

## Summary

Apply Override to an inherited member.

## Detection

- D1. Resolve the builtin Override attribute on a method. Report only when the complete parent and interface closure proves there is no overridable method of that name. From PHP 8.5 also check properties against inherited properties; methods and properties are separate lookup domains. A private parent method does not satisfy a method override; abstract interface requirements do.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Unknown ancestors or inherited traits prevent a definite missing-target proof. Exclude shadowed custom Override attributes and inaccessible private declarations.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore OverrideAttributeWithoutTarget` and `@noinspection OverrideAttributeWithoutTarget` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Apply Override to an inherited member.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 8.3. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
class Worker { <error descr="Apply Override to an inherited member.">#[\Override]</error> public function start() {} }
```

Valid case:

```php
<?php
interface Starter { public function start(); } class Worker implements Starter { #[\Override] public function start() {} }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
