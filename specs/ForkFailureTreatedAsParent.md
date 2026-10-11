---
id: ForkFailureTreatedAsParent
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ForkFailureTreatedAsParent

## Summary

Handle fork failure separately.

## Detection

- D1. Report direct/unchanged-local pcntl_fork result tested by an if truthiness or negation condition; -1 is truthy. Numeric three-way dispatch remains outside this check.
- D1a. No report when failure is already excluded on every path to the test. A failure test is a comparison between the fork result and a known integer that is true whenever the result is -1: `=== -1`, `== -1`, `< n` with n at least 0, `<= n` with n at least -1, and the mirrored spellings with the integer on the left (`-1 === $pid`, `0 > $pid`). The fork result is the local while it still resolves to the reported call, or the plain assignment that stores that call (`($pid = pcntl_fork()) === -1`). Failure is excluded when a failure test is the condition of an earlier if in the same statement list as the reported if, or as any statement enclosing it inside the same function, and the body of that earlier if always leaves (return, throw, exit, break, continue); or when the reported if lies in an else or elseif arm of a chain whose earlier condition is a failure test. Unrelated statements between guard and test do not matter.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. A guard that does not leave, one comparing another call or an earlier value of the local, one placed after the test or inside a closure, and comparisons that -1 does not satisfy (`> -1`, `< -1`, `=== 0`, `!== -1`, unknown bounds) do not. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Handle fork failure separately.`

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
$pid=pcntl_fork(); if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo "parent"; } else { echo "child"; }
```

Valid case:

```php
<?php
$pid=pcntl_fork(); if ($pid===-1) { throw new RuntimeException(); } if ($pid>0) { echo "parent"; } else { echo "child"; }
```

Failure handled in the same chain, or by a mirrored guard followed by other work:

```php
<?php
$pid = pcntl_fork();
if ($pid === -1) { exit(1); } else if ($pid) { echo "parent"; } else { echo "child"; }

$worker = pcntl_fork();
if (0 > $worker) { throw new RuntimeException(); }
$started = time();
if ($worker) { echo "parent"; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
