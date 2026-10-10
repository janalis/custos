---
id: FiberResumeBeforeStart
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# FiberResumeBeforeStart

## Summary

A newly created fiber must be started before it can be resumed. Resume only a suspended fiber.

## Detection

- D1. A local assigned new resolved Fiber is used as resume receiver before start or any unknown escape; require straight-line statements. Highlight resume call.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Start the fiber before resuming it.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 8.1 or later.

## Examples

```php
<?php
$f=new Fiber(fn()=>1); <warning descr="Start the fiber before resuming it.">$f->resume()</warning>;
```

Valid case:

```php
<?php
$f=new Fiber(function(){Fiber::suspend();}); $f->start(); $f->resume();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
