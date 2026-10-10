---
id: PdoArrayBoundToSinglePlaceholder
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoArrayBoundToSinglePlaceholder

## Summary

A single PDO placeholder cannot expand an array into multiple SQL values. Create a separate placeholder for each list element.

## Detection

- D1. Resolved PDOStatement::execute literal arguments array contains a nested array as a parameter value. Resolve statement provenance from PDO::prepare and exclude subclasses overriding execute.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Expand list values into separate SQL placeholders.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$pdo=new PDO("sqlite::memory:");$s=$pdo->prepare("SELECT ?");<warning descr="Expand list values into separate SQL placeholders.">$s->execute([[1,2]])</warning>;
```

Valid case:

```php
<?php
$p=new PDO($dsn);$s=$p->prepare("SELECT * FROM t WHERE id IN (?,?)");$s->execute([1,2]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
