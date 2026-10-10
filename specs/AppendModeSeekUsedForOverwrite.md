---
id: AppendModeSeekUsedForOverwrite
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# AppendModeSeekUsedForOverwrite

## Summary

Use an overwrite-capable stream mode.

## Detection

- D1. Track a successful fopen with literal append mode a/a+ (optional b/t/e suffixes), then successful seek to a proven position before a later fwrite/fputs on the same handle without reopening. Report the write because append writes ignore the seek position. Require seek position proven not EOF, such as rewind or fseek(handle,0,SEEK_SET), to avoid flagging harmless seek-to-end.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Unknown modes, seek-to-end, reads only and reopened handles are excluded. A false-open guard is required unless successful resource provenance is otherwise proven.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore AppendModeSeekUsedForOverwrite` and `@noinspection AppendModeSeekUsedForOverwrite` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Use an overwrite-capable stream mode.`

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
$fp = fopen($path, "a"); if ($fp === false) { return; } if (fseek($fp, 0) !== 0) { return; } <warning descr="Use an overwrite-capable stream mode.">fwrite($fp, "head")</warning>;
```

Valid case:

```php
<?php
$fp = fopen($path, "r+"); if ($fp === false) { return; } if (fseek($fp, 0) !== 0) { return; } fwrite($fp, "head");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
