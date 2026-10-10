---
id: CtypeIntegerInterpretedAsCharacterCode
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CtypeIntegerInterpretedAsCharacterCode

## Summary

An integer ctype argument can be interpreted as a character code. Pass a string when testing the characters of its decimal representation. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved ctype_digit receives integer literal in range 0–255; report disabled-by-default interpretation/deprecation hazard. String literals and explicit string casts excluded; initial detector restricts to ctype_digit rather than all ctype functions.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Pass a string to avoid character-code interpretation.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$ok=ctype_digit(<warning descr="Pass a string to avoid character-code interpretation.">48</warning>);
```

Valid case:

```php
<?php
$ok=ctype_digit("48");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
