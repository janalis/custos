---
id: SessionDestroyLeavesLocalAuthentication
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SessionDestroyLeavesLocalAuthentication

## Summary

Destroying stored session data does not clear the current session array. Clear local authentication state before reading it again.

## Detection

- D1. Resolved session_destroy precedes a read of $_SESSION in a condition in same scope, with no intervening $_SESSION=[]/unset or successful new session_start. Highlight post-destroy session read.
- D1c. The initial supported proof is an immediately preceding session_destroy and direct session array access as the if condition; guarded or nested conditions are excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Clear local session data after destroying the session.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
session_destroy(); if(<warning descr="Clear local session data after destroying the session.">$_SESSION["admin"]</warning>){echo "admin";}
```

Valid case:

```php
<?php
session_destroy();$_SESSION=[];if(isset($_SESSION["admin"])){echo "admin";}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
