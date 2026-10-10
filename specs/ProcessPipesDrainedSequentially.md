---
id: ProcessPipesDrainedSequentially
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ProcessPipesDrainedSequentially

## Summary

Reading a child's stdout and stderr sequentially can deadlock when the other pipe fills. Drain both output pipes concurrently. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved proc_open descriptors declare stdout and stderr as separate output pipes; parent performs blocking stream_get_contents on pipes[1] then pipes[2]. No nonblocking mode/select/concurrent draining. Require owned pipe array and no reassignment.
- D1c. The initial supported proof is three adjacent statements: proc_open, blocking stdout read, blocking stderr read, using the same simple pipes-array name and literal descriptor indices.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Drain process output pipes concurrently.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$p=proc_open("worker",[1=>["pipe","w"],2=>["pipe","w"]],$pipes);$stdout=stream_get_contents($pipes[1]);$stderr=<warning descr="Drain process output pipes concurrently.">stream_get_contents($pipes[2])</warning>;
```

Valid case:

```php
<?php
$p=proc_open($cmd,[1=>["pipe","w"],2=>["file","/tmp/custos-stderr","a"]],$pipes);$a=stream_get_contents($pipes[1]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
