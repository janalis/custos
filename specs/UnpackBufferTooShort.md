---
id: UnpackBufferTooShort
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnpackBufferTooShort

## Summary

Provide enough bytes for the unpack format.

## Detection

- D1. For constant unpack format, byte string and optional nonnegative offset, interpret widths and cursor moves and prove a required read exceeds available bytes. Supported widths: c/C one byte, n/v two bytes, N/V four bytes, J/P eight bytes, g/G four bytes, e/E eight bytes; a/A/Z byte counts and h/H nibble counts. Cursor directives x/X/@ update byte position. Unpack slash-separated field names do not consume bytes. A star read consumes only remaining bytes and does not itself establish underflow. Include network integer widths and fixed string counts; unsupported directives or platform-dependent widths invalidate proof. Exclude J/P before PHP 5.6, g/G/e/E before PHP 7.2, Z before PHP 5.5, and explicit offsets before PHP 7.1.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore UnpackBufferTooShort and @noinspection UnpackBufferTooShort suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Provide enough bytes for the unpack format.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
$n = <warning descr="Provide enough bytes for the unpack format.">unpack('Nnumber', "\x01\x02")</warning>;
```

Valid case:

```php
<?php
$n = unpack('Nnumber', "\x00\x00\x01\x02");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
