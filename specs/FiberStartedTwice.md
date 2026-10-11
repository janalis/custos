---
id: FiberStartedTwice
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# FiberStartedTwice

## Summary

Identify a second start on the same locally allocated Fiber after a successful first start on an unconditional straight-line path; no reset or reassignment may intervene.

## Detection

- D1. Report a second start on the same locally allocated Fiber after a successful first start on an unconditional straight-line path; no reset or reassignment may intervene.
  Only a real invocation counts as a start, for the earlier call and for the reported one: the first-class callable form `$fiber->start(...)` merely builds a closure.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. No report when the receiver variable is bound again between the two starts: any assignment whose target mentions it (plain, `[$f] = ...` or `list($f) = ...` destructuring), a `foreach` that uses it as key or value, or a `catch` clause that names it. The second call then acts on another value.
- E3. No report when either call is a first-class callable creation (`$f->start(...)`).
- E4. Both `@custos-ignore FiberStartedTwice` and `@noinspection FiberStartedTwice` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Start each fiber only once.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 8.1 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->start();
```

Valid case:

```php
<?php
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->resume();
```

Rebound receiver and callable creation, also valid:

```php
<?php
$job = new Fiber(fn() => 1);
$job->start();
[$job] = [new Fiber(fn() => 2)];
$job->start();

$task = new Fiber(fn() => 3);
$launch = $task->start(...);
$task->start();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
