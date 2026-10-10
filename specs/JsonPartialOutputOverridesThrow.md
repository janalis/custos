---
id: JsonPartialOutputOverridesThrow
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.3", max: "" }
---

# JsonPartialOutputOverridesThrow

## Summary

Partial JSON output takes precedence over throwing on encoding errors. Choose the error policy the caller actually expects.

## Detection

- D1. Resolved json_encode flag expression definitely contains both JSON_PARTIAL_OUTPUT_ON_ERROR and JSON_THROW_ON_ERROR. Highlight the flag expression.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Choose either partial JSON output or throwing on errors.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 7.3 or later.

## Examples

```php
<?php
json_encode($data,<warning descr="Choose either partial JSON output or throwing on errors.">JSON_PARTIAL_OUTPUT_ON_ERROR|JSON_THROW_ON_ERROR</warning>);
```

Valid case:

```php
<?php
$s=json_encode($data,JSON_THROW_ON_ERROR);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
