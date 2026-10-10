---
id: CsvBlankRecordShapeMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CsvBlankRecordShapeMismatch

## Summary

Recognize a blank CSV record as a null field.

## Detection

- D1. Track a fgetcsv or SplFileObject::fgetcsv result; report strict equality/inequality against an empty array when a blank record is proven from a local in-memory stream containing a single newline (LF or CRLF), at position zero. For an unknown stream, do not infer blank-record intent from the comparison alone.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Empty arrays produced by application filtering, unknown stream data and checks against [null] are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore CsvBlankRecordShapeMismatch` and `@noinspection CsvBlankRecordShapeMismatch` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Recognize a blank CSV record as a null field.`

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
$fp = fopen("php://memory", "r+"); fwrite($fp, "
"); rewind($fp); if (<warning descr="Recognize a blank CSV record as a null field.">fgetcsv($fp) === []</warning>) {}
```

Valid case:

```php
<?php
$fp = fopen("php://memory", "r+"); fwrite($fp, "
"); rewind($fp); if (fgetcsv($fp) === [null]) {}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
