---
id: ReadFromWriteOnlyStream
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ReadFromWriteOnlyStream

## Summary

Open the stream with read access.

## Detection

- D1. Track successful fopen with literal w/a/x/c mode and no plus. Report fread, fgets, fgetc, fgetcsv, stream_get_contents or fpassthru receiving that handle before close/reopen.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Plus modes, r modes, wrappers with unknown semantics and unknown resources are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ReadFromWriteOnlyStream` and `@noinspection ReadFromWriteOnlyStream` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Open the stream with read access.`

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
$fp = fopen($path, "w"); if ($fp === false) { return; } <error descr="Open the stream with read access.">fread($fp, 8)</error>;
```

Valid case:

```php
<?php
$fp = fopen($path, "w+"); if ($fp === false) { return; } fread($fp, 8);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
