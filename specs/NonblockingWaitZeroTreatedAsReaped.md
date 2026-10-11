---
id: NonblockingWaitZeroTreatedAsReaped
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# NonblockingWaitZeroTreatedAsReaped

## Summary

Require a positive PID from nonblocking wait.

## Detection

- D1. Report direct or unchanged-local pcntl_waitpid result compared !=/!== -1, or >= 0, with an explicit resolved WNOHANG option. Zero indicates no available status. Other three-way handling and compound flags remain outside this literal-flag subset.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

The literal comparison subset is exactly != or !== minus one, and >= zero. Comparisons != or !== zero already exclude a pending result and are not findings.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. No report when the same unchanged wait result is told apart from zero elsewhere, because the code then already handles the pending case. The separating test must start before the end of the statement that contains the comparison: an earlier statement (for example an early return on `=== 0`), another operand of the same condition, an enclosing condition, or a test inside the branch the comparison governs. A test placed after that statement cannot protect the guarded code and does not count.
- E4. Separating tests are: a comparison with zero using `===`, `!==`, `==`, `!=`, `> 0` or `<= 0`; `>= 1` or `< 1`; the mirrored forms with the literal on the left; and truthiness of the variable as an `if`/`elseif`/ternary condition, a `!` operand, an `&&`/`||` operand, or a `switch`/`match` subject. `>= 0`, `< 0`, comparisons with other values and tests of a different wait result do not separate zero from a reaped PID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Require a positive PID from nonblocking wait.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$r=pcntl_waitpid($pid,$status,WNOHANG); if (<warning descr="Require a positive PID from nonblocking wait.">$r!==-1</warning>) { echo pcntl_wexitstatus($status); }
```

Valid case:

```php
<?php
$r=pcntl_waitpid($pid,$status,WNOHANG); if ($r>0 && pcntl_wifexited($status)) { echo pcntl_wexitstatus($status); }
```

Zero handled by an earlier exit:

```php
<?php
function poll($pid) {
    $r = pcntl_waitpid($pid, $status, WNOHANG);
    if ($r === 0) { return 'running'; }
    if ($r !== -1) { return 'exited ' . pcntl_wexitstatus($status); }
    return 'error';
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
