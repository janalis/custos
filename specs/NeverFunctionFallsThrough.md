---
id: NeverFunctionFallsThrough
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# NeverFunctionFallsThrough

## Summary

Terminate every path in the never function.

## Detection

- D1. For a function, method or closure with resolved builtin never return type, report its return type when a reachable path reaches the closing brace. Use bounded control-flow proof, recognizing throw, exit/die, unconditional infinite loops and calls with a resolved never return type. An explicit return is already illegal and is not this fall-through finding.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Exclude unreachable closing braces, unknown control-flow summaries and budget exhaustion. A conditional throw proves termination only on its throwing branch.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore NeverFunctionFallsThrough` and `@noinspection NeverFunctionFallsThrough` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Terminate every path in the never function.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 8.1. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
function halt(bool $stop): <error descr="Terminate every path in the never function.">never</error> { if ($stop) { exit; } }
```

Valid case:

```php
<?php
function halt(): never { throw new RuntimeException(); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
