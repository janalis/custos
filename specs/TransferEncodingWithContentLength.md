---
id: TransferEncodingWithContentLength
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# TransferEncodingWithContentLength

## Summary

Choose a single HTTP message framing mechanism.

## Detection

- D1. Track effective Transfer-Encoding and Content-Length headers set by builtin header in the same straight-line response segment. Report the second header establishing both. Header replacement and header_remove update state; either order is supported.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. A removal before the second header and unknown/branch-dependent header state are excluded. This rule checks HTTP/1.1 sender framing; manually setting Transfer-Encoding for other protocols is not legitimized.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore TransferEncodingWithContentLength` and `@noinspection TransferEncodingWithContentLength` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Choose a single HTTP message framing mechanism.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
header("Transfer-Encoding: chunked"); <warning descr="Choose a single HTTP message framing mechanism.">header("Content-Length: 12")</warning>;
```

Valid case:

```php
<?php
header("Content-Length: 12"); header_remove("Content-Length"); header("Transfer-Encoding: chunked");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
