---
id: ContentDispositionFilenameNeedsQuoting
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ContentDispositionFilenameNeedsQuoting

## Summary

Quote the Content-Disposition filename.

## Detection

- D1. Parse a literal Content-Disposition header and identify an unquoted filename parameter containing characters outside HTTP token grammar, including whitespace. Report the full header call. Recognize quoted-string escaping and filename-star as separate syntax.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Already quoted filenames, token-safe filenames, RFC extended filename parameters, unknown values and ambiguous multiple filename parameters are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ContentDispositionFilenameNeedsQuoting` and `@noinspection ContentDispositionFilenameNeedsQuoting` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Quote the Content-Disposition filename.`

## Fix

For one unambiguous literal header containing one unquoted filename parameter, quote its value and escape double quotes/backslashes as HTTP quoted-string syntax, then encode the replacement PHP literal without altering its evaluated bytes elsewhere. Suppress for CR/LF, ambiguous parameter boundaries, interpolation or comment loss.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
<warning descr="Quote the Content-Disposition filename.">header("Content-Disposition: attachment; filename=monthly notes.pdf")</warning>;
```

Fixed output:

```php
<?php
header("Content-Disposition: attachment; filename=\"monthly notes.pdf\"");
```

Valid case:

```php
<?php
header('Content-Disposition: attachment; filename="monthly notes.pdf"');
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
