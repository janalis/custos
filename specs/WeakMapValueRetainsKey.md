---
id: WeakMapValueRetainsKey
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# WeakMapValueRetainsKey

## Summary

Store a value that does not retain the WeakMap key. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy requires automatic weak-key eviction. Report assigning a WeakMap value that is the identical key object or a literal array directly containing that same object; bounded local aliases of the key count. Highlight the assignment.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Do not speculate about object graphs or closure captures beyond proven direct retention. Different objects, scalars and WeakReference wrappers are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore WeakMapValueRetainsKey` and `@noinspection WeakMapValueRetainsKey` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Store a value that does not retain the WeakMap key.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Enabling the rule declares the policy stated in Detection; the rule is disabled by default.

## PHP versions

Minimum PHP 8.0. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding with the rule explicitly enabled:

```php
<?php
$m = new WeakMap(); $o = new stdClass(); <warning descr="Store a value that does not retain the WeakMap key.">$m[$o] = $o</warning>;
```

Valid case:

```php
<?php
$m = new WeakMap(); $o = new stdClass(); $m[$o] = "cached";
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
