---
id: ExitStatusReadWithoutNormalExitCheck
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ExitStatusReadWithoutNormalExitCheck

## Summary

Check normal termination before decoding the exit code.

## Detection

- D1. Track pcntl_wait/pcntl_waitpid output status into pcntl_wexitstatus. The read needs two facts, both established by conditions that dominate it: the wait succeeded, and pcntl_wifexited holds for the same status. Reassignment and intervening unknown calls receiving status invalidate provenance.
- D1a. Facts accumulate over every dominating condition; they do not have to share one condition. A condition known true contributes its `&&`/`and` operands; a condition known false contributes the negation of its `||`/`or` operands; `!` flips the polarity. Conditions known true: the condition of an enclosing `if`/`elseif` body, the condition of a ternary for its middle arm, the left operand of `&&`/`and` for its right operand. Conditions known false: every earlier `if`/`elseif` condition for a later `elseif`/`else` branch, the ternary condition for its last arm, the left operand of `||`/`or` for its right operand, and the condition of an earlier `if` in an enclosing statement list whose body always leaves (return, throw, exit, break, continue).
- D1b. Wait success is a comparison between the wait result and an integer, in either operand order, that implies a positive result: `> n` with n >= 0, `>= n` with n >= 1, or `!== -1` / `!= -1` when the wait call passes no flags (a blocking wait returns a child pid or -1). The negated forms (`<= 0`, `< 1`, `=== -1`) count when the condition is known false. Comparing the result with the awaited pid proves nothing: a pid argument of -1 equals the failure result.
- D1c. A pcntl_wifexited check counts only when it is written after the wait call; an earlier check describes an older status value.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape and unresolved calls that can mutate relevant state remove proof. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding, whether written as one condition, nested conditions, guard clauses that leave early, else branches, ternary arms or short-circuit operands. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. A status variable written between the wait and the read, by assignment, destructuring, increment, `foreach`, `unset`, an element write or a by-reference closure capture, no longer holds the wait status: no report.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Check normal termination before decoding the exit code.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
pcntl_waitpid($pid,$status); <warning descr="Check normal termination before decoding the exit code.">pcntl_wexitstatus($status)</warning>;
```

Valid case:

```php
<?php
$r=pcntl_waitpid($pid,$status); if ($r>0 && pcntl_wifexited($status)) { echo pcntl_wexitstatus($status); }
```

Guard clauses are equally valid:

```php
<?php
function exitCode(int $pid): ?int {
    $r = pcntl_waitpid($pid, $status);
    if ($r <= 0) { return null; }
    if (!pcntl_wifexited($status)) { return null; }
    return pcntl_wexitstatus($status);
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
