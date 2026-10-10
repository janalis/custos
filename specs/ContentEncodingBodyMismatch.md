---
id: ContentEncodingBodyMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ContentEncodingBodyMismatch

## Summary

Encode the body using the declared content coding.

## Detection

- D1. Track an effective literal Content-Encoding gzip, deflate or identity and an emitted body proven produced directly by gzencode, gzcompress or gzdeflate. Report the emitted compression call when its format contradicts the declared coding: gzip requires gzip framing; HTTP deflate requires zlib framing. Identity contradicts any of these proven compressed outputs.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Unknown body provenance, buffering/transformation, compound coding lists and subsequent replacement headers invalidate proof. Do not claim all plain strings are uncompressed binary data.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ContentEncodingBodyMismatch` and `@noinspection ContentEncodingBodyMismatch` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Encode the body using the declared content coding.`

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
header("Content-Encoding: gzip"); echo <warning descr="Encode the body using the declared content coding.">gzcompress("payload")</warning>; exit;
```

Valid case:

```php
<?php
header("Content-Encoding: gzip"); echo gzencode("payload"); exit;
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
