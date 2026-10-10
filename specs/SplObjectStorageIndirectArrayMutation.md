---
id: SplObjectStorageIndirectArrayMutation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# SplObjectStorageIndirectArrayMutation

## Summary

Assign the modified storage value back.

## Detection

- D1. Resolve SplObjectStorage offset whose most recent stored value is a proven array. Report nested offset/property-like array writes, increment/decrement or unset applied indirectly to its retrieved array. Highlight the complete mutation expression.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Stored objects are mutable handles and are excluded. Assigning a retrieved modified array back is valid. Unknown stored type or custom offsetGet behavior is excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore SplObjectStorageIndirectArrayMutation` and `@noinspection SplObjectStorageIndirectArrayMutation` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Assign the modified storage value back.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 5.3. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
$s = new SplObjectStorage(); $o = new stdClass(); $s[$o] = ["count" => 1]; <warning descr="Assign the modified storage value back.">$s[$o]["count"]++</warning>;
```

Valid case:

```php
<?php
$s = new SplObjectStorage(); $o = new stdClass(); $s[$o] = ["count" => 1]; $v = $s[$o]; $v["count"]++; $s[$o] = $v;
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
