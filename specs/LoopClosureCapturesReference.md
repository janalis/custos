---
id: LoopClosureCapturesReference
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# LoopClosureCapturesReference

## Summary

Deferred closures that capture a loop counter by reference observe its later value. Capture by value when each callback needs its own iteration value. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. A closure appended to a local array inside a for loop explicitly captures the loop counter by reference. The array is consumed after the loop and the closure body reads that counter in its own variable scope. Reads inside nested closures, arrow functions, functions or other variable scopes do not establish this proof. A plain assignment target is a write, not a read; compound assignments and increments also read their target. Require no intervening overwrite or escape of the array.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E1a. A nested callback parameter with the same name does not read the captured outer counter. A closure that only assigns the counter without reading it is excluded.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Capture the iteration counter by value for deferred callbacks.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=<warning descr="Capture the iteration counter by value for deferred callbacks.">function() use (&$i){return $i;}</warning>; } foreach($jobs as $job){echo $job();}
```

Valid case:

```php
<?php
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use ($i){return $i;}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){return function($i){return $i;};}; } foreach($jobs as $job){echo $job();}
$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){$i=1;return 2;}; } foreach($jobs as $job){echo $job();}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
