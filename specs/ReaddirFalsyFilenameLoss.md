---
id: ReaddirFalsyFilenameLoss
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ReaddirFalsyFilenameLoss

## Summary

Compare the directory entry with false.

## Detection

- D1. Report resolved readdir directly tested for truthiness in loop/if conditions, including assignment-in-condition and loose comparison to false. A legal filename consisting of digit zero is otherwise treated as end-of-directory.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Strict comparisons against false, merely assigned results and unrelated functions are excluded.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ReaddirFalsyFilenameLoss` and `@noinspection ReaddirFalsyFilenameLoss` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Compare the directory entry with false.`

## Fix

Replace direct positive truthiness or assignment-in-condition with a parenthesized !== false comparison; replace negated direct tests with === false. Preserve assignment and call evaluation exactly once. Fix loose comparisons only when their boolean direction is unambiguous, retaining comments.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
while (<warning descr="Compare the directory entry with false.">$entry = readdir($dir)</warning>) { echo $entry; }
```

Fixed output:

```php
<?php
while (($entry = readdir($dir)) !== false) { echo $entry; }
```

Valid case:

```php
<?php
while (($entry = readdir($dir)) !== false) { echo $entry; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
