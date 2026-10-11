---
id: SysvMessageExceedsReceiveCapacity
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SysvMessageExceedsReceiveCapacity

## Summary

Identify msg_receive whose positive maxsize is smaller than byte length of proven same-queue same-type successfully sent nonserialized constant string, without MSG_NOERROR. Serialized or competing messages exclude.

## Detection

- D1. Report msg_receive whose positive maxsize is smaller than byte length of proven same-queue same-type successfully sent nonserialized constant string, without MSG_NOERROR. Serialized or competing messages exclude.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. A variable read for the capacity, the receive flags, either message type, the sent payload or the serialize flag is proven only while its assigned literal still holds. An increment or decrement of that variable, or a closure capturing it by reference, placed before the read or anywhere in a loop that repeats the read, makes the value unknown: `$size = 9; $size++;` gives a capacity of 10, not 9.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E3. No report when D4 makes any required value unknown. A step placed after the receive and outside every loop around it, a by-value closure capture and plain reads leave the value proven.
- E2. Both `@custos-ignore SysvMessageExceedsReceiveCapacity` and `@noinspection SysvMessageExceedsReceiveCapacity` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Allow enough receive capacity for the message.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$queue = msg_get_queue(4312); if (msg_send($queue, 8, 'pineapple', false)) { msg_receive($queue, 8, $type, 4, $message, false); }
```

Valid case:

```php
<?php
$queue = msg_get_queue(4312); if (msg_send($queue, 8, 'pineapple', false)) { msg_receive($queue, 8, $type, 32, $message, false); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
