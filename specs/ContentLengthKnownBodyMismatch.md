---
id: ContentLengthKnownBodyMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ContentLengthKnownBodyMismatch

## Summary

Match Content-Length to the emitted bytes.

## Detection

- D1. Track effective literal Content-Length set by builtin header. Report that header when a complete straight-line local response path emits a statically known sequence of echo/print literal bytes and terminates, with total byte length unequal to the declared decimal length. Replacements override prior headers. Include final top-level path or a function ending in exit; do not treat returning from a helper as response completion.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Unknown output, output buffering/handlers, zlib compression, includes, unknown calls, HEAD handling, 1xx/204/304 responses and transfer encoding invalidate body-length proof. A configured SAPI transform cannot be assumed absent if evidence enables it.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ContentLengthKnownBodyMismatch` and `@noinspection ContentLengthKnownBodyMismatch` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Match Content-Length to the emitted bytes.`

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
<warning descr="Match Content-Length to the emitted bytes.">header("Content-Length: 2")</warning>; echo "hello"; exit;
```

Valid case:

```php
<?php
header("Content-Length: 5"); echo "hello"; exit;
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
