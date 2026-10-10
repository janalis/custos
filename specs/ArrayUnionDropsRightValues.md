---
id: ArrayUnionDropsRightValues
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayUnionDropsRightValues

## Summary

Array union keeps left-hand values when keys collide. Use merging when both lists must contribute their values. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Both operands of array union are nonempty literal arrays with implicit zero-based keys, and at least one right key collides. The result is subsequently iterated as a value list without key access; opt-in because first-value precedence can be intentional.
- D1a. The initial supported downstream proof is an immediately following by-value foreach without a key binding.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Use array merging to preserve the right-hand list values.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$all=<warning descr="Use array merging to preserve the right-hand list values.">[10,20]+[30,40]</warning>; foreach($all as $v){echo $v;}
```

Valid case:

```php
<?php
$a=array_merge([10,20],[30,40]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
