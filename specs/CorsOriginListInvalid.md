---
id: CorsOriginListInvalid
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CorsOriginListInvalid

## Summary

Send one allowed origin in the CORS header.

## Detection

- D1. Report builtin header setting Access-Control-Allow-Origin to a literal containing multiple comma-separated or whitespace-separated serialized origins. Accept exactly one valid serialized origin, null or star. Case-insensitive header name; whitespace around value is ignored.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Unknown values and ordinary non-CORS headers are excluded. Do not reinterpret a comma inside an invalid URL as a supported origin list.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore CorsOriginListInvalid` and `@noinspection CorsOriginListInvalid` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Send one allowed origin in the CORS header.`

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
<warning descr="Send one allowed origin in the CORS header.">header("Access-Control-Allow-Origin: https://one.test, https://two.test")</warning>;
```

Valid case:

```php
<?php
header("Access-Control-Allow-Origin: https://one.test");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
