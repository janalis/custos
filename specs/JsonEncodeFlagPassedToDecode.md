---
id: JsonEncodeFlagPassedToDecode
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.4", max: "" }
---

# JsonEncodeFlagPassedToDecode

## Summary

JSON encoding flags do not describe the intended decoding policy. Use the constants supported by json_decode.

## Detection

- D1. Resolved json_decode fourth flags argument syntactically includes an encoding-only resolved JSON constant: JSON_HEX_TAG, JSON_HEX_AMP, JSON_HEX_APOS, JSON_HEX_QUOT, JSON_FORCE_OBJECT, JSON_NUMERIC_CHECK, JSON_UNESCAPED_SLASHES, JSON_PRETTY_PRINT, JSON_UNESCAPED_UNICODE, JSON_PARTIAL_OUTPUT_ON_ERROR, or JSON_PRESERVE_ZERO_FRACTION. Shared decode-valid flags are excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use decoding flags in json_decode.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 5.4 or later.

## Examples

```php
<?php
<warning descr="Use decoding flags in json_decode.">json_decode($json,true,512,JSON_PRETTY_PRINT)</warning>;
```

Valid case:

```php
<?php
$a=json_decode($json,true,512,JSON_BIGINT_AS_STRING);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
