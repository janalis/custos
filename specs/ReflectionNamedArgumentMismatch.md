---
id: ReflectionNamedArgumentMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# ReflectionNamedArgumentMismatch

## Summary

Use a declared reflection parameter name.

## Detection

- D1. Resolve ReflectionFunction or ReflectionMethod constructed from a known declaration. For invokeArgs, inspect literal string keys against resolved parameter names; report an unknown key only when the target is nonvariadic. Numeric keys remain positional. Account for inherited methods and named constructor arguments.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Unknown reflected targets, unknown argument arrays and variadic targets accepting extra named arguments are excluded. Do not extend the named-key behavior to PHP 7.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ReflectionNamedArgumentMismatch` and `@noinspection ReflectionNamedArgumentMismatch` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Use a declared reflection parameter name.`

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
function welcome($name) {} $r = new ReflectionFunction("welcome"); <error descr="Use a declared reflection parameter name.">$r->invokeArgs(["nickname" => "Kai"])</error>;
```

Valid case:

```php
<?php
function welcome($name) {} $r = new ReflectionFunction("welcome"); $r->invokeArgs(["name" => "Kai"]);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
