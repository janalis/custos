---
id: StreamOutputReturnUsedAsContent
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# StreamOutputReturnUsedAsContent

## Summary

Output the stream without printing its byte count.

## Detection

- D1. Report a direct readfile or fpassthru return used as the sole operand of echo or print. These functions already output stream bytes; their numeric successful return is output again.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

Remove only the output keyword and its separating whitespace. Preserve suffix comments between the stream call and the statement terminator.

- E1. Storing or testing the byte count, calls to file_get_contents, shadowed builtin names and multioperand echo are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore StreamOutputReturnUsedAsContent` and `@noinspection StreamOutputReturnUsedAsContent` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Output the stream without printing its byte count.`

## Fix

Replace a standalone echo/print statement whose only operand is the resolved call with that call as an expression statement. Retain the complete original call bytes, arguments and trailing semicolon. Offer no fix when wrapper removal would drop comments or when print is used within a larger expression.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
<warning descr="Output the stream without printing its byte count.">echo readfile("ledger.txt");</warning>
```

Fixed output:

```php
<?php
readfile("ledger.txt");
```

Valid case:

```php
<?php
readfile("ledger.txt");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
