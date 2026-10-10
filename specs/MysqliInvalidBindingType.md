---
id: MysqliInvalidBindingType
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# MysqliInvalidBindingType

## Summary

mysqli bindings accept only the supported integer, floating-point, string and blob type letters. Correct unsupported characters in the type string.

## Detection

- D1. Resolved mysqli_stmt::bind_param literal type string contains a character other than i,d,s,b. Highlight the type-string argument.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use only supported mysqli binding type letters.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
function bad(mysqli_stmt $s){$s->bind_param(<warning descr="Use only supported mysqli binding type letters.">"x"</warning>,$v);}
```

Valid case:

```php
<?php
$db=new mysqli();$s=$db->prepare("SELECT ?");$s->bind_param("s",$v);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
