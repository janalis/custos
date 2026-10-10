---
id: CsvEscapeDefaultDependency
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.4", max: "" }
---

# CsvEscapeDefaultDependency

## Summary

Pass the CSV escape character explicitly.

## Detection

- D1. For resolved fgetcsv, fputcsv, str_getcsv and SplFileObject fgetcsv/fputcsv/setCsvControl, report omission of the escape argument. For SplFileObject reading/writing, a proven prior setCsvControl with explicit escape satisfies the contract if the call does not override it. Bind named arguments to their signature positions.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Calls explicitly supplying escape, unknown receiver types and PHP targets below 8.4 are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore CsvEscapeDefaultDependency` and `@noinspection CsvEscapeDefaultDependency` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Pass the CSV escape character explicitly.`

## Fix

Add the explicit current default backslash escape, preserving behavior. On PHP 8.4 use a named escape argument when legal; otherwise fill only missing optional separator/enclosure/length defaults according to the exact API signature. Do not mix positional after named arguments or duplicate arguments. Empty escape is documented as an interoperability choice but is never silently introduced.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 8.4. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
<warning descr="Pass the CSV escape character explicitly.">fputcsv($fp, ["north", "south"])</warning>;
```

Fixed output:

```php
<?php
fputcsv($fp, ["north", "south"], escape: "\\");
```

Valid case:

```php
<?php
fputcsv($fp, ["north", "south"], ",", '"', "\\");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
