---
id: AsyncSignalSetterReturnMisread
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.1", max: "" }
---

# AsyncSignalSetterReturnMisread

## Summary

Query the signal state after changing it.

## Detection

This is an opt-in policy. Truthiness, previous-state checks, raw wait-status comparisons and dispatch acknowledgements can be intentional; enabling this inspection requests explicit handling of the contract described below.

- D1. Report a resolved pcntl_async_signals setter with a proven nonnull scalar argument used directly in a truthiness or negation condition. Getter calls, null and unknown enable values are excluded. A standalone negated if condition may be repaired by moving the setter before the if and testing the getter.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

Getter-mode null is resolved through unchanged locals and parentheses. A fix that inserts a statement requires a top-level or braced statement list; withhold it for an unbraced enclosing statement.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Query the signal state after changing it.`

## Fix

For standalone if conditions containing exactly !setter or setter === false/true, move the original setter to a preceding expression statement and test a zero-argument getter with the original polarity. No fix for elseif, nested short-circuit conditions, assignments, comments across moved ranges or expressions evaluated conditionally.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 7.1 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
if (<warning descr="Query the signal state after changing it.">!pcntl_async_signals(true)</warning>) { throw new RuntimeException(); }
```

Fixed result:

```php
<?php
pcntl_async_signals(true);
if (!pcntl_async_signals()) { throw new RuntimeException(); }
```

Valid case:

```php
<?php
pcntl_async_signals(true); if (!pcntl_async_signals()) { throw new RuntimeException(); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
