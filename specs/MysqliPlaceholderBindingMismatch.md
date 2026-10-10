---
id: MysqliPlaceholderBindingMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# MysqliPlaceholderBindingMismatch

## Summary

Every SQL parameter placeholder needs a corresponding bound value. Match the binding argument count to the prepared statement.

## Detection

- D1. Resolved mysqli::prepare literal SQL produces statement whose bind_param binding argument count differs from SQL question-mark parameter count. Ignore question marks inside SQL quoted strings, backticks and comments. Exclude SQL parse ambiguity and unpacked bindings.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Bind one value for each SQL placeholder.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$db=new mysqli();$s=$db->prepare("SELECT ? + ?");<warning descr="Bind one value for each SQL placeholder.">$s->bind_param("i",$id)</warning>;
```

Valid case:

```php
<?php
$db=new mysqli();$s=$db->prepare("SELECT ? + ?");$s->bind_param("ii",$a,$b);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
