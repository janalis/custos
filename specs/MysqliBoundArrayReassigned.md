---
id: MysqliBoundArrayReassigned
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# MysqliBoundArrayReassigned

## Summary

mysqli parameter bindings retain references to the originally bound array elements. Update those elements rather than replacing the whole array.

## Detection

- D1. Resolved mysqli_stmt::bind_param binds a literal-key element of a local literal array. The whole array is reassigned before execute without rebinding; highlight whole-array assignment. Element updates are valid.
- D1d. The later execute call must use the existing reference bindings. An execute call supplying replacement parameters is excluded, including a known parameter array; unknown execution arguments cannot prove the stale binding is used.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Update the bound array element without replacing its array.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

See the [PHP mysqli execution parameter contract](https://www.php.net/manual/en/mysqli-stmt.execute.php) for the relevant builtin behavior.

## Examples

```php
<?php
$db=new mysqli();$s=$db->prepare("SELECT ?");$row=["id"=>1];$s->bind_param("i",$row["id"]);<warning descr="Update the bound array element without replacing its array.">$row=["id"=>2]</warning>;$s->execute();
```

Valid case:

```php
<?php
$db=new mysqli();$s=$db->prepare("SELECT ?");$a=["id"=>1];$s->bind_param("i",$a["id"]);$a["id"]=2;$s->execute();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
