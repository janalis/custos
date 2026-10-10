---
id: CsvDelimiterInvalidByteLength
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CsvDelimiterInvalidByteLength

## Summary

Use a one-byte CSV separator.

## Detection

- D1. For resolved fgetcsv, fputcsv, str_getcsv (PHP 5.3+) and SplFileObject CSV methods/setCsvControl, report an explicitly provided separator whose constant string length is not exactly one byte. Use byte length, not Unicode scalar length.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Omitted default separator and unknown strings are excluded. A single-byte non-ASCII value is legal; UTF-8 multibyte separators are not.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore CsvDelimiterInvalidByteLength` and `@noinspection CsvDelimiterInvalidByteLength` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Use a one-byte CSV separator.`

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
<error descr="Use a one-byte CSV separator.">fgetcsv($fp, 0, "::")</error>;
```

Valid case:

```php
<?php
fgetcsv($fp, 0, ";");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
