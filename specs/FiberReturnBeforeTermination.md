---
id: FiberReturnBeforeTermination
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# FiberReturnBeforeTermination

## Summary

Identify getReturn on a locally allocated Fiber after start of a callback proven to suspend before termination, without a subsequent completion-proving resume.

## Detection

- D1. Report getReturn on a locally allocated Fiber after start of a callback proven to suspend before termination, when nothing can have completed the fiber before the read. The proof requires that the receiver variable is named only by its allocation, by start calls that dominate the read, and by the read itself in all code that can run before the read: everything textually before it in the same variable scope, plus the whole outermost enclosing loop when the read sits in one.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Any other mention of the receiver variable before the read abandons the proof, wherever it sits: a resume or throw inside a loop, branch, try block or arrow function, a state query such as isTerminated or isSuspended guarding the read, an alias assignment, an array or callable holding the fiber, or the fiber passed as an argument. Variable variables and calls to compact, extract or get_defined_vars in that code abandon it too. Nested functions, methods, property hooks and closures have their own variables and are not searched.
- E3. Both `@custos-ignore FiberReturnBeforeTermination` and `@noinspection FiberReturnBeforeTermination` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Read the fiber return after termination.`

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
$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();
```

Valid cases:

```php
<?php
$f = new Fiber(fn() => 7); $f->start(); $f->getReturn();
```

```php
<?php
$job = new Fiber(function () { Fiber::suspend(); return 'done'; });
$job->start();
while ($job->isSuspended()) {
    $job->resume();
}
echo $job->getReturn();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
