---
id: TruncationAssumedToRewind
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# TruncationAssumedToRewind

## Summary

Rewind after truncating the stream.

## Detection

- D1. Track a successful writable stream, a proven positive current position, ftruncate(handle,0), and subsequent fwrite/fputs before repositioning. Report the write because ftruncate does not reset position. Positive position may come from a successful fseek or known-length prior write.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Rewind/fseek to zero, unknown position, failed truncation branches and intentional sparse-file construction explicitly outside this rule through suppression are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore TruncationAssumedToRewind` and `@noinspection TruncationAssumedToRewind` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Rewind after truncating the stream.`

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
$fp = fopen($path, "w+"); if ($fp === false) { return; } if (fseek($fp, 10) !== 0) { return; } if (ftruncate($fp, 0)) { <warning descr="Rewind after truncating the stream.">fwrite($fp, "new")</warning>; }
```

Valid case:

```php
<?php
$fp = fopen($path, "w+"); if ($fp === false) { return; } if (fseek($fp, 10) !== 0) { return; } if (ftruncate($fp, 0)) { rewind($fp); fwrite($fp, "new"); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
