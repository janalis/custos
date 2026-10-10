---
id: ContentRangeLengthMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ContentRangeLengthMismatch

## Summary

Match the body to the inclusive byte range.

## Detection

- D1. Track effective Content-Range bytes START-END/TOTAL (decimal numeric bounds, valid single range) and a proven status 206. Report the range header when complete emitted constant bytes differ from END - START + 1. Support repeated literal concatenations and constant-folded str_repeat. Use the same complete response-path proof and invalidations as ContentLengthKnownBodyMismatch.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- Redirect Location and CGI Status headers can change response status implicitly; exclude their body proof.

- E1. Multipart ranges, unknown output, malformed ranges, unsatisfied bytes */TOTAL, buffering/transformation, HEAD and incomplete response paths are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ContentRangeLengthMismatch` and `@noinspection ContentRangeLengthMismatch` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Match the body to the inclusive byte range.`

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
http_response_code(206); <warning descr="Match the body to the inclusive byte range.">header("Content-Range: bytes 0-4/20")</warning>; echo "four"; exit;
```

Valid case:

```php
<?php
http_response_code(206); header("Content-Range: bytes 0-3/20"); echo "four"; exit;
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
