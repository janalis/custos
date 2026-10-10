---
id: RecursiveTraversalOmitsDirectories
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# RecursiveTraversalOmitsDirectories

## Summary

Include parent nodes when traversing directories. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy requires directory nodes as well as leaves. Report construction of RecursiveIteratorIterator wrapping a proven RecursiveDirectoryIterator, directly or through RecursiveCallbackFilterIterator, when mode is omitted or explicitly LEAVES_ONLY. Highlight constructor call.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. SELF_FIRST and CHILD_FIRST satisfy the policy. Unknown iterator source and arbitrary recursive iterators are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore RecursiveTraversalOmitsDirectories` and `@noinspection RecursiveTraversalOmitsDirectories` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Include parent nodes when traversing directories.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Enabling the rule declares the policy stated in Detection; the rule is disabled by default.

## PHP versions

Minimum PHP 5.3. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding with the rule explicitly enabled:

```php
<?php
$dir = new RecursiveDirectoryIterator("."); $it = <warning descr="Include parent nodes when traversing directories.">new RecursiveIteratorIterator($dir)</warning>;
```

Valid case:

```php
<?php
$dir = new RecursiveDirectoryIterator("."); $it = new RecursiveIteratorIterator($dir, RecursiveIteratorIterator::SELF_FIRST);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
