---
id: CatchVariableOverwritesLocal
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CatchVariableOverwritesLocal

## Summary

A catch variable shares its surrounding local scope. Use a different name when the previous local value is needed after exception handling. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. A simple local has a definite non-exception assignment before try; a catch binds that same name and the variable is read after try without reassignment. Require lexical same-scope statements and a reachable throwing operation.
- D1d. The catch path must reach the later read while retaining the caught exception binding. Exclude catches that terminate by return/throw/exit, reassign or unset the local, and finally blocks that restore or replace its value. Unknown catch/finally operations discard the retained-binding proof.
- D1e. Exclude terminating catches and any assignment, increment/decrement or unset of the catch variable in the catch body or finally. These operations invalidate the retained exception-binding proof; do not treat the subsequent read as an overwrite hazard.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Use a different catch variable to preserve the local value.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$result="ready"; try{work();}catch(Throwable <warning descr="Use a different catch variable to preserve the local value.">$result</warning>){} echo $result;
```

Valid case:

```php
<?php
$result="ready"; try{work();}catch(Throwable $error){} echo $result;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
