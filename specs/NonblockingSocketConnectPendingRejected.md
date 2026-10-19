---
id: NonblockingSocketConnectPendingRejected
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# NonblockingSocketConnectPendingRejected

## Summary

Identify a `socket_connect` failure branch that throws or returns false although the same socket was switched to nonblocking mode and the branch never looks at the failure reason. A nonblocking connect normally reports failure while the connection is still being established, so the branch rejects a connection that may yet succeed.

## Detection

- D1. Report a `socket_connect` call whose result is tested for failure (`!call`, `call == false`, optionally through `@`, parentheses or an assignment) by an `if` whose body throws or returns false, when all of the following hold:
  - the last `socket_set_nonblock` / `socket_set_block` call on the same socket before the connect is `socket_set_nonblock`;
  - that call is proven to have run: its success guards the connect: a test of that call's result, written on the call itself, assigned in the condition or kept in a boolean, holds on every path to the connect (`if (socket_set_nonblock($s)) { ... }`, a preceding terminating `if (!socket_set_nonblock($s))`, `socket_set_nonblock($s) or exit;`), or it belongs to a statement that always executes before the connect. A switch inside an earlier conditional branch or behind a short-circuit operator proves nothing;
  - the failure branch does not inspect the failure reason (D4).
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. The failure branch inspects the reason when, outside nested closures and functions, it calls `socket_last_error`, `socket_get_option` or `socket_getopt`, or calls anything that is not a builtin function (user or unresolved function, dynamic callee, instance or static method) whose receiver, callee or arguments mention a variable used to name the socket. Builtin functions other than the three listed never count, even when they receive the socket.

- D6. A failure test must select a falsy result: count negation and loose equality with literal false, on either side, by parity. Parentheses, error suppression and the right-hand value of a plain non-reference assignment preserve the test. Compound assignments and assignment targets do not.
- D7. Require every supported path through the failure body to throw or return literal false. A conditional rejection alone is insufficient. Structured blocks and conditional branches are bounded; suspension, jumps, unsupported control flow and budget exhaustion establish no rejection proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. An unchecked statement-level `socket_set_nonblock` can fail in theory; it is still accepted as proof because a failing switch leaves an unusable socket for which the connect fails as well.
- E3. A failure branch that hands the socket to a helper (D4) is not reported: the helper may accept the pending state.
- E4. Both `@custos-ignore NonblockingSocketConnectPendingRejected` and `@noinspection NonblockingSocketConnectPendingRejected` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Handle a pending nonblocking connection.`

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
if (socket_set_nonblock($s)) { if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); } }
```

Finding, unchecked switch:

```php
<?php
function open_async($s, $host) {
    socket_set_nonblock($s);
    if (!@socket_connect($s, $host, 443)) {
        return false;
    }
    return true;
}
```

Valid case:

```php
<?php
socket_set_block($s); if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); }
```

Valid case, the helper receives the socket:

```php
<?php
function open_checked($s, $host) {
    socket_set_nonblock($s);
    if (!socket_connect($s, $host, 443)) {
        if (still_connecting($s)) {
            return true;
        }
        return false;
    }
    return true;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
