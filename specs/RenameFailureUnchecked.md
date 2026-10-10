---
id: RenameFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RenameFailureUnchecked

## Summary

Renaming a file can fail. Check the result before returning success from the publication operation. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved rename result is discarded immediately before containing function returns literal true. Restrict to proven success return rather than arbitrary success-named consumers; guarded/returned rename results excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Check rename success before returning success.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
function bad($a,$b){<warning descr="Check rename success before returning success.">rename($a,$b)</warning>;return true;}
```

Valid case:

```php
<?php
function publish($a,$b){return rename($a,$b);}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
