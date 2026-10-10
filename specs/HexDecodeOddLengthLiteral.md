---
id: HexDecodeOddLengthLiteral
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.4", max: "" }
---

# HexDecodeOddLengthLiteral

## Summary

Hexadecimal decoding requires complete pairs of digits. Supply an even number of hexadecimal characters.

## Detection

- D1. Resolved hex2bin receives a constant string of odd byte length containing only ASCII hexadecimal digits. Highlight input argument.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Provide an even number of hexadecimal digits.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 5.4 or later.

## Examples

```php
<?php
hex2bin(<warning descr="Provide an even number of hexadecimal digits.">"abc"</warning>);
```

Valid case:

```php
<?php
$b=hex2bin("abcd");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
