---
id: UnserializeFalseAmbiguity
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnserializeFalseAmbiguity

## Summary

A false unserialize result can represent the valid serialized boolean false. Do not classify decoding failure solely from that value.

## Detection

- D1. Resolved unserialize result is compared strictly to false in an if condition whose branch unconditionally throws/returns failure. Include direct calls or unmodified local result. Allowed_classes does not eliminate serialized false.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Distinguish valid serialized false from decoding failure.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
if(<warning descr="Distinguish valid serialized false from decoding failure.">unserialize($bytes,["allowed_classes"=>false])</warning>===false){throw new RuntimeException();}
```

Valid case:

```php
<?php
$v=unserialize("b:0;",["allowed_classes"=>false]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
