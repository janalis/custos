---
id: MultibytePositionUsedAsByteOffset
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# MultibytePositionUsedAsByteOffset

## Summary

Multibyte character positions are not byte offsets. Use matching multibyte operations when slicing at a multibyte search result.

## Detection

- D1. Resolved mb_strpos/mb_stripos on a known UTF-8 literal containing multibyte characters before a found needle yields index used as offset of substr on same source. Require literal needle and explicit UTF-8 encoding, no arithmetic conversion.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use multibyte offsets with multibyte slicing.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$s="été";$p=mb_strpos($s,"t",0,"UTF-8");echo <warning descr="Use multibyte offsets with multibyte slicing.">substr($s,$p,1)</warning>;
```

Valid case:

```php
<?php
$s="été";$p=mb_strpos($s,"t",0,"UTF-8");echo mb_substr($s,$p,1,"UTF-8");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
