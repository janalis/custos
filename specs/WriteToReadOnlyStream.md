---
id: WriteToReadOnlyStream
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# WriteToReadOnlyStream

## Summary

Open the stream with write access.

## Detection

- D1. Track successful fopen with literal r mode and no plus. Report fwrite, fputs, fputcsv or ftruncate receiving that handle before close/reopen.
- D4. `php://memory` and `php://temp` (including `php://temp/maxmemory:N`) follow their own rule: the stream is read-only only when the literal mode contains none of `w`, `a` and `+` at any position. So `r`, `rb`, `x` and `c` are reported there, while `rw`, `rbw`, `ra` and `wr` are writable.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. r+ and other writable modes, unknown resources and wrappers with unknown semantics are excluded.
- E3. A memory or temporary stream whose mode holds `w`, `a` or `+` anywhere is writable even when the mode starts with `r` (`fopen('php://memory', 'rw')`). The same `rw` mode on a regular file stays read-only and is reported.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore WriteToReadOnlyStream` and `@noinspection WriteToReadOnlyStream` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Open the stream with write access.`

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
$fp = fopen($path, "r"); if ($fp === false) { return; } <error descr="Open the stream with write access.">fwrite($fp, "change")</error>;
```

Valid case:

```php
<?php
$fp = fopen($path, "r+"); if ($fp === false) { return; } fwrite($fp, "change");
```

Valid case, in-memory stream opened for reading and writing:

```php
<?php
$buffer = fopen('php://memory', 'rw'); if ($buffer === false) { return; } fwrite($buffer, "draft");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
