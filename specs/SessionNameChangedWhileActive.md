---
id: SessionNameChangedWhileActive
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SessionNameChangedWhileActive

## Summary

A session name must be configured before the session becomes active. Set the name before starting the session.

## Detection

- D1. Resolved session_start in straight-line scope precedes session_name setter before any session_write_close/session_abort/session_destroy. Require definite active state; unknown calls invalidate proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E1a. A session_start call with read_and_close true does not establish an active session. Unknown option values discard active-state proof.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Set the session name before starting the session.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
session_start(); <warning descr="Set the session name before starting the session.">session_name("CUSTOM")</warning>;
```

Valid case:

```php
<?php
session_name("CUSTOM");session_start();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
