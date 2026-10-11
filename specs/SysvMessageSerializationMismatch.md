---
id: SysvMessageSerializationMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SysvMessageSerializationMismatch

## Summary

Identify msg_receive when local successful msg_send on same queue and exact positive matching type uses a different serialize flag. Require known flags, no unknown queue operation and no competing-sender evidence; unknown message selection excludes.

## Detection

- D1. Report msg_receive when local successful msg_send on same queue and exact positive matching type uses a different serialize flag. Require known flags, no unknown queue operation and no competing-sender evidence; unknown message selection excludes.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- Receiving serialized bytes with decoding disabled is valid. Raw payloads with a possible PHP serialization prefix, or unknown payloads, cannot prove a decoding mismatch.

- A flag, message type or payload read from a variable is not a known value once that variable may have been stepped with `++` or `--`, or captured by reference by a closure, before the call reads it. A step later in a loop that also encloses the read counts too, because the next iteration observes it. A step after the read outside any shared loop, or a by-value capture, leaves the value known.

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore SysvMessageSerializationMismatch` and `@noinspection SysvMessageSerializationMismatch` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Use matching serialization flags for the message.`

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
$queue = msg_get_queue(4311); if (msg_send($queue, 7, 'payload', false)) { msg_receive($queue, 7, $type, 64, $message, true); }
```

Valid case:

```php
<?php
$queue = msg_get_queue(4311); if (msg_send($queue, 7, 'payload', false)) { msg_receive($queue, 7, $type, 64, $message, false); }
```

No finding, because the send flag is no longer the assigned literal:

```php
<?php
$queue = msg_get_queue(4312); $encode = 0; $encode++; if (msg_send($queue, 7, 'payload', $encode)) { msg_receive($queue, 7, $type, 64, $message, true); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
