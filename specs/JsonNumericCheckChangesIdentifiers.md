---
id: JsonNumericCheckChangesIdentifiers
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# JsonNumericCheckChangesIdentifiers

## Summary

Numeric JSON conversion can remove leading zeros or a leading plus sign from identifier strings. Keep those values as strings when their formatting matters. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved json_encode uses JSON_NUMERIC_CHECK and a literal structure contains a string with a leading plus or leading zero followed by digits; classify formatting loss, not field-name intent. Highlight flag.
- D1d. Evaluate the effective array after PHP key normalization and duplicate-key overwrites, recursively where nested values supply evidence. Only surviving entries establish row collisions, nested arrays, non-stringable objects, formatted numeric strings or nested list shapes. Earlier entries overwritten by a later occurrence of the same normalized key are not evidence; unknown keys or unsupported effective-array evaluation discard proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Preserve identifier strings during JSON encoding.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

Effective array-key behavior follows the [PHP array contract](https://www.php.net/manual/en/language.types.array.php).

## Examples

```php
<?php
json_encode(["phone"=>"+331234"],<warning descr="Preserve identifier strings during JSON encoding.">JSON_NUMERIC_CHECK</warning>);
```

Valid case:

```php
<?php
$s=json_encode(["code"=>"00123"]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
