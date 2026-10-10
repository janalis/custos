---
id: ProcessClosedBeforeOwnedPipes
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ProcessClosedBeforeOwnedPipes

## Summary

Waiting for a child process while its owned pipes remain open can deadlock. Close or drain the pipes before waiting for process exit.

## Detection

- D1. Resolved proc_open declares at least one parent-owned pipe and assigns process local; proc_close occurs before all declared pipes are definitely closed/drained. Highlight proc_close; unknown descriptor shape excluded.
- D1c. The initial supported proof is an immediately preceding local assignment from proc_open, with explicit owned pipe descriptors and no intervening statements.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Close owned process pipes before waiting for process exit.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$p=proc_open("worker",[1=>["pipe","w"]],$pipes);<warning descr="Close owned process pipes before waiting for process exit.">proc_close($p)</warning>;
```

Valid case:

```php
<?php
$p=proc_open($cmd,[0=>["pipe","r"]],$pipes);fclose($pipes[0]);proc_close($p);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
