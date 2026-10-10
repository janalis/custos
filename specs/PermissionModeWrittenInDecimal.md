---
id: PermissionModeWrittenInDecimal
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PermissionModeWrittenInDecimal

## Summary

A decimal permission literal can represent a different bitmask from the intended octal notation. Write Unix permission masks with an explicit octal prefix. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved chmod or mkdir mode argument is a decimal integer token of exactly three/four octal digits with nonzero first digit and numeric value exceeding 0777, such as 755. Report disabled-by-default notation hazard; explicit octal, hex, symbolic/bitwise mode excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Write the permission mask in octal notation.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
<warning descr="Write the permission mask in octal notation.">chmod($path,755)</warning>;
<warning descr="Write the permission mask in octal notation.">mkdir($path,1777)</warning>;
```

Valid case:

```php
<?php
chmod($p,0755);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
